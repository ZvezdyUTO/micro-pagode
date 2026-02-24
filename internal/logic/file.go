package logic

import (
	"context"
	"io"
	"mp/internal/errno"
	"mp/internal/search"
	"mp/pkg/logx"
	"os"
	"path/filepath"
	"strings"

	"mp/internal/domain"
	"mp/internal/repository"
)

type File interface {
	List(ctx context.Context, req *domain.FilePathReq) (resp *domain.FileListResp, err error)
	CreateDir(ctx context.Context, req *domain.FilePathReq) error
	CreateFile(ctx context.Context, req *domain.FilePathReq) error
	Delete(ctx context.Context, req *domain.FileDeleteReq) error
	Upload(ctx context.Context, req *domain.UploadFileReq) (*domain.File, error)
	Download(ctx context.Context, req *domain.FilePathReq) (io.ReadSeekCloser, *domain.File, error)
	Search(ctx context.Context, req *domain.FileSearchReq) (*domain.FileSearchResp, error)
}

type file struct {
	repo     repository.FileRepository
	basePath string
	search   search.FileSearchService
}

func NewFile(repo repository.FileRepository, basePath string, search search.FileSearchService) File {
	return &file{
		repo:     repo,
		basePath: basePath,
		search:   search,
	}
}

func (l *file) resolvePath(userPath string) (absPath string, relPath string, err error) {
	relPath = filepath.Clean(userPath)
	if relPath == "." {
		relPath = ""
	}

	base := filepath.Clean(l.basePath)
	absPath = filepath.Join(base, relPath)

	if !strings.HasPrefix(absPath, base) {
		return "", "", errno.ErrInvalidPath
	}

	return absPath, relPath, nil
}

func (l *file) List(ctx context.Context, req *domain.FilePathReq) (*domain.FileListResp, error) {
	absPath, relPath, err := l.resolvePath(req.Path)
	if err != nil {
		return nil, err
	}

	// 调 repository 的功能
	list, err := l.repo.List(ctx, absPath)
	if err != nil {
		logx.Errors(ctx, "file", "get_file_list_faild", logx.Fields{
			"stage": "list",
			"path":  absPath,
			"error": err.Error(),
		})
		return nil, err
	}

	// 补业务 Path ，那个函数返回的是文件，但我们需要补上用户之前提供的路径
	for _, f := range list {
		f.Path = filepath.Join(relPath, f.Filename)
	}

	return &domain.FileListResp{
		List: list,
	}, nil
}

func (l *file) CreateDir(ctx context.Context, req *domain.FilePathReq) error {
	absPath, _, err := l.resolvePath(req.Path)
	if err != nil {
		return err
	}

	// 判断是否已存在
	if _, err := os.Stat(absPath); err == nil {
		return errno.ErrAlreadyExist
	}

	l.search.MarkDirty()
	return l.repo.CreateDir(ctx, absPath)
}

func (l *file) CreateFile(ctx context.Context, req *domain.FilePathReq) error {
	absPath, _, err := l.resolvePath(req.Path)
	if err != nil {
		return err
	}

	parent := filepath.Dir(absPath)
	if _, err := os.Stat(parent); err != nil {
		return errno.ErrParentNotExist
	}

	l.search.MarkDirty()
	return l.repo.CreateFile(ctx, absPath)
}

func (l *file) Delete(ctx context.Context, req *domain.FileDeleteReq) error {
	absPath, _, err := l.resolvePath(req.Path)
	if err != nil {
		return err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return errno.ErrNotExist
	}

	if info.IsDir() {
		if !req.Recursive {
			entries, _ := os.ReadDir(absPath)
			if len(entries) > 0 {
				return errno.ErrDirNotEmpty
			}
		}
		l.search.MarkDirty()
		return l.repo.RemoveDir(ctx, absPath)
	}

	l.search.MarkDirty()
	return l.repo.RemoveFile(ctx, absPath)
}

func (l *file) Upload(
	ctx context.Context,
	req *domain.UploadFileReq,
) (*domain.File, error) {

	absPath, relPath, err := l.resolvePath(req.Path)
	if err != nil {
		return nil, err
	}

	if req.File == nil {
		return nil, errno.ErrInvalidPath
	}
	defer func() { _ = req.File.Close() }()

	// 父目录存在性校验
	parent := filepath.Dir(absPath)
	exists, err := l.repo.Exists(ctx, parent)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errno.ErrParentNotExist
	}

	// 不允许覆盖
	exists, err = l.repo.Exists(ctx, absPath)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errno.ErrAlreadyExist
	}

	// 保证路径为受控路径后，执行操作，repo：保存文件 + 计算 hash
	file, actualHash, err := l.repo.SaveFile(ctx, absPath, req.File)
	if err != nil {
		return nil, err
	}

	// hash 校验
	if req.ExpectedSHA256 != "" &&
		!strings.EqualFold(req.ExpectedSHA256, actualHash) {

		// hash 不一致，说明上传不可信
		_ = l.repo.RemoveFile(ctx, absPath)
		return nil, errno.ErrHashMismatch
	}

	// 赋业务字段
	file.Path = relPath
	file.SHA256 = actualHash

	l.search.MarkDirty()
	return file, nil
}

func (l *file) Download(ctx context.Context, req *domain.FilePathReq) (io.ReadSeekCloser, *domain.File, error) {
	absPath, relPath, err := l.resolvePath(req.Path)
	if err != nil {
		return nil, nil, err
	}

	// 下载的本质不是返回数据而是暴露读取能力
	reader, file, err := l.repo.OpenFile(ctx, absPath)
	if err != nil {
		return nil, nil, err
	}

	file.Path = relPath
	return reader, file, nil
}

func (l *file) Search(
	ctx context.Context,
	req *domain.FileSearchReq,
) (*domain.FileSearchResp, error) {

	// 1. 参数整理
	keyword := strings.TrimSpace(req.Keyword)
	if keyword == "" {
		return &domain.FileSearchResp{
			Total: 0,
			List:  nil,
		}, nil
	}

	opt := search.SearchOption{
		Prefix: req.Prefix,
	}

	// 2. 调用搜索服务
	matches, err := l.search.Search(ctx, keyword, opt)
	if err != nil {
		return nil, err
	}

	// 3. 结果裁剪 & 转换
	limit := req.Limit
	if limit <= 0 || limit > len(matches) {
		limit = len(matches)
	}

	list := make([]domain.FileSummary, 0, limit)
	for i := 0; i < limit; i++ {
		m := matches[i]
		list = append(list, domain.FileSummary{
			Path:  m.Path,
			Name:  m.Name,
			IsDir: m.IsDir,
		})
	}

	return &domain.FileSearchResp{
		Total: len(matches),
		List:  list,
	}, nil
}

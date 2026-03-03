/*
此处负责文件 IO、hash 计算、系统调用
纯粹负责功能，不关心业务逻辑
*/

package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mp/internal/errno"
	"os"
	"path/filepath"
	"sync"

	"mp/internal/domain"
)

type LocalFileRepository struct {
	root string
	mu   sync.Mutex
}

func NewLocalFileRepository() *LocalFileRepository {
	return &LocalFileRepository{}
}

func (r *LocalFileRepository) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (r *LocalFileRepository) List(ctx context.Context, path string) ([]*domain.File, error) {
	entries, err := os.ReadDir(path) // 读取目录下的所有文件，此处为操作系统原始信息
	if err != nil {
		return nil, err
	}

	files := make([]*domain.File, 0, len(entries))
	for _, entry := range entries { // 将文件翻译为上层能接受的形式
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, &domain.File{
			Filename: entry.Name(),
			Path:     filepath.Join(path, entry.Name()),
			Mtime:    info.ModTime().Unix(),
			Size:     info.Size(),
			IsDir:    entry.IsDir(),
		})
	}

	return files, nil
}

func (r *LocalFileRepository) CreateDir(ctx context.Context, path string) error {
	return os.MkdirAll(path, 0755) // MkdirAll: 如果中间目录不存在也一起创建
}

func (r *LocalFileRepository) CreateFile(ctx context.Context, path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0644) // 创建文件，但如果文件已存在则返回报错
	if err != nil {
		return err
	}
	return file.Close()
}

func (r *LocalFileRepository) RemoveFile(ctx context.Context, path string) error {
	return os.Remove(path)
}

func (r *LocalFileRepository) RemoveDir(ctx context.Context, path string) error {
	return os.RemoveAll(path) // 递归地删除整个目录，判断是否为空目录由 logic 决定
}

func (r *LocalFileRepository) SaveFile(
	ctx context.Context,
	path string,
	src io.Reader,
) (*domain.File, string, error) {

	// 1. 创建 SHA-256 计算器
	hasher := sha256.New()

	// 2. TeeReader：一边读数据，一边算 hash
	reader := io.TeeReader(src, hasher)

	// 3. 临时文件路径
	tmpPath := path + ".uploading"

	// 4. 打开一个“写文件出口”
	dst, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = dst.Close()
	}()

	// 5. 流式写入（真正的上传发生在这里）
	// 因为文件大小不受控制，因此必须使用流式 IO
	if _, err := io.Copy(dst, reader); err != nil {
		_ = os.Remove(tmpPath)
		return nil, "", err
	}

	// 6. 原子替换为正式文件，要么完全成功，要么什么也没有发生
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return nil, "", err
	}

	// 7. 获取文件信息
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", err
	}

	// 8. 生成 hash 字符串
	hash := hex.EncodeToString(hasher.Sum(nil))

	return &domain.File{
		Filename: filepath.Base(path),
		Size:     info.Size(),
		Mtime:    info.ModTime().Unix(),
		IsDir:    false,
	}, hash, nil
}

func (r *LocalFileRepository) OpenFile(
	ctx context.Context,
	path string,
) (io.ReadSeekCloser, *domain.File, error) {

	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	if info.IsDir() {
		return nil, nil, errno.ErrInvalidPath
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	return f, &domain.File{
		Filename: filepath.Base(path),
		Size:     info.Size(),
		Mtime:    info.ModTime().Unix(),
		IsDir:    false,
	}, nil
}

func (r *LocalFileRepository) WalkFiles(
	ctx context.Context,
	root string,
	fn func(path string, name string, isDir bool) error, // 使用回调函数
) error {

	var walk func(path string) error // 定义一个函数，这个函数可以调用它自己
	walk = func(path string) error {
		entries, err := os.ReadDir(path) // 读取当前目录，entires 就是当前目录下所有文件和子文件夹
		if err != nil {
			return err
		}

		for _, e := range entries { // 遍历所有文件，执行函数操作，边遍历边执行，此处只负责遍历
			full := filepath.Join(path, e.Name())

			// 执行外部传进来的函数
			if err := fn(full, e.Name(), e.IsDir()); err != nil {
				return err
			}

			// 如果是目录，则递归执行
			if e.IsDir() {
				if err := walk(full); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return walk(root)
}

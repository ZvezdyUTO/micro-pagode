package search

import (
	"context"
	"mp/internal/repository"
	"strings"
)

// BuildSnapshot 创建文件快照，返回文件元信息切片
func BuildSnapshot(
	ctx context.Context,
	repo repository.FileRepository,
	root string,
) ([]FileMeta, error) {

	var result []FileMeta

	err := repo.WalkFiles(ctx, root, func( // 传给 walk 函数，负责边遍历边执行这些操作
		path string,
		name string,
		isDir bool,
	) error { // 此处是需要执行的函数，包括将遍历目录时遇到的所有文件元信息写入快照
		result = append(result, FileMeta{
			Path:       path,
			Name:       name,
			SearchName: strings.ToLower(name),
			IsDir:      isDir,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

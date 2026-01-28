package search

import (
	"context"
	"mp/internal/repository"
	"strings"
)

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
	) error {
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

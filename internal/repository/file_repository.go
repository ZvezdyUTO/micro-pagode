package repository

import (
	"context"
	"io"
	"mp/internal/domain"
)

// 把“是否允许”和“如何获取”分开，这样存储方式变化不会影响业务规则

type FileRepository interface {
	Exists(ctx context.Context, path string) (bool, error)
	List(ctx context.Context, path string) ([]*domain.File, error)

	CreateDir(ctx context.Context, path string) error
	CreateFile(ctx context.Context, path string) error

	RemoveFile(ctx context.Context, path string) error
	RemoveDir(ctx context.Context, path string) error

	SaveFile(ctx context.Context, path string, src io.Reader) (*domain.File, string, error)
	OpenFile(ctx context.Context, path string) (io.ReadSeekCloser, *domain.File, error)

	// WalkFiles 用于递归地遍历所有目录，并且执行外部传进来的函数
	WalkFiles(
		ctx context.Context,
		root string,
		fn func(path string, name string, isDir bool) error,
	) error
}

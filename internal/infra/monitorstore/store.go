package monitorstore

import "context"

type Configurable interface {
	GetConfig(ctx context.Context) (maxRecord int, path string)
	UpdateConfig(ctx context.Context, maxRecord int) error
	Flush(ctx context.Context) error
}

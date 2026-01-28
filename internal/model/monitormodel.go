package model

import (
	"context"
)

type MonitorModel interface {
	InsertOne(ctx context.Context, data map[MonitorType]*MonitorKV) error
	State(ctx context.Context, t MonitorType, startTime, endTime int64) ([]*MonitorKV, error)
}

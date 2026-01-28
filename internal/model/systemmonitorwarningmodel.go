package model

import (
	"context"

	"gorm.io/gorm"
)

type SystemMonitorWarningModel interface {
	Inserts(ctx context.Context, data []*SystemMonitorWarning) error
	Insert(ctx context.Context, data *SystemMonitorWarning) error
}

type defaultSystemMonitorWarning struct {
	db *gorm.DB
}

func NewSystemMonitorWarningModel(db *gorm.DB) SystemMonitorWarningModel {
	return &defaultSystemMonitorWarning{db: db}
}

func (m *defaultSystemMonitorWarning) Inserts(ctx context.Context, data []*SystemMonitorWarning) error {
	if len(data) == 0 {
		return nil
	}
	return m.db.WithContext(ctx).Create(data).Error
}

func (m *defaultSystemMonitorWarning) Insert(ctx context.Context, data *SystemMonitorWarning) error {
	return m.db.WithContext(ctx).Create(data).Error
}

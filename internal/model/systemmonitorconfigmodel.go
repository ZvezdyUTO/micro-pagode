package model

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type SystemMonitorConfigModel interface {
	Insert(ctx context.Context, data *SystemMonitorConfig) error
	Get(ctx context.Context) (*SystemMonitorConfig, error)
	Update(ctx context.Context, data *SystemMonitorConfig) error
}

type defaultSystemMonitorConfig struct {
	db *gorm.DB
}

func NewSystemMonitorConfigModel(db *gorm.DB) SystemMonitorConfigModel {
	return &defaultSystemMonitorConfig{db: db}
}

func (m *defaultSystemMonitorConfig) model() *gorm.DB {
	return m.db.Model(&SystemMonitorConfig{})
}

func (m *defaultSystemMonitorConfig) Insert(ctx context.Context, data *SystemMonitorConfig) error {
	return m.db.WithContext(ctx).Create(data).Error
}

func (m *defaultSystemMonitorConfig) Get(ctx context.Context) (*SystemMonitorConfig, error) {
	var res SystemMonitorConfig
	err := m.db.WithContext(ctx).First(&res).Error
	switch err {
	case nil:
		return &res, nil
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultSystemMonitorConfig) Update(ctx context.Context, data *SystemMonitorConfig) error {
	if data.Id == 0 {
		return errors.New("update config with zero id")
	}

	return m.db.WithContext(ctx).
		Model(&SystemMonitorConfig{}).
		Where("id = ?", data.Id).
		Updates(map[string]interface{}{
			"is_start":       data.IsStart,
			"cpu_limit":      data.CpuLimit,
			"disk_limit":     data.DiskLimit,
			"men_limit":      data.MenLimit,
			"net_send_limit": data.NetSendLimit,
			"net_recv_limit": data.NetRecvLimit,
			"notify_type":    data.NotifyType,
			"email":          data.Email,
		}).Error
}

// internal/model/systemmonitorconfigtype.go
package model

import "time"

type NotifyType int

const (
	NotifyTypeEmail NotifyType = iota + 1
)

type SystemMonitorConfig struct {
	Id           int64      `gorm:"column:id;primaryKey"`
	IsStart      bool       `gorm:"column:is_start"`
	CpuLimit     float64    `gorm:"column:cpu_limit"`
	DiskLimit    float64    `gorm:"column:disk_limit"`
	MenLimit     float64    `gorm:"column:men_limit"`
	NetSendLimit float64    `gorm:"column:net_send_limit"`
	NetRecvLimit float64    `gorm:"column:net_recv_limit"`
	NotifyType   NotifyType `gorm:"column:notify_type"`
	Email        string     `gorm:"column:email"`
	CreateAt     time.Time  `gorm:"column:create_at;autoCreateTime"`
	UpdateAt     time.Time  `gorm:"column:update_at;autoUpdateTime"`
}

func (SystemMonitorConfig) TableName() string {
	return "system_monitor_config"
}

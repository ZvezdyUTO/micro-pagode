// internal/model/systemmonitorwarningtype.go
package model

import "time"

type SystemMonitorWarning struct {
	Id         int64       `gorm:"column:id;primaryKey"`
	StateType  MonitorType `gorm:"column:state_type"`
	LimitValue float64     `gorm:"column:limit_value"`
	StateValue float64     `gorm:"column:state_value"`
	Occurrence time.Time   `gorm:"column:occurrence"`
	IsNotify   int64       `gorm:"column:is_notify"`
	Day        int         `gorm:"column:day"`
	CreateAt   time.Time   `gorm:"column:create_at"`
	UpdateAt   time.Time   `gorm:"column:update_at"`
}

func (SystemMonitorWarning) TableName() string {
	return "system_monitor_warning"
}

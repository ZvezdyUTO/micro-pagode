package model

import (
	"mp/internal/domain"
	"time"
)

type UserStatus int

const (
	UserStatusNormal UserStatus = iota + 1
	UserStatusFreeze
	UserStatusLogout
)

const IsSystemUser = 1

type Users struct {
	Id       int        `gorm:"column:id;primary_key"`
	Name     string     `gorm:"column:name"`
	Phone    string     `gorm:"column:phone"`
	Password string     `gorm:"column:password"`
	Status   UserStatus `gorm:"column:status"`
	IsSystem int64      `gorm:"column:is_system"`
	CreateAt time.Time  `gorm:"column:create_at"`
	UpdateAt time.Time  `gorm:"column:update_at"`
	DeleteAt time.Time  `gorm:"column:delete_at"`
}

func (m *Users) ToDomainUser() *domain.User {
	return &domain.User{
		Id:     m.Id,
		Name:   m.Name,
		Phone:  m.Phone,
		Status: int64(m.Status),
	}
}

func (m *Users) TableName() string {
	return "users"
}

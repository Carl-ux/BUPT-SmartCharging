package model

import (
	"gorm.io/gorm"
)

// 用户
type User struct {
	gorm.Model
	Name     string `gorm:"type:varchar(20);not null"`
	Password string `gorm:"size:255;not null"`
	Admin    bool   `gorm:"default:false"`
}

func (User) TableName() string {
	return "user"
}

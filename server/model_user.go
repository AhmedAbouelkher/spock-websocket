package main

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID UUID `json:"id" gorm:"primaryKey"`

	Name             string  `json:"name" gorm:"column:name"`
	ProfileImageIcon *string `json:"profile_image_icon" gorm:"column:profile_image_icon"`
	Email            string  `json:"email" gorm:"column:email"`
	Password         string  `json:"-" gorm:"column:password"`

	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (User) TableName() string { return "users" }

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID.IsEmpty() {
		u.ID = NewUUIDv4()
	}
	n := time.Now()
	u.CreatedAt = n
	u.UpdatedAt = n
	return
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()
	return
}

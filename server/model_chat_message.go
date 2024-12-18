package main

import (
	"time"

	"gorm.io/gorm"
)

type CMType string

const (
	CMTypeText     CMType = "text"
	CMTypeDocument CMType = "document"
	CMTypeImage    CMType = "image"
	CMTypeVideo    CMType = "video"
	CMTypeAudio    CMType = "audio"
	CMTypeLocation CMType = "location"
)

type ChatMessage struct {
	ID uint `json:"id" gorm:"primaryKey"`

	ChatRoomID UUID      `json:"chat_room_id" gorm:"column:chat_room_id"`
	ChatRoom   *ChatRoom `json:"chat_room,omitempty" gorm:"foreignKey:ChatRoomID;references:ID"`

	CreatedByID UUID  `json:"created_by_id" gorm:"column:created_by_id"`
	CreatedBy   *User `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID;references:ID"`

	Content string `json:"content" gorm:"column:content"`
	Type    CMType `json:"type" gorm:"column:type"`

	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	EditedAt  *time.Time     `json:"edited_at" gorm:"column:edited_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ChatMessage) TableName() string { return "chat_messages" }

func (msg *ChatMessage) BeforeCreate(tx *gorm.DB) (err error) {
	n := time.Now()
	msg.CreatedAt = n
	return
}

func (msg *ChatMessage) BeforeUpdate(tx *gorm.DB) (err error) {
	return
}

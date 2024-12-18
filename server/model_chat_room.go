package main

import (
	"time"

	"gorm.io/gorm"
)

type ChatRoom struct {
	ID UUID `json:"id" gorm:"primaryKey"`

	Name       string      `json:"name" gorm:"column:name"`
	UsersLimit int         `json:"users_limit" gorm:"column:users_limit"`
	UsersIDs   StringArray `json:"users_ids" gorm:"column:users_ids"`
	PeerToPeer *bool       `json:"peer_to_peer" gorm:"column:peer_to_peer"` // not null, but we are using gorm

	LatestMessageID *uint        `json:"latest_message_id,omitempty" gorm:"column:latest_message_id"`
	LatestMessage   *ChatMessage `json:"latest_message,omitempty" gorm:"foreignKey:LatestMessageID;references:ID"`

	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (ChatRoom) TableName() string { return "chat_rooms" }

func (cr *ChatRoom) BeforeCreate(tx *gorm.DB) (err error) {
	if cr.ID.IsEmpty() {
		cr.ID = NewUUIDv4()
	}
	if cr.UsersLimit == 0 {
		cr.UsersLimit = 2
	}
	if cr.PeerToPeer == nil {
		b := true
		cr.PeerToPeer = &b
	}
	n := time.Now()
	cr.CreatedAt = n
	cr.UpdatedAt = n
	return
}

func (cr *ChatRoom) BeforeUpdate(tx *gorm.DB) (err error) {
	cr.UpdatedAt = time.Now()
	return
}

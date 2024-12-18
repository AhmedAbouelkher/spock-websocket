package main

import "time"

type RoomType string

const (
	RTPrivate RoomType = "private"
	RTGroup   RoomType = "group"
)

type ChatRoomResource struct {
	RoomID    UUID      `json:"room_id"`
	Name      string    `json:"name"`
	Type      RoomType  `json:"type"`
	CreatedAt time.Time `json:"created_at"`

	NumberOfUsers *int     `json:"number_of_users,omitempty"`
	Users         []User   `json:"users,omitempty"`
	UserIDs       []string `json:"user_ids,omitempty"`

	OtherUser *User `json:"other_user,omitempty"`

	LastMessage *SentMessageResource `json:"last_message"`
}

type SentMessageResource struct {
	ID        uint       `json:"id"`
	Content   string     `json:"content"`
	Type      CMType     `json:"type"`
	SentAt    time.Time  `json:"sent_at"`
	EditedAt  *time.Time `json:"edited_at"`
	MyMassage bool       `json:"my_message"`
	SenderID  *UUID      `json:"sender_id,omitempty"`
	RoomID    *UUID      `json:"room_id,omitempty"`
	User      User       `json:"sent_by"`
}

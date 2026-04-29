package model

import "github.com/google/uuid"

type Message struct {
	Model
	SenderID   uuid.UUID `gorm:"type:varchar(36);not null;index" json:"sender_id"`
	ReceiverID uuid.UUID `gorm:"type:varchar(36);not null;index" json:"receiver_id"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	IsRead     bool      `gorm:"default:false" json:"is_read"`
}

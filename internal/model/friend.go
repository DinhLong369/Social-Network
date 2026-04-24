package model

import "github.com/google/uuid"

const (
	PENDING  = "pending"
	ACCEPTED = "accepted"
	REJECTED = "rejected"
)
// FriendRequest - Yêu cầu kết bạn
type FriendRequest struct {
	Model      `gorm:"embedded"`
	SenderID   uuid.UUID `json:"sender_id"`
	ReceiverID uuid.UUID `json:"receiver_id"`
	Status     string    `json:"status"` // pending, accepted, rejected
	Sender     *User     `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	Receiver   *User     `json:"receiver,omitempty" gorm:"foreignKey:ReceiverID"`
}

// Friendship - Quan hệ bạn bè
type Friendship struct {
	Model    `gorm:"embedded"`
	UserID   uuid.UUID `json:"user_id" gorm:"index"`
	FriendID uuid.UUID `json:"friend_id" gorm:"index"`
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Friend   *User     `json:"friend,omitempty" gorm:"foreignKey:FriendID"`
}

package model

import (
	"time"

	"github.com/google/uuid"
)

type OTP struct {
	Model      `gorm:"embedded"`
	Code       string     `gorm:"uniqueIndex" json:"code"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	Used       bool       `gorm:"default:false" json:"used,omitempty"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	ExpiredAt  time.Time  `json:"expired_at"`
}

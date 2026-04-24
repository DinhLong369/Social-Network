package repo

import (
	"core/app"
	"core/internal/model"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func CreateOTPForUser(userID uuid.UUID) (string, error) {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	expireTime := time.Now().Add(5 * time.Minute)

	otp := model.OTP{
		Code:      code,
		UserID:    &userID,
		Used:      false,
		ExpiredAt: expireTime,
	}

	if err := app.Database.DB.Create(&otp).Error; err != nil {
		return "", err
	}

	return otp.Code, nil
}

func GetLatestOtp(userID uuid.UUID, code string) (model.OTP, error) {
	var otp model.OTP
	err := app.Database.DB.Where("user_id = ? AND code = ?", userID, code).Order("created_at DESC").First(&otp).Error
	return otp, err
}

func MarkOtpAsVerified(otpID uuid.UUID) error {
	now := time.Now()
	return app.Database.DB.Model(&model.OTP{}).Where("id = ?", otpID).Update("verified_at", now).Error
}

func MarkOtpAsUsed(otpID uuid.UUID) error {
	return app.Database.DB.Model(&model.OTP{}).Where("id = ?", otpID).Update("used", true).Error
}

// Đánh dấu OTP của user là used (không thể reuse)
func MarkOtpAsUsedByUserID(userID uuid.UUID) error {
	return app.Database.DB.Model(&model.OTP{}).Where("user_id = ?", userID).Update("used", true).Error
}

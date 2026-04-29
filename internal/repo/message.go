package repo

import (
	"core/app"
	"core/internal/model"

	"github.com/google/uuid"
)

// SaveMessage lưu tin nhắn vào DB
func SaveMessage(senderID, receiverID uuid.UUID, content string) (*model.Message, error) {
	msg := &model.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
	}
	result := app.Database.DB.Create(msg)
	if result.Error != nil {
		return nil, result.Error
	}
	return msg, nil
}

// GetConversation lấy lịch sử hội thoại giữa 2 người dùng
func GetConversation(userID, partnerID string) ([]model.Message, error) {
	var messages []model.Message
	result := app.Database.DB.
		Where(
			"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID, partnerID, partnerID, userID,
		).
		Order("created_at ASC").
		Find(&messages)
	return messages, result.Error
}

// MarkMessagesAsRead đánh dấu đã đọc toàn bộ tin nhắn từ sender gửi đến receiver
func MarkMessagesAsRead(senderID, receiverID string) error {
	return app.Database.DB.Model(&model.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = false", senderID, receiverID).
		Update("is_read", true).Error
}

package repo

import (
	"core/app"
	"core/internal/model"
	"time"

	"github.com/google/uuid"
)

// ConversationSummary chứa thông tin đối tác + tin nhắn cuối trong hội thoại
type ConversationSummary struct {
	PartnerID       string     `json:"partner_id"`
	PartnerName     string     `json:"partner_name"`
	PartnerUsername string     `json:"partner_username"`
	PartnerAvatar   string     `json:"partner_avatar"`
	LastMessage     string     `json:"last_message"`
	LastMessageAt   *time.Time `json:"last_message_at"`
	IsRead          bool       `json:"is_read"`
	SenderID        string     `json:"sender_id"`
}

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

// GetConversationList lấy danh sách hội thoại của user, mỗi hội thoại kèm tin nhắn cuối
func GetConversationList(userID string) ([]ConversationSummary, error) {
	var results []ConversationSummary
	query := `
		SELECT
			u.id          AS partner_id,
			u.name        AS partner_name,
			u.username    AS partner_username,
			u.avatar      AS partner_avatar,
			m.content     AS last_message,
			m.created_at  AS last_message_at,
			m.is_read     AS is_read,
			m.sender_id   AS sender_id
		FROM messages m
		JOIN users u
			ON u.id = CASE WHEN m.sender_id = ? THEN m.receiver_id ELSE m.sender_id END
		WHERE (m.sender_id = ? OR m.receiver_id = ?)
		  AND m.deleted_at IS NULL
		  AND m.created_at = (
			  SELECT MAX(m2.created_at)
			  FROM messages m2
			  WHERE m2.deleted_at IS NULL
			    AND (
			        (m2.sender_id = m.sender_id AND m2.receiver_id = m.receiver_id)
			        OR (m2.sender_id = m.receiver_id AND m2.receiver_id = m.sender_id)
			    )
		  )
		ORDER BY m.created_at DESC
	`
	err := app.Database.DB.Raw(query, userID, userID, userID).Scan(&results).Error
	return results, err
}

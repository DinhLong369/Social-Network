package handler

import (
	"core/app"
	"core/internal/repo"
	"encoding/json"
	"log"
	"os"

	fiberws "github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// incomingMsg là format JSON client gửi lên qua WebSocket
type incomingMsg struct {
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
}

// WSAuthMiddleware xác thực JWT từ query param ?token=... trước khi upgrade WebSocket
func WSAuthMiddleware(c fiber.Ctx) error {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		return fiber.ErrUnauthorized
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(os.Getenv("SECRETKEY_USER")), nil
	})
	if err != nil || !token.Valid {
		return fiber.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fiber.ErrUnauthorized
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return fiber.ErrUnauthorized
	}

	if _, err := uuid.Parse(userIDStr); err != nil {
		return fiber.ErrUnauthorized
	}

	c.Locals("user_id", userIDStr)
	return c.Next()
}

// ChatWebSocket xử lý kết nối WebSocket realtime
func ChatWebSocket(c *fiberws.Conn) {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		c.Close()
		return
	}

	app.GlobalHub.Register(userID, c)
	defer app.GlobalHub.Unregister(userID)

	log.Printf("[WS] user %s connected", userID)

	for {
		_, rawMsg, err := c.ReadMessage()
		if err != nil {
			log.Printf("[WS] user %s disconnected: %v", userID, err)
			break
		}

		var incoming incomingMsg
		if err := json.Unmarshal(rawMsg, &incoming); err != nil {
			continue
		}

		receiverUUID, err := uuid.Parse(incoming.ReceiverID)
		if err != nil || incoming.Content == "" {
			continue
		}

		senderUUID, _ := uuid.Parse(userID)

		// Lưu tin nhắn vào DB
		msg, err := repo.SaveMessage(senderUUID, receiverUUID, incoming.Content)
		if err != nil {
			log.Printf("[WS] save message error: %v", err)
			continue
		}

		// Serialize và gửi cho cả 2 phía
		out, _ := json.Marshal(msg)
		app.GlobalHub.Send(incoming.ReceiverID, out) // gửi đến người nhận
		app.GlobalHub.Send(userID, out)              // echo lại cho người gửi
	}
}

// GetMessages trả về lịch sử hội thoại giữa user hiện tại và partner
func GetMessages(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	partnerID := c.Query("partner_id")

	if _, err := uuid.Parse(partnerID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid partner_id",
		})
	}

	messages, err := repo.GetConversation(userID, partnerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to load messages",
		})
	}

	// Đánh dấu đã đọc các tin nhắn từ partner gửi đến mình
	_ = repo.MarkMessagesAsRead(partnerID, userID)

	return c.JSON(fiber.Map{
		"status": true,
		"data":   messages,
	})
}

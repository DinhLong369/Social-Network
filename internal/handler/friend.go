package handler

import (
	"core/internal/model"
	"core/internal/repo"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func SendFriendRequest(c fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "error": nil, "message": "Please login again"})
	}
	receiver_id := c.Query("receiver_id")
	if receiver_id == "" {
		return c.JSON(fiber.Map{"status": false, "message": "Receiver ID is required"})
	}
	receiverUUID, err := uuid.Parse(receiver_id)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid Receiver ID"})
	}
	friendship := repo.CheckFriendship(user.ID, receiverUUID)
	if friendship {
		return c.JSON(fiber.Map{"status": false, "message": "You are already friends or have a pending request"})
	}
	friendRequest := model.FriendRequest{
		SenderID:   user.ID,
		ReceiverID: receiverUUID,
		Status:     model.PENDING,
	}
	if err := repo.SendFriendRequest(friendRequest); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to send friend request", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Friend request sent successfully"})
}

func AcceptFriendRequest(c fiber.Ctx) error {
	request_id := c.Query("request_id")
	if request_id == "" {
		return c.JSON(fiber.Map{"status": false, "message": "Request ID is required"})
	}
	requestUUID, err := uuid.Parse(request_id)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid Request ID"})
	}
	if err := repo.AcceptFriendRequest(requestUUID); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to accept friend request", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Friend request accepted successfully"})
}

func RejectFriendRequest(c fiber.Ctx) error {
	request_id := c.Query("request_id")
	if request_id == "" {
		return c.JSON(fiber.Map{"status": false, "message": "Request ID is required"})
	}
	requestUUID, err := uuid.Parse(request_id)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid Request ID"})
	}
	if err := repo.RejectFriendRequest(requestUUID); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to reject friend request", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Friend request rejected successfully"})
}

func GetPendingFriendRequests(c fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "message": "Please login again"})
	}
	requests, err := repo.GetPendingFriendRequests(user.ID)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to get pending friend requests", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Pending friend requests retrieved successfully", "data": requests})
}

func GetFriends(c fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "message": "Please login again"})
	}
	friends, err := repo.GetFriendsByUserIDFromCache(user.ID.String())
	log.Println("🚀 ~ file: friend.go ~ line 86 ~ funcGetFriends ~ friends : ", friends)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to get friends from cache", "error": err.Error()})
	}
	if friends == nil {
		friends, err = repo.GetFriends(user.ID)
		if err != nil {
			return c.JSON(fiber.Map{"status": false, "message": "Failed to get friends", "error": err.Error()})
		}
		if err := repo.SetFriendsByUserIDCache(user.ID.String(), friends); err != nil {
			log.Println("warning: cannot set friends cache:", err)
		}
	}
	return c.JSON(fiber.Map{"status": true, "message": "Friends retrieved successfully", "data": friends})
}

func RemoveFriend(c fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "message": "Please login again"})
	}
	friend_id := c.Query("friend_id")
	if friend_id == "" {
		return c.JSON(fiber.Map{"status": false, "message": "Friend ID is required"})
	}
	friendUUID, err := uuid.Parse(friend_id)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid Friend ID"})
	}
	if err := repo.RemoveFriend(user.ID, friendUUID); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to remove friend", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Friend removed successfully"})
}

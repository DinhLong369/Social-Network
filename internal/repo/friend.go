package repo

import (
	"core/app"
	"core/internal/model"
	"fmt"

	"github.com/google/uuid"
)

func CheckFriendship(userID, receiverID uuid.UUID) bool {
	tx := app.Database.DB.Where("user_id = ? AND friend_id = ?", userID, receiverID).
		First(&model.Friendship{})
	return tx.Error == nil
}

func SendFriendRequest(friendRequest model.FriendRequest) error {
	tx := app.Database.DB.Create(&friendRequest)
	return tx.Error
}

func AcceptFriendRequest(requestID uuid.UUID) error {
	var friendReq model.FriendRequest
	if err := app.Database.DB.First(&friendReq, requestID).Error; err != nil {
		return err
	}

	// Tạo quan hệ bạn bè 2 chiều
	tx := app.Database.DB.Begin(nil, nil)

	// Tạo friendship từ sender → receiver
	f1 := model.Friendship{
		UserID:   friendReq.SenderID,
		FriendID: friendReq.ReceiverID,
	}
	if err := tx.Create(&f1).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Tạo friendship từ receiver → sender
	f2 := model.Friendship{
		UserID:   friendReq.ReceiverID,
		FriendID: friendReq.SenderID,
	}
	if err := tx.Create(&f2).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update status
	if err := tx.Model(&friendReq).Update("status", model.ACCEPTED).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return InvalidateFriendsByUserIDCache(friendReq.SenderID.String(), friendReq.ReceiverID.String())
}

// Từ chối yêu cầu kết bạn
func RejectFriendRequest(requestID uuid.UUID) error {
	var friendReq model.FriendRequest
	if err := app.Database.DB.First(&friendReq, requestID).Error; err != nil {
		return fmt.Errorf("friend request not found")
	}

	return app.Database.DB.Model(&friendReq).Update("status", model.REJECTED).Error
}

// Lấy danh sách yêu cầu kết bạn chưa xử lý
func GetPendingFriendRequests(userID uuid.UUID) ([]model.FriendRequest, error) {
	var requests []model.FriendRequest
	err := app.Database.DB.
		Where("receiver_id = ? AND status = ?", userID, model.PENDING).
		Preload("Sender").
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

// Lấy danh sách bạn bè
func GetFriends(userID uuid.UUID) ([]model.User, error) {
	var friends []model.User
	err := app.Database.DB.
		Table("friendships").
		Select("users.*").
		Joins("JOIN users ON friendships.friend_id = users.id").
		Where("friendships.user_id = ?", userID).
		Scan(&friends).Error
	return friends, err
}

// Xóa bạn bè
func RemoveFriend(userID, friendID uuid.UUID) error {
	tx := app.Database.DB.Begin(nil, nil)

	// Xóa 2 chiều
	if err := tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friendID, friendID, userID).Delete(&model.Friendship{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	err := tx.Commit().Error
	if err != nil {
		return err
	}

	return InvalidateFriendsByUserIDCache(userID.String(), friendID.String())
}

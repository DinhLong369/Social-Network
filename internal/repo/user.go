package repo

import (
	"core/app"
	"core/internal/model"
	"fmt"

	"github.com/google/uuid"
)

func Register(user model.User) error {
	tx := app.Database.DB.Create(&user)
	return tx.Error
}

func FindUsersByUsername(username string) (*model.User, error) {
	var user model.User
	tx := app.Database.DB.Where("username = ?", username).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func ExistInfo(username, email string) error {
	var count int64
	tx := app.Database.DB.Model(&model.User{})
	if username != "" {
		tx = tx.Where("username = ?", username).Count(&count)
		if count > 0 {
			return fmt.Errorf("username already exists")
		}

	}
	if email != "" {
		tx = tx.Where("email = ?", email).Count(&count)
		if count > 0 {
			return fmt.Errorf("email already exists")
		}
	}
	return nil
}

func GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	tx := app.Database.DB.Where("email = ?", email).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func VerifyUserEmail(userID uuid.UUID) error {
	tx := app.Database.DB.Model(&model.User{}).Where("id = ?", userID).Update("verified", true)
	return tx.Error
}
func GetUserByID(userID uuid.UUID) (*model.User, error) {
	var user model.User
	tx := app.Database.DB.Where("id = ?", userID).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func UserLoginByEmail(email string) (*model.User, error) {
	var user model.User
	tx := app.Database.DB.Where("email = ?", email).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &user, nil
}

func ChangePassword(userID uuid.UUID, hashedPassword string) error {
	return app.Database.DB.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

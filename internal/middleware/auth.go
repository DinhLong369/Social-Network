package middleware

import (
	"core/internal/repo"
	"os"
	"strings"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Middleware xác thực user
func AuthenUser(c fiber.Ctx) error {
	if strings.Contains(c.Path(), "/api/login.json") {
		return c.Next()
	} else {
		return jwtware.New(jwtware.Config{
			SigningKey:     jwtware.SigningKey{Key: []byte(os.Getenv("SECRETKEY_USER"))},
			ErrorHandler:   jwtError,
			SuccessHandler: jwtUserSuccess,
		})(c)
	}
}

// Xử lý thành công cho user
func jwtUserSuccess(c fiber.Ctx) error {
	token := jwtware.FromContext(c)
	if token == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": "Invalid token payload"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": "Invalid token payload"})
	}

	userId, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": "Invalid token payload"})
	}

	user_id, err := uuid.Parse(userId)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": "Invalid user id"})
	}

	user, err := repo.GetUserByID(user_id)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": "User not found"})
	}

	c.Locals("user", *user)
	c.Locals("user_id", user.ID.String())
	_ = repo.SetUserOnline(user.ID.String())
	return c.Next()
}

func jwtError(c fiber.Ctx, err error) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": "Invalid or expired JWT"})
}

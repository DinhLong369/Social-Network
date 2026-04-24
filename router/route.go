package router

import (
	mainapp "core/app"
	"core/internal/handler"
	"core/internal/middleware"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func Setup() {
	app := fiber.New(fiber.Config{})
	app.Use(cors.New())
	app.Use(recover.New())
	setupRouter(app)
	port := mainapp.Config("WEB_PORT")
	if len(port) == 0 {
		port = "8386"
	}
	fmt.Println("port=", port)
	app.Listen(":" + port)
}

func setupRouter(fiberApp *fiber.App) {
	api := fiberApp.Group("/api", logger.New())
	api.Get("", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": true, "message": "API is working"})
	})

	api.Post("/register.json", handler.Register)
	api.Post("/login.json", handler.Login)
	api.Post("/verify_account.json", handler.VerifyAccount)
	api.Post("/forgot-password.json", handler.ForgetPassword)
	api.Post("/verify_otp_password.json", handler.VerifyOTPPassword)
	api.Post("/change-password.json", handler.ChangePassword)

	api.Use(middleware.AuthenUser)
	//friendship routes
	api.Get("/friendship.json", handler.GetFriends)
	api.Post("/send_friend_request.json", handler.SendFriendRequest)
	api.Post("/accept_friend_request.json", handler.AcceptFriendRequest)
	api.Post("/reject_friend_request.json", handler.RejectFriendRequest)
	api.Get("/friend_requests.json", handler.GetPendingFriendRequests)
	api.Post("/unfriend.json", handler.RemoveFriend)

}

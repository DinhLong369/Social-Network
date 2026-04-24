package handler

import (
	"core/app"
	"core/internal/model"
	"core/internal/repo"
	"core/internal/utils"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// Helper function - Validate OTP
func validateOTP(email, otpCode string) (*model.User, *model.OTP, error) {
	if strings.TrimSpace(email) == "" {
		return nil, nil, fmt.Errorf("email cannot be empty")
	}

	user, err := repo.GetUserByEmail(email)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	otp, err := repo.GetLatestOtp(user.ID, otpCode)
	if err != nil {
		return nil, nil, fmt.Errorf("OTP not found or invalid")
	}

	if otp.Used {
		return nil, nil, fmt.Errorf("OTP code has been used")
	}

	if time.Now().After(otp.ExpiredAt) {
		return nil, nil, fmt.Errorf("OTP has expired")
	}

	return user, &otp, nil
}
func Register(c fiber.Ctx) error {
	type Input struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	var input Input
	if err := c.Bind().Body(&input); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid input", "error": err.Error()})
	}
	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to hash password", "error": err.Error()})
	}

	if input.Username == "" || input.Email == "" || input.Password == "" {
		return c.JSON(fiber.Map{"status": false, "message": "All fields are required"})
	}
	err = repo.ExistInfo(input.Username, input.Email)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": err.Error()})
	}
	user := model.User{
		Username: input.Username,
		Email:    input.Email,
		Password: hashedPassword,
	}
	if err := repo.Register(user); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to register user", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": true, "message": "User registered successfully"})
}

func Login(c fiber.Ctx) error {
	type Input struct {
		Email    string `json:"email" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	type DataUserReturn struct {
		User         model.User `json:"user"`
		AccessToken  string     `json:"access_token"`
		RefreshToken string     `json:"refresh_token"`
	}
	var input Input
	if err := c.Bind().Body(&input); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid input", "error": err.Error()})
	}
	if input.Email == "" || input.Password == "" {
		return c.JSON(fiber.Map{"status": false, "message": "All fields are required"})
	}
	user, err := repo.GetUserByEmail(input.Email)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": err.Error()})
	}
	if user.Verified == false {

		go func() {
			otpCode, err := repo.CreateOTPForUser(user.ID)
			if err != nil {
				log.Printf("failed to create OTP: %v", err)
			}
			// Gửi OTP qua email
			htmlContent := fmt.Sprintf(`
			<html>
			<body>
				<p>Mã OTP xác thực email của bạn là:</p>
				<h2>%s</h2>
				<p>Mã có hiệu lực trong 5 phút.</p>
				<p>Nếu bạn không yêu cầu thao tác này, hãy bỏ qua email.</p>
			</body>
			</html>`, otpCode)
			if err := utils.SendMail(user.Email, "Mã OTP xác thực email", htmlContent); err != nil {
				log.Printf("failed to send OTP email: %v", err)
			}
		}()

		return c.JSON(fiber.Map{"status": false, "message": "Email not verified"})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid password"})
	}
	claims_new := jwt.MapClaims{
		"user_id": user.ID,
		"name":    user.Name,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 ngày
	}
	claims_refresh := jwt.MapClaims{
		"user_id": user.ID,
		"name":    user.Name,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24 * 30).Unix(), // 30 ngày
	}
	token_new := jwt.NewWithClaims(jwt.SigningMethodHS256, claims_new)
	acesss_token, errs := token_new.SignedString([]byte(app.Config("SECRETKEY_USER")))
	if errs != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Token generation failed", "error": errs.Error(), "user": nil})
	}
	token_refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, claims_refresh)
	refresh_token, errs1 := token_refresh.SignedString([]byte(app.Config("SECRETKEY_USER")))
	if errs1 != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Token generation failed", "error": errs1.Error(), "user": nil})
	}
	data := DataUserReturn{
		User:         *user,
		AccessToken:  acesss_token,
		RefreshToken: refresh_token,
	}
	return c.JSON(fiber.Map{"status": true, "message": "Login successful", "data": data})
}

func VerifyAccount(c fiber.Ctx) error {
	type Input struct {
		OtpCode string `json:"otp_code"`
		Email   string `json:"email"`
	}

	var input Input
	if err := c.Bind().Body(&input); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid input"})
	}

	user, otp, err := validateOTP(input.Email, input.OtpCode)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": err.Error()})
	}

	if err := repo.MarkOtpAsUsed(otp.ID); err != nil {
		logrus.Error("Failed to mark OTP as used:", err)
	}

	if err := repo.VerifyUserEmail(user.ID); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to verify email", "error": err.Error()})
	}

	return c.JSON(fiber.Map{"status": true, "message": "Email verified successfully"})
}

func ForgetPassword(c fiber.Ctx) error {
	type Input struct {
		Email string `json:"email"`
	}
	var input Input
	if err := c.Bind().Body(&input); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Review your input", "error": err.Error()})
	}
	user, err := repo.GetUserByEmail(input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(fiber.Map{"status": false, "message": "Email not found"})
		}
		return c.JSON(fiber.Map{"status": false, "message": "Failed to get user", "error": err.Error()})
	}

	otpCode, err := repo.CreateOTPForUser(user.ID)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to create OTP", "error": err.Error()})
	}

	htmlContent := fmt.Sprintf(`
	<html>
	<body>
		<p>Yêu cầu đặt lại mật khẩu</p>
		<p>Mã xác thực của bạn là:</p>
		<h2>%s</h2>
		<p>Mã có hiệu lực trong 5 phút.</p>
		<p>Nếu bạn không yêu cầu thao tác này, hãy bỏ qua email.</p>
	</body>
	</html>`, otpCode)

	if err := utils.SendMail(user.Email, "Mã xác thực đặt lại mật khẩu", htmlContent); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to send reset password email", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Reset password email sent successfully"})
}

func VerifyOTPPassword(c fiber.Ctx) error {
	type Input struct {
		OtpCode string `json:"otp_code"`
		Email   string `json:"email"`
	}

	var input Input
	if err := c.Bind().Body(&input); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid input"})
	}

	user, otp, err := validateOTP(input.Email, input.OtpCode)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": err.Error()})
	}

	// Mark OTP as verified
	if err := repo.MarkOtpAsVerified(otp.ID); err != nil {
		logrus.Error("Failed to mark OTP as verified:", err)
	}

	// Tạo reset password token (5 phút)
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"type":    "password_reset",
		"exp":     time.Now().Add(5 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	resetToken, err := token.SignedString([]byte(app.Config("SECRET")))
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to generate reset token"})
	}

	return c.JSON(fiber.Map{
		"status":      true,
		"message":     "OTP verified successfully!",
		"reset_token": resetToken,
	})
}

func ChangePassword(c fiber.Ctx) error {
	type Input struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
		ResetToken      string `json:"reset_token"`
	}
	var input Input
	if err := c.Bind().Body(&input); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid input", "error": err.Error()})
	}

	if input.ResetToken == "" {
		return c.JSON(fiber.Map{"status": false, "message": "Reset token is required"})
	}

	if input.Password != input.ConfirmPassword {
		return c.JSON(fiber.Map{"status": false, "message": "Password and re-enter password do not match"})
	}

	// Xác thực reset token
	token, err := jwt.ParseWithClaims(input.ResetToken, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(app.Config("SECRET")), nil
	})
	if err != nil || !token.Valid {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid or expired reset token"})
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid token claims"})
	}

	// Kiểm tra token type
	if (*claims)["type"] != "password_reset" {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid token type"})
	}

	userID, err := uuid.Parse((*claims)["user_id"].(string))
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Invalid user ID in token"})
	}

	// Lấy user để verify
	user, err := repo.GetUserByID(userID)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "User not found"})
	}

	hash_password, err := hashPassword(input.Password)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to hash password", "error": err.Error()})
	}

	// Cập nhật mật khẩu
	if err := repo.ChangePassword(user.ID, hash_password); err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Update user failed", "error": err.Error()})
	}

	// Đánh dấu OTP là used (không thể reuse)
	if err := repo.MarkOtpAsUsedByUserID(user.ID); err != nil {
		logrus.Error("Failed to mark OTP as used:", err)
	}

	return c.JSON(fiber.Map{"status": true, "message": "Password changed successfully"})
}

func FindUser(c fiber.Ctx) error {
	query := c.Query("username")
	if query == "" {
		return c.JSON(fiber.Map{"status": false, "message": "Username query is required"})
	}
	users, err := repo.FindUsersByUsername(query)
	if err != nil {
		return c.JSON(fiber.Map{"status": false, "message": "Failed to search users", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": true, "message": "Users retrieved successfully", "data": users})
}
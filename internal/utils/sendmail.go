package utils

import (
	"log"
	"os"

	"gopkg.in/gomail.v2"
)

func SendMail(to, subject, html string) error {
	// Lấy cấu hình SMTP từ environment variables
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("SMTP_USER")

	// Kiểm tra các biến cần thiết
	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPassword == "" || fromEmail == "" {
		log.Fatal("Missing SMTP configuration in environment variables")
	}

	// Chuyển port từ string sang int
	port := 587
	if smtpPort == "465" {
		port = 465
	} else if smtpPort == "25" {
		port = 25
	}

	// Tạo message
	m := gomail.NewMessage()
	m.SetHeader("From", fromEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", html)

	// Tạo dialer
	d := gomail.NewDialer(smtpHost, port, smtpUser, smtpPassword)

	// Gửi email
	if err := d.DialAndSend(m); err != nil {
		log.Printf("failed to send email: %v", err)
		return err
	}

	log.Printf("email sent successfully to: %s", to)
	return nil
}

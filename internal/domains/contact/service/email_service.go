package service

import (
	"fmt"
	"net/smtp"
	"os"

	"api/internal/domains/contact/dto"
)
// SendContactRequest sends an email with the contact request details.
// It uses environment variables for SMTP configuration and recipient email.
// It returns an error if sending fails.
func SendContactRequest(req dto.ContactRequest) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")

	// Send to both Rise email addresses
	recipients := []string{
		"info@risesportscomplex.com",
		"riseballtech@gmail.com",
	}

	subject := "New Contact Request"
	body := fmt.Sprintf("Name: %s\nEmail: %s\nPhone: %s\nMessage:\n%s", req.Name, req.Email, req.Phone, req.Message)

	// Format recipients for email header
	toHeader := "info@risesportscomplex.com, riseballtech@gmail.com"
	msg := "From: " + from + "\n" +
		"To: " + toHeader + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, recipients, []byte(msg))
	if err != nil {
		fmt.Printf("❌ SMTP error: %v\n", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

package service

import (
	"fmt"
	"net/smtp"
	"os"

	"api/internal/domains/contact/dto"
)

func SendContactRequest(req dto.ContactRequest) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	to := os.Getenv("CONTACT_RECIPIENT_EMAIL")

	subject := "New Contact Request"
	body := fmt.Sprintf("Name: %s\nEmail: %s\nPhone: %s\nMessage:\n%s", req.Name, req.Email, req.Phone, req.Message)

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))
	if err != nil {
		fmt.Printf("❌ SMTP error: %v\n", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

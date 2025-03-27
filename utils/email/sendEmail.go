package email

import (
	"api/config"
	errLib "api/internal/libs/errors"
	"log"
	"net/http"
	"net/smtp"
)

func SendEmail(to, subject, body string) *errLib.CommonError {
	// Email credentials
	from := "klintlee1@gmail.com"
	password := config.Env.GmailSmtpPassword

	// Gmail SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Prepare the email message
	message := []byte("MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + body)

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending the email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return errLib.New("Failed to send email", http.StatusInternalServerError)
	}

	return nil
}

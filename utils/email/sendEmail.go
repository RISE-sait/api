package email

import (
	"api/config"
	errLib "api/internal/libs/errors"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
)

func SendConfirmChildEmail(to, child string) *errLib.CommonError {
	// Email credentials
	from := "klintlee1@gmail.com"
	password := config.Envs.GmailSmtpPassword

	// Gmail SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Email message
	subject := "Subject: Confirmation of child\n"
	body := fmt.Sprintf("Click on the link to confirm the following email: %s, as your child", child)
	message := []byte(subject + "\n" + body)

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

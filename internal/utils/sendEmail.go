package utils

import (
	"api/config"
	"fmt"
	"log"
	"net/smtp"
)

func SendEmail(to string) {
	// Email credentials
	from := "klintlee1@gmail.com"
	password := config.Envs.GmailSmtpPassword

	// Gmail SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Email message
	subject := "Subject: THIS IS A SCAM\n"
	body := "A CHINESE CHING CHONG SCAM!"
	message := []byte(subject + "\n" + body)

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending the email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	fmt.Println("Email sent successfully!")
}

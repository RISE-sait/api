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

	log.Println("Sending email to: ", to)
	log.Println("Child: ", child)
	log.Println("Password: ", config.Envs.GmailSmtpPassword)

	// Email credentials
	from := "klintlee1@gmail.com"
	password := config.Envs.GmailSmtpPassword

	// Gmail SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Email message
	subject := "Subject: Confirmation of child"
	body := fmt.Sprintf(`
		<html>
			<body>
				<p>Click on the link below to confirm the following email: %s, as your child.</p>
				<p><b>Child:</b> %s</p>
				<p><b>Parent:</b> %s</p>
				<a href="http://localhost:8080/api/confirm-child?child=%s&parent=%s">Confirm Child</a>
			</body>
		</html>`, child, child, to, child, to)

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

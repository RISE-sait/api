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
	from := "riseballtech@gmail.com"
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
// SendSignUpConfirmationEmail sends a welcome message to newly registered users.
func SendSignUpConfirmationEmail(to, firstName string) {
	body := SignUpConfirmationBody(firstName)
	if err := SendEmail(to, "Welcome to Rise", body); err != nil {
		log.Println("failed to send signup email:", err.Message)
	}
}

// SendMembershipPurchaseEmail sends a confirmation email after a membership purchase.
func SendMembershipPurchaseEmail(to, firstName, plan string) {
	body := MembershipPurchaseBody(firstName, plan)
	if err := SendEmail(to, "Membership Purchase Confirmation", body); err != nil {
		log.Println("failed to send membership email:", err.Message)
	}
}

// SendEmailVerification sends an email verification link to newly registered users.
func SendEmailVerification(to, firstName, verificationURL string) *errLib.CommonError {
	body := EmailVerificationBody(firstName, verificationURL)
	if err := SendEmail(to, "Verify Your Email - Rise", body); err != nil {
		log.Println("failed to send verification email:", err.Message)
		return err
	}
	log.Printf("Verification email sent successfully to %s", to)
	return nil
}

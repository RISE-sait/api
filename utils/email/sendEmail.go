package email

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"

	"api/config"
	errLib "api/internal/libs/errors"
)

func SendEmail(to, subject, body string) *errLib.CommonError {
	// Email credentials
	from := "info@risesportscomplex.com"
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

// SendSubsidyApprovedEmail sends a notification when a subsidy is approved for a customer
func SendSubsidyApprovedEmail(to, firstName, providerName string, amount float64, validUntil string) {
	body := SubsidyApprovedBody(firstName, providerName, amount, validUntil)
	if err := SendEmail(to, "Subsidy Approved - Rise", body); err != nil {
		log.Println("failed to send subsidy approved email:", err.Message)
	} else {
		log.Printf("Subsidy approved email sent successfully to %s", to)
	}
}

// SendSubsidyUsedEmail sends a notification when subsidy is used (non-depleting transactions)
func SendSubsidyUsedEmail(to, firstName string, amountUsed, remainingBalance float64, transactionType string) {
	body := SubsidyUsedBody(firstName, amountUsed, remainingBalance, transactionType)
	if err := SendEmail(to, "Subsidy Applied - Rise", body); err != nil {
		log.Println("failed to send subsidy used email:", err.Message)
	} else {
		log.Printf("Subsidy used email sent successfully to %s", to)
	}
}

// SendSubsidyDepletedEmail sends a notification when subsidy is fully depleted
func SendSubsidyDepletedEmail(to, firstName string, totalUsed float64) {
	body := SubsidyDepletedBody(firstName, totalUsed)
	if err := SendEmail(to, "Subsidy Fully Used - Rise", body); err != nil {
		log.Println("failed to send subsidy depleted email:", err.Message)
	} else {
		log.Printf("Subsidy depleted email sent successfully to %s", to)
	}
}

// SendSubsidyExpiringEmail sends a notification when subsidy is about to expire
func SendSubsidyExpiringEmail(to, firstName string, remainingBalance float64, expiryDate string) {
	body := SubsidyExpiringBody(firstName, remainingBalance, expiryDate)
	if err := SendEmail(to, "Subsidy Expiring Soon - Rise", body); err != nil {
		log.Println("failed to send subsidy expiring email:", err.Message)
	} else {
		log.Printf("Subsidy expiring email sent successfully to %s", to)
	}
}

// SendPaymentFailedEmail sends a notification when a membership payment fails
func SendPaymentFailedEmail(to, firstName, membershipPlan, updatePaymentURL string) {
	body := PaymentFailedBody(firstName, membershipPlan, updatePaymentURL)
	if err := SendEmail(to, "Action Required: Payment Failed - Rise", body); err != nil {
		log.Println("failed to send payment failed email:", err.Message)
	} else {
		log.Printf("Payment failed email sent successfully to %s", to)
	}
}

// SendPaymentFailedReminderEmail sends a reminder about failed payment
func SendPaymentFailedReminderEmail(to, firstName, membershipPlan, updatePaymentURL string, daysUntilSuspension int) {
	body := PaymentFailedReminderBody(firstName, membershipPlan, updatePaymentURL, daysUntilSuspension)
	if err := SendEmail(to, "Reminder: Update Payment Method - Rise", body); err != nil {
		log.Println("failed to send payment failed reminder email:", err.Message)
	} else {
		log.Printf("Payment failed reminder email sent successfully to %s", to)
	}
}

// SendAccountRecoveryEmail sends a password reset link to users who need to recover their account
func SendAccountRecoveryEmail(to, resetURL string) error {
	body := AccountRecoveryBody(resetURL)
	if err := SendEmail(to, "Reset Your Password - Rise", body); err != nil {
		log.Printf("failed to send account recovery email to %s: %s", to, err.Message)
		return fmt.Errorf(err.Message)
	}
	log.Printf("Account recovery email sent successfully to %s", to)
	return nil
}

// SendEmailChangeVerification sends a verification link to the new email address for email change
func SendEmailChangeVerification(to, firstName, newEmail, verificationURL string) *errLib.CommonError {
	body := EmailChangeVerificationBody(firstName, newEmail, verificationURL)
	if err := SendEmail(to, "Verify Your New Email - Rise", body); err != nil {
		log.Println("failed to send email change verification email:", err.Message)
		return err
	}
	log.Printf("Email change verification sent successfully to %s", to)
	return nil
}

// SendMembershipCheckoutLinkEmail sends a checkout link email to a customer for admin-initiated membership assignment
func SendMembershipCheckoutLinkEmail(to, firstName, planName, checkoutURL string) {
	body := MembershipCheckoutLinkBody(firstName, planName, checkoutURL)
	if err := SendEmail(to, "Complete Your Membership - Rise", body); err != nil {
		log.Println("failed to send membership checkout link email:", err.Message)
	} else {
		log.Printf("Membership checkout link email sent successfully to %s", to)
	}
}

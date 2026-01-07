//go:build ignore

package main

import (
	"fmt"
	"os"

	"api/utils/email"
)

func main() {
	// Generate all email templates for preview
	templates := map[string]string{
		"01_signup.html":           email.SignUpConfirmationBody("John"),
		"02_membership.html":       email.MembershipPurchaseBody("John", "Elite Training"),
		"03_verify_email.html":     email.EmailVerificationBody("John", "https://rise.com/verify?token=abc123"),
		"04_subsidy_approved.html": email.SubsidyApprovedBody("John", "City of Calgary", 500.00, "December 31, 2025"),
		"05_subsidy_used.html":     email.SubsidyUsedBody("John", 150.00, 350.00, "membership"),
		"06_subsidy_depleted.html": email.SubsidyDepletedBody("John", 500.00),
		"07_subsidy_expiring.html": email.SubsidyExpiringBody("John", 200.00, "January 15, 2025"),
		"08_payment_failed.html":   email.PaymentFailedBody("John", "Elite Training", "https://rise.com/update-payment"),
		"09_account_recovery.html": email.AccountRecoveryBody("https://rise.com/reset?token=xyz789"),
		"10_payment_reminder.html": email.PaymentFailedReminderBody("John", "Elite Training", "https://rise.com/update-payment", 3),
		"11_event_notification.html": email.EventNotificationBody("John", "Event Update: Basketball Training", "Time changed from Monday, January 5 at 10:00 AM to Monday, January 5 at 2:00 PM\nLocation changed from Court A to Court B"),
	}

	// Create preview directory
	os.MkdirAll("email_previews", 0755)

	for filename, content := range templates {
		path := "email_previews/" + filename
		os.WriteFile(path, []byte(content), 0644)
		fmt.Printf("Created: %s\n", path)
	}

	fmt.Println("\nDone! Open the HTML files in email_previews/ folder in your browser.")
}

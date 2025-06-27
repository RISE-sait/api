package email

import "fmt"

func SignUpConfirmationBody(firstName string) string {
	return fmt.Sprintf(`<p>Hi %s,</p><p>Thanks for creating an account at Rise.</p><p>Next steps:</p><ol><li>Download our mobile app and log in.</li><li>Complete your profile information.</li><li>Browse memberships and programs to get started.</li></ol><p>Welcome to the community!</p>`, firstName)
}

func MembershipPurchaseBody(firstName, plan string) string {
	return fmt.Sprintf(`<p>Hi %s,</p><p>Thank you for purchasing the %s membership.</p><p>Next steps:</p><ol><li>Check your dashboard for membership details.</li><li>Enjoy all the benefits available to you.</li><li>Contact us if you have any questions.</li></ol><p>We appreciate your support!</p>`, firstName, plan)
}

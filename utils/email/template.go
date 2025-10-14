package email

import "fmt"

func SignUpConfirmationBody(firstName string) string {
	return fmt.Sprintf(`<p>Hi %s,</p><p>Thanks for creating an account at Rise.</p><p>Next steps:</p><ol><li>Download our mobile app and log in.</li><li>Complete your profile information.</li><li>Browse memberships and programs to get started.</li></ol><p>Welcome to the community!</p>`, firstName)
}

func MembershipPurchaseBody(firstName, plan string) string {
	return fmt.Sprintf(`<p>Hi %s,</p><p>Thank you for purchasing the %s membership.</p><p>Next steps:</p><ol><li>Check your dashboard for membership details.</li><li>Enjoy all the benefits available to you.</li><li>Contact us if you have any questions.</li></ol><p>We appreciate your support!</p>`, firstName, plan)
}

func EmailVerificationBody(firstName, verificationURL string) string {
	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #007bff; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
				.content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
				.button { display: inline-block; padding: 12px 30px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
				.button:hover { background-color: #0056b3; }
				.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
				.warning { background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 10px; margin: 15px 0; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Verify Your Email Address</h1>
				</div>
				<div class="content">
					<p>Hi %s,</p>

					<p>Welcome to Rise! To complete your registration and start using your account, please verify your email address.</p>

					<p style="text-align: center;">
						<a href="%s" class="button">Verify Email Address</a>
					</p>

					<p>Or copy and paste this link into your browser:</p>
					<p style="background-color: white; padding: 10px; border: 1px solid #ddd; word-break: break-all;">
						%s
					</p>

					<div class="warning">
						<strong>Important:</strong> This verification link will expire in 24 hours.
					</div>

					<p>If you didn't create an account with Rise, you can safely ignore this email.</p>

					<div class="footer">
						<p>Thanks,<br>The Rise Team</p>
						<p>This is an automated message, please do not reply to this email.</p>
					</div>
				</div>
			</div>
		</body>
		</html>
	`, firstName, verificationURL, verificationURL)
}

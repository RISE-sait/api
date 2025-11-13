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

func SubsidyApprovedBody(firstName, providerName string, amount float64, validUntil string) string {
	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #28a745; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
				.content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
				.highlight { background-color: #d4edda; border-left: 4px solid #28a745; padding: 15px; margin: 20px 0; }
				.amount { font-size: 32px; font-weight: bold; color: #28a745; text-align: center; margin: 20px 0; }
				.info-box { background-color: white; border: 1px solid #ddd; padding: 15px; margin: 15px 0; border-radius: 5px; }
				.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>🎉 Subsidy Approved!</h1>
				</div>
				<div class="content">
					<p>Hi %s,</p>

					<p>Great news! You've been approved for a subsidy to help cover your membership costs at Rise Sports Complex.</p>

					<div class="amount">$%.2f</div>

					<div class="info-box">
						<strong>Subsidy Details:</strong>
						<ul style="margin: 10px 0;">
							<li><strong>Provider:</strong> %s</li>
							<li><strong>Amount:</strong> $%.2f</li>
							<li><strong>Valid Until:</strong> %s</li>
						</ul>
					</div>

					<div class="highlight">
						<strong>What happens next?</strong>
						<p style="margin: 10px 0;">Your subsidy will be automatically applied to your next membership purchase or renewal. You'll only pay the remaining balance after the subsidy is applied.</p>
					</div>

					<p>If you have any questions about your subsidy, please don't hesitate to contact us.</p>

					<div class="footer">
						<p>Thanks,<br>The Rise Team</p>
						<p>This is an automated message, please do not reply to this email.</p>
					</div>
				</div>
			</div>
		</body>
		</html>
	`, firstName, amount, providerName, amount, validUntil)
}

func SubsidyUsedBody(firstName string, amountUsed, remainingBalance float64, transactionType string) string {
	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #007bff; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
				.content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
				.transaction-box { background-color: white; border: 1px solid #ddd; padding: 20px; margin: 20px 0; border-radius: 5px; }
				.amount-row { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #eee; }
				.balance { background-color: #e3f2fd; border-left: 4px solid #007bff; padding: 15px; margin: 20px 0; font-size: 18px; }
				.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Subsidy Applied</h1>
				</div>
				<div class="content">
					<p>Hi %s,</p>

					<p>Your subsidy has been applied to your recent %s transaction.</p>

					<div class="transaction-box">
						<div class="amount-row">
							<span><strong>Subsidy Applied:</strong></span>
							<span style="color: #28a745; font-weight: bold;">$%.2f</span>
						</div>
						<div class="amount-row" style="border-bottom: none;">
							<span><strong>Remaining Balance:</strong></span>
							<span style="color: #007bff; font-weight: bold;">$%.2f</span>
						</div>
					</div>

					<div class="balance">
						💰 You still have <strong>$%.2f</strong> available for future memberships.
					</div>

					<p>Your subsidy will continue to be applied automatically to eligible transactions.</p>

					<div class="footer">
						<p>Thanks,<br>The Rise Team</p>
						<p>This is an automated message, please do not reply to this email.</p>
					</div>
				</div>
			</div>
		</body>
		</html>
	`, firstName, transactionType, amountUsed, remainingBalance, remainingBalance)
}

func SubsidyDepletedBody(firstName string, totalUsed float64) string {
	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #ffc107; color: #333; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
				.content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
				.notice-box { background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; }
				.summary { background-color: white; border: 1px solid #ddd; padding: 20px; margin: 20px 0; border-radius: 5px; text-align: center; }
				.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Subsidy Fully Used</h1>
				</div>
				<div class="content">
					<p>Hi %s,</p>

					<p>Your subsidy balance has been fully used.</p>

					<div class="summary">
						<p style="margin: 10px 0; color: #666;">Total Subsidy Used</p>
						<p style="font-size: 32px; font-weight: bold; color: #28a745; margin: 10px 0;">$%.2f</p>
					</div>

					<div class="notice-box">
						<strong>What's Next?</strong>
						<p style="margin: 10px 0;">Future membership purchases will be charged at the regular price. If you need assistance with membership costs, please contact our team to discuss available options.</p>
					</div>

					<p>Thank you for being part of the Rise community!</p>

					<div class="footer">
						<p>Thanks,<br>The Rise Team</p>
						<p>This is an automated message, please do not reply to this email.</p>
					</div>
				</div>
			</div>
		</body>
		</html>
	`, firstName, totalUsed)
}

func SubsidyExpiringBody(firstName string, remainingBalance float64, expiryDate string) string {
	return fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #dc3545; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
				.content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
				.urgent { background-color: #f8d7da; border-left: 4px solid #dc3545; padding: 15px; margin: 20px 0; }
				.balance-box { background-color: white; border: 1px solid #ddd; padding: 20px; margin: 20px 0; border-radius: 5px; text-align: center; }
				.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>⚠️ Subsidy Expiring Soon</h1>
				</div>
				<div class="content">
					<p>Hi %s,</p>

					<p>This is a reminder that your subsidy will expire soon.</p>

					<div class="balance-box">
						<p style="margin: 10px 0; color: #666;">Remaining Balance</p>
						<p style="font-size: 32px; font-weight: bold; color: #007bff; margin: 10px 0;">$%.2f</p>
						<p style="margin: 10px 0; color: #666;">Expires: <strong>%s</strong></p>
					</div>

					<div class="urgent">
						<strong>Action Required:</strong>
						<p style="margin: 10px 0;">To use your remaining subsidy balance, please purchase or renew your membership before the expiry date. After expiration, any unused balance will no longer be available.</p>
					</div>

					<p>If you have any questions, please contact us.</p>

					<div class="footer">
						<p>Thanks,<br>The Rise Team</p>
						<p>This is an automated message, please do not reply to this email.</p>
					</div>
				</div>
			</div>
		</body>
		</html>
	`, firstName, remainingBalance, expiryDate)
}

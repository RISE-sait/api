package email

import "fmt"

// Brand colors
const (
	colorBlack      = "#000000"
	colorYellow     = "#FFD700"
	colorDarkGray   = "#1a1a1a"
	colorLightGray  = "#f5f5f5"
	colorWhite      = "#ffffff"
	colorTextDark   = "#333333"
	colorTextMuted  = "#666666"
)

// baseTemplate wraps content in the Rise branded email template
func baseTemplate(title, content string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<style>
		body {
			font-family: 'Helvetica Neue', Arial, sans-serif;
			line-height: 1.6;
			color: %s;
			margin: 0;
			padding: 0;
			background-color: %s;
		}
		.wrapper {
			max-width: 600px;
			margin: 0 auto;
			padding: 20px;
		}
		.container {
			background-color: %s;
			border-radius: 8px;
			overflow: hidden;
			box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
		}
		.header {
			background-color: %s;
			padding: 30px 20px;
			text-align: center;
		}
		.logo {
			font-size: 42px;
			font-weight: 900;
			letter-spacing: 8px;
			color: %s;
			margin: 0;
			text-transform: uppercase;
		}
		.tagline {
			font-size: 10px;
			letter-spacing: 3px;
			color: %s;
			margin-top: 8px;
			text-transform: uppercase;
		}
		.title-bar {
			background-color: %s;
			padding: 15px 20px;
			text-align: center;
		}
		.title-bar h1 {
			color: %s;
			margin: 0;
			font-size: 18px;
			font-weight: 700;
			text-transform: uppercase;
			letter-spacing: 2px;
		}
		.content {
			padding: 30px;
		}
		.button {
			display: inline-block;
			padding: 14px 35px;
			background-color: %s;
			color: %s !important;
			text-decoration: none;
			border-radius: 4px;
			font-weight: 700;
			text-transform: uppercase;
			letter-spacing: 1px;
			font-size: 14px;
		}
		.button:hover {
			background-color: #e6c200;
		}
		.button-dark {
			background-color: %s;
			color: %s !important;
		}
		.info-box {
			background-color: %s;
			border-left: 4px solid %s;
			padding: 20px;
			margin: 20px 0;
		}
		.alert-box {
			background-color: #fff3cd;
			border-left: 4px solid %s;
			padding: 20px;
			margin: 20px 0;
		}
		.danger-box {
			background-color: #ffe6e6;
			border-left: 4px solid #dc3545;
			padding: 20px;
			margin: 20px 0;
		}
		.success-box {
			background-color: #e6ffe6;
			border-left: 4px solid #28a745;
			padding: 20px;
			margin: 20px 0;
		}
		.stat-box {
			background-color: %s;
			padding: 25px;
			text-align: center;
			margin: 20px 0;
			border-radius: 4px;
		}
		.stat-number {
			font-size: 48px;
			font-weight: 900;
			color: %s;
			margin: 0;
		}
		.stat-label {
			font-size: 12px;
			text-transform: uppercase;
			letter-spacing: 2px;
			color: %s;
			margin-top: 5px;
		}
		.footer {
			padding: 25px 30px;
			border-top: 1px solid #eee;
			font-size: 12px;
			color: %s;
			text-align: center;
		}
		.footer-logo {
			font-size: 18px;
			font-weight: 900;
			letter-spacing: 4px;
			color: %s;
			margin-bottom: 10px;
		}
		.divider {
			height: 3px;
			background: linear-gradient(90deg, %s 0%%, %s 50%%, %s 100%%);
			margin: 0;
		}
	</style>
</head>
<body>
	<div class="wrapper">
		<div class="container">
			<div class="header">
				<p class="logo">RISE</p>
				<p class="tagline">Basketball • Performance • Community</p>
			</div>
			<div class="divider"></div>
			<div class="title-bar">
				<h1>%s</h1>
			</div>
			<div class="content">
				%s
			</div>
			<div class="footer">
				<p class="footer-logo">RISE</p>
				<p>This is an automated message from Rise Sports Complex.<br>Please do not reply to this email.</p>
			</div>
		</div>
	</div>
</body>
</html>
`, colorTextDark, colorLightGray, colorWhite, colorBlack, colorYellow, colorYellow,
		colorYellow, colorBlack, colorYellow, colorBlack, colorBlack, colorYellow,
		colorLightGray, colorYellow, colorYellow, colorBlack, colorYellow, colorTextMuted,
		colorTextMuted, colorBlack, colorYellow, colorBlack, colorYellow, title, content)
}

func SignUpConfirmationBody(firstName string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p><strong>Welcome to the team!</strong> Your Rise account is all set up and ready to go.</p>

		<div class="info-box">
			<strong>🏀 GET STARTED:</strong>
			<ol style="margin: 15px 0 0 0; padding-left: 20px;">
				<li>Download our mobile app and log in</li>
				<li>Complete your profile information</li>
				<li>Browse memberships and programs</li>
			</ol>
		</div>

		<p>Time to level up. See you on the court!</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName)
	return baseTemplate("Welcome to Rise", content)
}

func MembershipPurchaseBody(firstName, plan string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p><strong>You're officially in!</strong> Your <strong>%s</strong> membership is now active.</p>

		<div class="success-box">
			<strong>✓ MEMBERSHIP CONFIRMED</strong>
			<p style="margin: 10px 0 0 0;">You now have access to all the benefits included in your plan.</p>
		</div>

		<div class="info-box">
			<strong>WHAT'S NEXT:</strong>
			<ol style="margin: 15px 0 0 0; padding-left: 20px;">
				<li>Check your dashboard for membership details</li>
				<li>Book your first session</li>
				<li>Start training!</li>
			</ol>
		</div>

		<p>Let's get to work.</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, plan)
	return baseTemplate("Membership Confirmed", content)
}

func EmailVerificationBody(firstName, verificationURL string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p>Welcome to Rise! Just one more step to complete your registration.</p>

		<p style="text-align: center; margin: 30px 0;">
			<a href="%s" class="button">VERIFY EMAIL</a>
		</p>

		<p style="font-size: 13px; color: #666;">Or copy and paste this link into your browser:</p>
		<p style="background-color: #f5f5f5; padding: 12px; border-radius: 4px; word-break: break-all; font-size: 12px; font-family: monospace;">
			%s
		</p>

		<div class="alert-box">
			<strong>⏰ HEADS UP:</strong>
			<p style="margin: 10px 0 0 0;">This verification link expires in 24 hours.</p>
		</div>

		<p style="font-size: 13px; color: #666;">If you didn't create an account with Rise, you can safely ignore this email.</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, verificationURL, verificationURL)
	return baseTemplate("Verify Your Email", content)
}

func SubsidyApprovedBody(firstName, providerName string, amount float64, validUntil string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p><strong>Great news!</strong> You've been approved for financial assistance.</p>

		<div class="stat-box">
			<p class="stat-number">$%.2f</p>
			<p class="stat-label">Subsidy Approved</p>
		</div>

		<div class="info-box">
			<strong>📋 DETAILS:</strong>
			<ul style="margin: 15px 0 0 0; padding-left: 20px; list-style: none;">
				<li><strong>Provider:</strong> %s</li>
				<li><strong>Amount:</strong> $%.2f</li>
				<li><strong>Valid Until:</strong> %s</li>
			</ul>
		</div>

		<div class="success-box">
			<strong>HOW IT WORKS:</strong>
			<p style="margin: 10px 0 0 0;">Your subsidy will be automatically applied to your next membership purchase. You'll only pay the remaining balance.</p>
		</div>

		<p>Questions? Don't hesitate to reach out.</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, amount, providerName, amount, validUntil)
	return baseTemplate("Subsidy Approved!", content)
}

func SubsidyUsedBody(firstName string, amountUsed, remainingBalance float64, transactionType string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p>Your subsidy was applied to your recent %s transaction.</p>

		<div style="display: flex; gap: 15px; margin: 20px 0;">
			<div class="stat-box" style="flex: 1; background-color: #e6ffe6;">
				<p style="font-size: 28px; font-weight: 900; color: #28a745; margin: 0;">$%.2f</p>
				<p class="stat-label">Applied</p>
			</div>
			<div class="stat-box" style="flex: 1;">
				<p style="font-size: 28px; font-weight: 900; color: #FFD700; margin: 0;">$%.2f</p>
				<p class="stat-label">Remaining</p>
			</div>
		</div>

		<div class="info-box">
			<strong>💰 BALANCE UPDATE:</strong>
			<p style="margin: 10px 0 0 0;">You still have <strong>$%.2f</strong> available for future memberships.</p>
		</div>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, transactionType, amountUsed, remainingBalance, remainingBalance)
	return baseTemplate("Subsidy Applied", content)
}

func SubsidyDepletedBody(firstName string, totalUsed float64) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p>Your subsidy balance has been fully used.</p>

		<div class="stat-box">
			<p class="stat-number">$%.2f</p>
			<p class="stat-label">Total Used</p>
		</div>

		<div class="alert-box">
			<strong>WHAT'S NEXT:</strong>
			<p style="margin: 10px 0 0 0;">Future membership purchases will be charged at the regular price. If you need assistance with costs, please contact our team to discuss available options.</p>
		</div>

		<p>Thanks for being part of the Rise community!</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, totalUsed)
	return baseTemplate("Subsidy Fully Used", content)
}

func SubsidyExpiringBody(firstName string, remainingBalance float64, expiryDate string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p><strong>Heads up!</strong> Your subsidy is expiring soon.</p>

		<div class="stat-box" style="border: 2px solid #dc3545;">
			<p style="font-size: 28px; font-weight: 900; color: #FFD700; margin: 0;">$%.2f</p>
			<p class="stat-label">Remaining Balance</p>
			<p style="color: #dc3545; font-weight: bold; margin-top: 10px;">Expires: %s</p>
		</div>

		<div class="danger-box">
			<strong>⚠️ ACTION REQUIRED:</strong>
			<p style="margin: 10px 0 0 0;">Purchase or renew your membership before the expiry date to use your remaining balance. After expiration, unused funds will no longer be available.</p>
		</div>

		<p>Don't miss out — use it before you lose it!</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, remainingBalance, expiryDate)
	return baseTemplate("Subsidy Expiring Soon", content)
}

func PaymentFailedBody(firstName, membershipPlan, updatePaymentURL string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p>We couldn't process your payment for your <strong>%s</strong> membership.</p>

		<div class="danger-box">
			<strong>WHAT HAPPENED:</strong>
			<p style="margin: 10px 0 0 0;">Your payment method was declined. This could be due to insufficient funds, an expired card, or your bank blocking the transaction.</p>
		</div>

		<div class="info-box">
			<strong>FIX IT NOW:</strong>
			<ol style="margin: 15px 0 0 0; padding-left: 20px;">
				<li>Check that your payment method has sufficient funds</li>
				<li>Make sure your card hasn't expired</li>
				<li>Update your payment method below</li>
			</ol>
		</div>

		<p style="text-align: center; margin: 30px 0;">
			<a href="%s" class="button">UPDATE PAYMENT</a>
		</p>

		<div class="alert-box">
			<strong>⏰ IMPORTANT:</strong>
			<p style="margin: 10px 0 0 0;">If payment isn't received within 7 days, your membership access may be suspended.</p>
		</div>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, membershipPlan, updatePaymentURL)
	return baseTemplate("Payment Failed", content)
}

func AccountRecoveryBody(resetURL string) string {
	content := fmt.Sprintf(`
		<p>Hey,</p>
		<p>We've made some updates to our system and need you to reset your password to continue accessing your Rise account.</p>

		<div class="success-box">
			<strong>✓ YOUR ACCOUNT IS SAFE</strong>
			<p style="margin: 10px 0 0 0;">All your membership details, credits, and account information are intact. You just need to set a new password.</p>
		</div>

		<p style="text-align: center; margin: 30px 0;">
			<a href="%s" class="button">RESET PASSWORD</a>
		</p>

		<p style="font-size: 13px; color: #666;">Or copy and paste this link into your browser:</p>
		<p style="background-color: #f5f5f5; padding: 12px; border-radius: 4px; word-break: break-all; font-size: 12px; font-family: monospace;">
			%s
		</p>

		<p>Questions? We're here to help.</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, resetURL, resetURL)
	return baseTemplate("Reset Your Password", content)
}

func PaymentFailedReminderBody(firstName, membershipPlan, updatePaymentURL string, daysUntilSuspension int) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p>This is a reminder — we still haven't received payment for your <strong>%s</strong> membership.</p>

		<div class="stat-box" style="border: 3px solid #dc3545;">
			<p class="stat-number" style="color: #dc3545;">%d</p>
			<p class="stat-label">Days Until Suspension</p>
		</div>

		<div class="alert-box">
			<strong>DON'T LOSE ACCESS:</strong>
			<p style="margin: 10px 0 0 0;">Update your payment method now to keep your membership benefits without interruption.</p>
		</div>

		<p style="text-align: center; margin: 30px 0;">
			<a href="%s" class="button" style="background-color: #dc3545;">UPDATE PAYMENT NOW</a>
		</p>

		<p style="font-size: 13px; color: #666;">If you've already updated your payment method, please disregard this email.</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, membershipPlan, daysUntilSuspension, updatePaymentURL)
	return baseTemplate("Payment Reminder", content)
}

func EventNotificationBody(firstName, subject, message string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>

		<div style="background-color: #f5f5f5; border-left: 4px solid #FFD700; padding: 20px; margin: 20px 0; white-space: pre-wrap;">%s</div>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, message)
	return baseTemplate(subject, content)
}

func PaymentRequestBody(firstName string, amount float64, paymentURL string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p>You have an outstanding balance that requires your attention.</p>

		<div class="stat-box">
			<p class="stat-number">$%.2f</p>
			<p class="stat-label">Amount Due</p>
		</div>

		<p>Please complete your payment at your earliest convenience to keep your account in good standing.</p>

		<p style="text-align: center; margin: 30px 0;">
			<a href="%s" class="button">PAY NOW</a>
		</p>

		<p style="font-size: 13px; color: #666;">This is a secure payment link. If you have any questions about this charge, please contact us.</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, amount, paymentURL)
	return baseTemplate("Payment Request", content)
}

func PaymentReceivedBody(firstName string, amount float64) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p>Great news! We've received your payment.</p>

		<div class="success-box">
			<strong>✓ PAYMENT RECEIVED</strong>
			<p style="margin: 10px 0 0 0;">Amount: $%.2f</p>
		</div>

		<p>Thank you for your payment. Your account is now up to date.</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, amount)
	return baseTemplate("Payment Received", content)
}

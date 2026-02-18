package email

import (
	"fmt"
	"log"
)

// SendParentLinkRequestEmail sends a verification code to a parent when a child initiates a link request
func SendParentLinkRequestEmail(to, parentName, childName, code string) {
	body := ParentLinkRequestBody(parentName, childName, code)
	if err := SendEmail(to, "Parent Link Request - Rise", body); err != nil {
		log.Println("failed to send parent link request email:", err.Message)
	} else {
		log.Printf("Parent link request email sent successfully to %s", to)
	}
}

// SendChildLinkRequestEmail sends a verification code to a child when a parent initiates a link request
func SendChildLinkRequestEmail(to, childName, parentName, code string) {
	body := ChildLinkRequestBody(childName, parentName, code)
	if err := SendEmail(to, "Parent Link Request - Rise", body); err != nil {
		log.Println("failed to send child link request email:", err.Message)
	} else {
		log.Printf("Child link request email sent successfully to %s", to)
	}
}

// SendTransferApprovalEmail sends a verification code to the old parent when a transfer is requested
func SendTransferApprovalEmail(to, oldParentName, childName, newParentName, code string) {
	body := TransferApprovalBody(oldParentName, childName, newParentName, code)
	if err := SendEmail(to, "Child Transfer Approval Required - Rise", body); err != nil {
		log.Println("failed to send transfer approval email:", err.Message)
	} else {
		log.Printf("Transfer approval email sent successfully to %s", to)
	}
}

// SendLinkCompleteEmail sends a confirmation to the child when the link is complete
func SendLinkCompleteEmail(to, childName, parentName string) {
	body := LinkCompleteBody(childName, parentName)
	if err := SendEmail(to, "Parent Link Complete - Rise", body); err != nil {
		log.Println("failed to send link complete email:", err.Message)
	} else {
		log.Printf("Link complete email sent successfully to %s", to)
	}
}

// ParentLinkRequestBody creates the email body for a parent link request (sent to parent)
func ParentLinkRequestBody(parentName, childName, code string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p><strong>%s</strong> has requested to link their account to yours as their parent/guardian.</p>

		<div class="stat-box">
			<p class="stat-number" style="letter-spacing: 8px;">%s</p>
			<p class="stat-label">Verification Code</p>
		</div>

		<div class="info-box">
			<strong>WHAT THIS MEANS:</strong>
			<ul style="margin: 15px 0 0 0; padding-left: 20px;">
				<li>You'll be able to view %s's schedule, events, and activities</li>
				<li>You'll receive notifications about their account</li>
				<li>You can manage their memberships and credits</li>
			</ul>
		</div>

		<div class="alert-box">
			<strong>HOW TO CONFIRM:</strong>
			<p style="margin: 10px 0 0 0;">Enter the verification code above in the Rise app to confirm this link request. This code expires in 24 hours.</p>
		</div>

		<p style="font-size: 13px; color: #666;">If you didn't expect this request or don't recognize the user, please ignore this email.</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, parentName, childName, code, childName)
	return baseTemplate("Parent Link Request", content)
}

// ChildLinkRequestBody creates the email body for a link request (sent to child)
func ChildLinkRequestBody(childName, parentName, code string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p><strong>%s</strong> has requested to link to your account as your parent/guardian.</p>

		<div class="stat-box">
			<p class="stat-number" style="letter-spacing: 8px;">%s</p>
			<p class="stat-label">Verification Code</p>
		</div>

		<div class="info-box">
			<strong>WHAT THIS MEANS:</strong>
			<ul style="margin: 15px 0 0 0; padding-left: 20px;">
				<li>%s will be able to view your schedule, events, and activities</li>
				<li>They'll receive copies of notifications sent to you</li>
				<li>They can help manage your memberships and credits</li>
			</ul>
		</div>

		<div class="alert-box">
			<strong>HOW TO CONFIRM:</strong>
			<p style="margin: 10px 0 0 0;">Enter the verification code above in the Rise app to confirm this link. This code expires in 24 hours.</p>
		</div>

		<p style="font-size: 13px; color: #666;">If you didn't expect this request or don't recognize the user, please ignore this email.</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, childName, parentName, code, parentName)
	return baseTemplate("Parent Link Request", content)
}

// TransferApprovalBody creates the email body for transfer approval (sent to old parent)
func TransferApprovalBody(oldParentName, childName, newParentName, code string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p>A request has been made to transfer <strong>%s</strong>'s account to a new parent/guardian: <strong>%s</strong>.</p>

		<div class="stat-box">
			<p class="stat-number" style="letter-spacing: 8px;">%s</p>
			<p class="stat-label">Approval Code</p>
		</div>

		<div class="alert-box">
			<strong>YOUR APPROVAL IS REQUIRED:</strong>
			<p style="margin: 10px 0 0 0;">As %s's current parent/guardian on file, your approval is required to complete this transfer.</p>
		</div>

		<div class="info-box">
			<strong>WHAT HAPPENS IF YOU APPROVE:</strong>
			<ul style="margin: 15px 0 0 0; padding-left: 20px;">
				<li>%s will be linked to %s instead of you</li>
				<li>You will no longer receive notifications for this account</li>
				<li>You will lose access to view their schedule and activities</li>
			</ul>
		</div>

		<p style="font-size: 13px; color: #666;">To approve this transfer, enter the code above in the Rise app. If you did not expect this request or do not approve, simply ignore this email.</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, oldParentName, childName, newParentName, code, childName, childName, newParentName)
	return baseTemplate("Transfer Approval Required", content)
}

// LinkCompleteBody creates the email body for link completion (sent to child)
func LinkCompleteBody(childName, parentName string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p><strong>Great news!</strong> Your account has been successfully linked to <strong>%s</strong>.</p>

		<div class="success-box">
			<strong>LINK COMPLETE</strong>
			<p style="margin: 10px 0 0 0;">Your parent/guardian can now view your schedule, activities, and help manage your account.</p>
		</div>

		<div class="info-box">
			<strong>WHAT'S NEXT:</strong>
			<ul style="margin: 15px 0 0 0; padding-left: 20px;">
				<li>%s will receive copies of important notifications</li>
				<li>They can view your upcoming events and activities</li>
				<li>They can help manage your credits and memberships</li>
			</ul>
		</div>

		<p>If you have any questions about this link, please contact us.</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, childName, parentName, parentName)
	return baseTemplate("Account Linked", content)
}

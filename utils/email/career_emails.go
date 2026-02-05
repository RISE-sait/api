package email

import (
	"fmt"
	"log"
)

func ApplicationReceivedBody(firstName, jobTitle string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p><strong>Thanks for applying!</strong> We've received your application for the <strong>%s</strong> position at Rise.</p>

		<div class="success-box">
			<strong>WHAT'S NEXT:</strong>
			<p style="margin: 10px 0 0 0;">Our team will review your application and get back to you. This typically takes 1-2 weeks.</p>
		</div>

		<div class="info-box">
			<strong>IN THE MEANTIME:</strong>
			<ul style="margin: 15px 0 0 0; padding-left: 20px;">
				<li>Keep an eye on your email for updates</li>
				<li>Feel free to reach out if you have any questions</li>
			</ul>
		</div>

		<p>We appreciate your interest in joining the Rise team!</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, jobTitle)
	return baseTemplate("Application Received", content)
}

func NewApplicationAlertBody(applicantName, applicantEmail, jobTitle string) string {
	content := fmt.Sprintf(`
		<p>A new application has been submitted.</p>

		<div class="info-box">
			<strong>APPLICATION DETAILS:</strong>
			<ul style="margin: 15px 0 0 0; padding-left: 20px; list-style: none;">
				<li><strong>Position:</strong> %s</li>
				<li><strong>Applicant:</strong> %s</li>
				<li><strong>Email:</strong> %s</li>
			</ul>
		</div>

		<p>Please log in to the admin dashboard to review this application.</p>

		<p style="margin-top: 30px;"><strong>— Rise Careers System</strong></p>
	`, jobTitle, applicantName, applicantEmail)
	return baseTemplate("New Job Application", content)
}

func InterviewInvitationBody(firstName, jobTitle string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p><strong>Great news!</strong> We'd like to invite you for an interview for the <strong>%s</strong> position.</p>

		<div class="success-box">
			<strong>NEXT STEPS:</strong>
			<p style="margin: 10px 0 0 0;">A member of our team will reach out shortly to schedule a time that works for you.</p>
		</div>

		<div class="info-box">
			<strong>TIPS FOR YOUR INTERVIEW:</strong>
			<ul style="margin: 15px 0 0 0; padding-left: 20px;">
				<li>Review the job description and requirements</li>
				<li>Prepare examples of your relevant experience</li>
				<li>Have questions ready about the role and team</li>
			</ul>
		</div>

		<p>We're looking forward to meeting you!</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, jobTitle)
	return baseTemplate("Interview Invitation", content)
}

func OfferNotificationBody(firstName, jobTitle string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p><strong>Congratulations!</strong> We're thrilled to extend you an offer for the <strong>%s</strong> position at Rise.</p>

		<div class="success-box">
			<strong>WHAT'S NEXT:</strong>
			<p style="margin: 10px 0 0 0;">A member of our team will be in touch shortly with the details of your offer.</p>
		</div>

		<p>We're excited about the possibility of you joining our team!</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, jobTitle)
	return baseTemplate("Job Offer", content)
}

func RejectionNotificationBody(firstName, jobTitle string) string {
	content := fmt.Sprintf(`
		<p>Hey %s,</p>
		<p>Thank you for your interest in the <strong>%s</strong> position at Rise and for taking the time to apply.</p>

		<p>After careful consideration, we've decided to move forward with other candidates whose experience more closely aligns with our current needs.</p>

		<div class="info-box">
			<strong>KEEP IN TOUCH:</strong>
			<p style="margin: 10px 0 0 0;">We encourage you to check our careers page for future opportunities that may be a great fit.</p>
		</div>

		<p>We wish you all the best in your career journey!</p>

		<p style="margin-top: 30px;"><strong>— The Rise Team</strong></p>
	`, firstName, jobTitle)
	return baseTemplate("Application Update", content)
}

func SendApplicationReceivedEmail(to, firstName, jobTitle string) {
	body := ApplicationReceivedBody(firstName, jobTitle)
	if err := SendEmail(to, "Application Received - Rise", body); err != nil {
		log.Println("failed to send application received email:", err.Message)
	}
}

func SendNewApplicationAlertEmail(to, applicantName, applicantEmail, jobTitle string) {
	body := NewApplicationAlertBody(applicantName, applicantEmail, jobTitle)
	if err := SendEmail(to, "New Application: "+jobTitle+" - Rise", body); err != nil {
		log.Println("failed to send new application alert email:", err.Message)
	}
}

func SendInterviewInvitationEmail(to, firstName, jobTitle string) {
	body := InterviewInvitationBody(firstName, jobTitle)
	if err := SendEmail(to, "Interview Invitation - Rise", body); err != nil {
		log.Println("failed to send interview invitation email:", err.Message)
	}
}

func SendOfferNotificationEmail(to, firstName, jobTitle string) {
	body := OfferNotificationBody(firstName, jobTitle)
	if err := SendEmail(to, "Job Offer - Rise", body); err != nil {
		log.Println("failed to send offer notification email:", err.Message)
	}
}

func SendRejectionNotificationEmail(to, firstName, jobTitle string) {
	body := RejectionNotificationBody(firstName, jobTitle)
	if err := SendEmail(to, "Application Update - Rise", body); err != nil {
		log.Println("failed to send rejection notification email:", err.Message)
	}
}

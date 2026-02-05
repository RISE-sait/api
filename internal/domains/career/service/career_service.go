package service

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	errLib "api/internal/libs/errors"
	"api/internal/services/gcp"
	"api/utils/email"
)

const (
	maxResumeSize = 5 << 20 // 5MB
	adminEmail    = "info@risesportscomplex.com"
)

var allowedResumeTypes = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
}

func UploadResume(file io.Reader, fileName string, fileSize int64) (string, *errLib.CommonError) {
	if fileSize > maxResumeSize {
		return "", errLib.New("Resume file size must be under 5MB", http.StatusBadRequest)
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if !allowedResumeTypes[ext] {
		return "", errLib.New("Resume must be a PDF, DOC, or DOCX file", http.StatusBadRequest)
	}

	gcpFileName := fmt.Sprintf("resumes/%d_%s", time.Now().UnixNano(), fileName)

	url, err := gcp.UploadImageToGCP(file, gcpFileName)
	if err != nil {
		return "", err
	}

	return url, nil
}

func SendApplicationEmails(applicantEmail, applicantFirstName, applicantLastName, jobTitle string) {
	applicantName := applicantFirstName + " " + applicantLastName

	go func() {
		log.Printf("Sending application confirmation to %s", applicantEmail)
		email.SendApplicationReceivedEmail(applicantEmail, applicantFirstName, jobTitle)
	}()

	go func() {
		log.Printf("Sending new application alert to admin for %s", jobTitle)
		email.SendNewApplicationAlertEmail(adminEmail, applicantName, applicantEmail, jobTitle)
	}()
}

func SendStatusChangeEmail(applicantEmail, applicantFirstName, jobTitle, newStatus string) {
	go func() {
		switch newStatus {
		case "interview":
			log.Printf("Sending interview invitation to %s", applicantEmail)
			email.SendInterviewInvitationEmail(applicantEmail, applicantFirstName, jobTitle)
		case "offer":
			log.Printf("Sending offer notification to %s", applicantEmail)
			email.SendOfferNotificationEmail(applicantEmail, applicantFirstName, jobTitle)
		case "rejected":
			log.Printf("Sending rejection notification to %s", applicantEmail)
			email.SendRejectionNotificationEmail(applicantEmail, applicantFirstName, jobTitle)
		}
	}()
}

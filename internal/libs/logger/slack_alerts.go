package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// SlackPayload represents a Slack webhook message
type SlackPayload struct {
	Text        string       `json:"text"`
	Channel     string       `json:"channel,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color     string  `json:"color"`
	Title     string  `json:"title"`
	Text      string  `json:"text"`
	Timestamp int64   `json:"ts"`
	Fields    []Field `json:"fields,omitempty"`
}

// Field represents a field in a Slack attachment
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// SendSlackAlert sends critical payment alerts to Slack
func SendSlackAlert(alertType string, message string, fields map[string]interface{}) {
	SendSlackAlertAsync(alertType, message, fields)
}

// SendSlackAlertSync sends critical payment alerts to Slack synchronously (blocks)
func SendSlackAlertSync(alertType string, message string, fields map[string]interface{}) {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		fmt.Println("ERROR: SLACK_WEBHOOK_URL not found in environment")
		return
	}
	
	fmt.Printf("URGENT: Sending sync Slack alert: %s - %s\n", alertType, message)
	
	// Build the same payload
	var color string
	var emoji string
	
	switch alertType {
	case "PAYMENT_DOWN":
		color = "danger"
		emoji = "[CRITICAL]"
	case "WEBHOOK_FAILED":
		color = "warning" 
		emoji = "[WARNING]"
	case "DB_CONNECTION":
		color = "danger"
		emoji = "[ERROR]"
	default:
		color = "warning"
		emoji = "[ALERT]"
	}

	var slackFields []Field
	for key, value := range fields {
		slackFields = append(slackFields, Field{
			Title: key,
			Value: fmt.Sprintf("%v", value),
			Short: true,
		})
	}

	payload := SlackPayload{
		Channel:   "#payment-alerts",
		Username:  "Payment-Bot",
		IconEmoji: ":warning:",
		Attachments: []Attachment{
			{
				Color:     color,
				Title:     fmt.Sprintf("%s %s Alert", emoji, alertType),
				Text:      message,
				Timestamp: time.Now().Unix(),
				Fields:    slackFields,
			},
		},
	}

	// Send synchronously for critical alerts
	sendSlackWebhookFast(webhookURL, payload)
}

// SendSlackAlertAsync sends critical payment alerts to Slack asynchronously
func SendSlackAlertAsync(alertType string, message string, fields map[string]interface{}) {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		fmt.Println("ERROR: SLACK_WEBHOOK_URL not found in environment")
		return // Slack alerts not configured
	}
	
	fmt.Printf("Sending Slack alert: %s - %s\n", alertType, message)

	// Determine alert severity and color
	var color string
	var emoji string
	
	switch alertType {
	case "PAYMENT_DOWN":
		color = "danger"
		emoji = "[CRITICAL]"
	case "WEBHOOK_FAILED":
		color = "warning" 
		emoji = "[WARNING]"
	case "DB_CONNECTION":
		color = "danger"
		emoji = "[ERROR]"
	default:
		color = "warning"
		emoji = "[ALERT]"
	}

	// Build field list from provided fields
	var slackFields []Field
	for key, value := range fields {
		slackFields = append(slackFields, Field{
			Title: key,
			Value: fmt.Sprintf("%v", value),
			Short: true,
		})
	}

	payload := SlackPayload{
		Channel:   "#payment-alerts",
		Username:  "Payment-Bot",
		IconEmoji: ":warning:",
		Attachments: []Attachment{
			{
				Color:     color,
				Title:     fmt.Sprintf("%s %s Alert", emoji, alertType),
				Text:      message,
				Timestamp: time.Now().Unix(),
				Fields:    slackFields,
			},
		},
	}

	// Send to Slack immediately with shorter timeout
	go sendSlackWebhookFast(webhookURL, payload)
}

// sendSlackWebhookFast sends the payload to Slack webhook with optimized settings
func sendSlackWebhookFast(webhookURL string, payload SlackPayload) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("ERROR: Failed to marshal Slack payload: %v\n", err)
		return
	}

	// Use shorter timeout and better HTTP client settings for faster delivery
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	
	fmt.Printf("Sending Slack webhook to %s...\n", webhookURL[:50]+"...")
	start := time.Now()
	
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("ERROR: Slack webhook failed after %v: %v\n", time.Since(start), err)
		return
	}
	defer resp.Body.Close()
	
	fmt.Printf("SUCCESS: Slack webhook sent successfully in %v (status: %d)\n", time.Since(start), resp.StatusCode)
}

// IsPaymentCritical determines if a log entry should trigger Slack alerts
func IsPaymentCritical(level LogLevel, message string, component string) bool {
	// Only alert on errors and fatals
	if level != LevelError && level != LevelFatal {
		return false
	}

	// Payment-related components
	paymentComponents := []string{
		"stripe-webhooks",
		"payment-security", 
		"pci-compliance",
		"checkout-service",
	}

	// Check if component is payment-related
	for _, pc := range paymentComponents {
		if component == pc {
			return true
		}
	}

	// Check if message contains payment-related keywords
	criticalKeywords := []string{
		"payment processing failed",
		"webhook processing failed", 
		"database connection",
		"stripe api error",
		"checkout failed",
		"subscription failed",
	}

	lowerMessage := strings.ToLower(message)
	for _, keyword := range criticalKeywords {
		if strings.Contains(lowerMessage, keyword) {
			return true
		}
	}

	return false
}
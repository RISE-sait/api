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

// WebhookAlertDetails contains comprehensive webhook failure information
type WebhookAlertDetails struct {
	EventID         string
	EventType       string 
	SessionID       string
	CustomerID      string
	StripePriceID   string
	PlanID          string
	UserEmail       string
	ErrorType       string
	ErrorMessage    string
	FailureStep     string
	RetryAttempt    int
	SessionStatus   string
	PaymentStatus   string
	SubscriptionID  string
	TroubleshootingSteps []string
}

// SendWebhookFailureAlert sends comprehensive webhook failure alerts
func SendWebhookFailureAlert(details WebhookAlertDetails) {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		fmt.Println("ERROR: SLACK_WEBHOOK_URL not found in environment")
		return
	}

	// Determine error severity and emoji
	var color string
	var emoji string
	var priority string
	
	switch details.ErrorType {
	case "DUPLICATE_ENROLLMENT", "IDEMPOTENCY":
		color = "warning"
		emoji = "⚠️"
		priority = "LOW"
	case "FOREIGN_KEY_VIOLATION", "DATA_INTEGRITY":
		color = "#ff9500" // Orange
		emoji = "🔍" 
		priority = "MEDIUM"
	case "STRIPE_API_FAILURE", "NETWORK_ERROR":
		color = "danger"
		emoji = "🚨"
		priority = "HIGH"
	case "CRITICAL_SYSTEM_FAILURE":
		color = "danger" 
		emoji = "💥"
		priority = "CRITICAL"
	default:
		color = "warning"
		emoji = "❗"
		priority = "MEDIUM"
	}

	// Build the comprehensive alert
	attachment := Attachment{
		Color: color,
		Title: fmt.Sprintf("%s WEBHOOK FAILURE - %s Priority", emoji, priority),
		Text:  fmt.Sprintf("```%s```", details.ErrorMessage),
		Timestamp: time.Now().Unix(),
		Fields: []Field{
			{Title: "Event ID", Value: details.EventID, Short: true},
			{Title: "Event Type", Value: details.EventType, Short: true},
			{Title: "Session ID", Value: details.SessionID, Short: false},
			{Title: "Customer", Value: fmt.Sprintf("%s (%s)", details.CustomerID, details.UserEmail), Short: false},
			{Title: "Failure Point", Value: details.FailureStep, Short: true},
			{Title: "Error Category", Value: details.ErrorType, Short: true},
		},
	}

	// Add context fields if available
	if details.StripePriceID != "" {
		attachment.Fields = append(attachment.Fields, Field{
			Title: "Stripe Price ID", Value: details.StripePriceID, Short: true,
		})
	}
	if details.PlanID != "" {
		attachment.Fields = append(attachment.Fields, Field{
			Title: "Membership Plan", Value: details.PlanID, Short: true,
		})
	}
	if details.SubscriptionID != "" {
		attachment.Fields = append(attachment.Fields, Field{
			Title: "Subscription ID", Value: details.SubscriptionID, Short: true,
		})
	}
	if details.RetryAttempt > 1 {
		attachment.Fields = append(attachment.Fields, Field{
			Title: "Retry Attempt", Value: fmt.Sprintf("%d", details.RetryAttempt), Short: true,
		})
	}

	// Add troubleshooting steps
	if len(details.TroubleshootingSteps) > 0 {
		troubleshootText := "🔧 **Troubleshooting Steps:**\n"
		for i, step := range details.TroubleshootingSteps {
			troubleshootText += fmt.Sprintf("%d. %s\n", i+1, step)
		}
		attachment.Fields = append(attachment.Fields, Field{
			Title: "Next Actions", Value: troubleshootText, Short: false,
		})
	}

	// Add quick investigation links
	investigationLinks := fmt.Sprintf(
		"🔗 **Quick Links:**\n"+
		"• [Stripe Event](%s)\n"+
		"• [Stripe Session](%s)\n"+
		"• [Database Query](%s)",
		fmt.Sprintf("https://dashboard.stripe.com/test/events/%s", details.EventID),
		fmt.Sprintf("https://dashboard.stripe.com/test/checkout/sessions/%s", details.SessionID),
		fmt.Sprintf("#webhook-debug-customer-%s", details.CustomerID[:8]),
	)
	
	attachment.Fields = append(attachment.Fields, Field{
		Title: "Investigation", Value: investigationLinks, Short: false,
	})

	payload := SlackPayload{
		Channel:   "#webhook-alerts", // Dedicated webhook channel
		Username:  "Webhook-Debug-Bot",
		IconEmoji: ":robot_face:",
		Text:      fmt.Sprintf("*Webhook Processing Failed* - %s Priority", priority),
		Attachments: []Attachment{attachment},
	}

	// Send with appropriate urgency
	if priority == "CRITICAL" || priority == "HIGH" {
		sendSlackWebhookFast(webhookURL, payload)
	} else {
		go sendSlackWebhookFast(webhookURL, payload)
	}
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

// ClassifyWebhookError analyzes error messages and provides troubleshooting steps
func ClassifyWebhookError(errorMessage string) (errorType string, troubleshootingSteps []string) {
	lowerError := strings.ToLower(errorMessage)
	
	// Classify error types and provide specific troubleshooting
	if strings.Contains(lowerError, "customer is already enrolled") ||
	   strings.Contains(lowerError, "unique constraint") ||
	   strings.Contains(lowerError, "duplicate") {
		return "DUPLICATE_ENROLLMENT", []string{
			"Check if webhook idempotency is working correctly",
			"Verify customer enrollment status in database", 
			"Review webhook retry logic and event deduplication",
			"Check if Stripe is sending duplicate events",
		}
	}
	
	if strings.Contains(lowerError, "foreign key violation") ||
	   strings.Contains(lowerError, "invalid customer or membership plan id") {
		return "FOREIGN_KEY_VIOLATION", []string{
			"Verify customer exists in users.users table",
			"Check if membership plan exists for the Stripe price ID",
			"Validate UUID format in webhook metadata",
			"Review data sync between Stripe and database",
		}
	}
	
	if strings.Contains(lowerError, "stripe") && 
	   (strings.Contains(lowerError, "timeout") || 
	    strings.Contains(lowerError, "network") ||
	    strings.Contains(lowerError, "connection") ||
	    strings.Contains(lowerError, "api")) {
		return "STRIPE_API_FAILURE", []string{
			"Check Stripe API status at https://status.stripe.com",
			"Review network connectivity and firewall settings",
			"Verify Stripe API key is valid and has correct permissions",
			"Check if rate limits are being exceeded",
			"Consider implementing exponential backoff retry",
		}
	}
	
	if strings.Contains(lowerError, "database") &&
	   (strings.Contains(lowerError, "connection") ||
	    strings.Contains(lowerError, "timeout") ||
	    strings.Contains(lowerError, "unavailable")) {
		return "DATABASE_CONNECTION", []string{
			"Check database connection pool status",
			"Verify database server is running and accessible",
			"Review connection string and credentials",
			"Check for database locks or deadlocks",
			"Monitor database CPU and memory usage",
		}
	}
	
	if strings.Contains(lowerError, "failed to set cancel date") ||
	   strings.Contains(lowerError, "subscription update") {
		return "SUBSCRIPTION_UPDATE_FAILURE", []string{
			"Check if subscription exists in Stripe",
			"Verify subscription is in updatable state",
			"Review Stripe webhook retry configuration",
			"Consider making subscription updates non-blocking",
		}
	}
	
	// Default classification
	return "UNKNOWN_ERROR", []string{
		"Review full error logs and stack trace",
		"Check recent code deployments or configuration changes",
		"Verify all external service dependencies",
		"Contact development team for detailed investigation",
	}
}

// GetUserEmailFromID retrieves user email for better alert context
func GetUserEmailFromID(customerID string) string {
	// This would ideally query the database, but for now return placeholder
	// In production, you'd want to implement a quick lookup
	return fmt.Sprintf("customer-%s@unknown.com", customerID[:8])
}
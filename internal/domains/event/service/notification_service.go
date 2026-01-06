package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"api/internal/di"
	dto "api/internal/domains/event/dto"
	errLib "api/internal/libs/errors"
	"api/utils/email"

	"github.com/google/uuid"
)

// EventNotificationService handles sending notifications to event attendees
type EventNotificationService struct {
	db *sql.DB
}

// NewEventNotificationService creates a new EventNotificationService
func NewEventNotificationService(container *di.Container) *EventNotificationService {
	return &EventNotificationService{
		db: container.DB,
	}
}

// EventCustomer represents a customer enrolled in an event
type EventCustomer struct {
	ID           uuid.UUID
	FirstName    string
	LastName     string
	Email        string
	HasPushToken bool
}

// EventCustomerEmail represents customer data for email sending
type EventCustomerEmail struct {
	ID        uuid.UUID
	FirstName string
	Email     string
}

// EventCustomerPushToken represents customer push token data
type EventCustomerPushToken struct {
	UserID         uuid.UUID
	FirstName      string
	ExpoPushToken  string
	DeviceType     string
}

// NotificationResult tracks the results of sending notifications
type NotificationResult struct {
	NotificationID uuid.UUID
	RecipientCount int
	EmailSent      int
	EmailFailed    int
	PushSent       int
	PushFailed     int
}

// GetEventCustomers retrieves all enrolled customers for an event
func (s *EventNotificationService) GetEventCustomers(ctx context.Context, eventID uuid.UUID) ([]EventCustomer, *errLib.CommonError) {
	query := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.email,
			EXISTS(SELECT 1 FROM notifications.push_tokens pt WHERE pt.user_id = u.id) AS has_push_token
		FROM events.customer_enrollment ce
		JOIN users.users u ON ce.customer_id = u.id
		WHERE ce.event_id = $1
		  AND ce.is_cancelled = false
		  AND ce.payment_status = 'paid'
		ORDER BY u.last_name, u.first_name
	`

	rows, err := s.db.QueryContext(ctx, query, eventID)
	if err != nil {
		log.Printf("[EVENT-NOTIFICATION] Error querying enrolled customers: %v", err)
		return nil, errLib.New("Failed to get enrolled customers", http.StatusInternalServerError)
	}
	defer rows.Close()

	var customers []EventCustomer
	for rows.Next() {
		var c EventCustomer
		var emailNull sql.NullString
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &emailNull, &c.HasPushToken); err != nil {
			log.Printf("[EVENT-NOTIFICATION] Error scanning customer row: %v", err)
			continue
		}
		if emailNull.Valid {
			c.Email = emailNull.String
		}
		customers = append(customers, c)
	}

	return customers, nil
}

// GetEventCustomersByIDs retrieves specific enrolled customers by their IDs
func (s *EventNotificationService) GetEventCustomersByIDs(ctx context.Context, eventID uuid.UUID, customerIDs []uuid.UUID) ([]EventCustomer, *errLib.CommonError) {
	query := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.email,
			EXISTS(SELECT 1 FROM notifications.push_tokens pt WHERE pt.user_id = u.id) AS has_push_token
		FROM events.customer_enrollment ce
		JOIN users.users u ON ce.customer_id = u.id
		WHERE ce.event_id = $1
		  AND ce.is_cancelled = false
		  AND ce.payment_status = 'paid'
		  AND u.id = ANY($2)
		ORDER BY u.last_name, u.first_name
	`

	rows, err := s.db.QueryContext(ctx, query, eventID, customerIDs)
	if err != nil {
		log.Printf("[EVENT-NOTIFICATION] Error querying specific customers: %v", err)
		return nil, errLib.New("Failed to get customers", http.StatusInternalServerError)
	}
	defer rows.Close()

	var customers []EventCustomer
	for rows.Next() {
		var c EventCustomer
		var emailNull sql.NullString
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &emailNull, &c.HasPushToken); err != nil {
			log.Printf("[EVENT-NOTIFICATION] Error scanning customer row: %v", err)
			continue
		}
		if emailNull.Valid {
			c.Email = emailNull.String
		}
		customers = append(customers, c)
	}

	return customers, nil
}

// GetEventEnrollmentCount returns the count of enrolled customers for an event
func (s *EventNotificationService) GetEventEnrollmentCount(ctx context.Context, eventID uuid.UUID) (int, *errLib.CommonError) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM events.customer_enrollment
		WHERE event_id = $1 AND is_cancelled = false AND payment_status = 'paid'
	`, eventID).Scan(&count)

	if err != nil {
		return 0, errLib.New("Failed to get enrollment count", http.StatusInternalServerError)
	}
	return count, nil
}

// CheckCoachHasAccessToEvent verifies if a coach has access to the event
func (s *EventNotificationService) CheckCoachHasAccessToEvent(ctx context.Context, eventID, staffID uuid.UUID) (bool, *errLib.CommonError) {
	var hasAccess bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM events.events e
			LEFT JOIN athletic.teams t ON e.team_id = t.id
			LEFT JOIN events.staff es ON es.event_id = e.id
			WHERE e.id = $1
			  AND (
				t.coach_id = $2
				OR es.staff_id = $2
			  )
		)
	`, eventID, staffID).Scan(&hasAccess)

	if err != nil {
		return false, errLib.New("Failed to check coach access", http.StatusInternalServerError)
	}
	return hasAccess, nil
}

// GetEventDetails retrieves event details for including in notifications
func (s *EventNotificationService) GetEventDetails(ctx context.Context, eventID uuid.UUID) (name string, startAt time.Time, locationName string, err *errLib.CommonError) {
	query := `
		SELECT
			COALESCE(p.name, t.name, 'Event') AS event_name,
			e.start_at,
			COALESCE(l.name, '') AS location_name
		FROM events.events e
		LEFT JOIN program.programs p ON e.program_id = p.id
		LEFT JOIN athletic.teams t ON e.team_id = t.id
		LEFT JOIN location.locations l ON e.location_id = l.id
		WHERE e.id = $1
	`

	var eventName, locName string
	var start time.Time
	dbErr := s.db.QueryRowContext(ctx, query, eventID).Scan(&eventName, &start, &locName)
	if dbErr != nil {
		if dbErr == sql.ErrNoRows {
			return "", time.Time{}, "", errLib.New("Event not found", http.StatusNotFound)
		}
		return "", time.Time{}, "", errLib.New("Failed to get event details", http.StatusInternalServerError)
	}

	return eventName, start, locName, nil
}

// SendNotification sends notifications to event attendees
func (s *EventNotificationService) SendNotification(
	ctx context.Context,
	eventID uuid.UUID,
	senderID uuid.UUID,
	request dto.SendNotificationRequestDto,
) (*NotificationResult, *errLib.CommonError) {
	// Get target customers
	var customers []EventCustomer
	var err *errLib.CommonError

	if len(request.CustomerIDs) > 0 {
		customers, err = s.GetEventCustomersByIDs(ctx, eventID, request.CustomerIDs)
	} else {
		customers, err = s.GetEventCustomers(ctx, eventID)
	}

	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, errLib.New("No customers found to notify", http.StatusBadRequest)
	}

	result := &NotificationResult{
		RecipientCount: len(customers),
	}

	// Get event details if needed
	var eventDetails string
	if request.IncludeEventDetails {
		eventName, startAt, locationName, _ := s.GetEventDetails(ctx, eventID)
		if eventName != "" {
			eventDetails = fmt.Sprintf("\n\nEvent: %s\nDate: %s\nLocation: %s",
				eventName,
				startAt.Format("Monday, January 2, 2006 at 3:04 PM"),
				locationName,
			)
		}
	}

	message := request.Message
	if eventDetails != "" {
		message += eventDetails
	}

	// Send emails if channel is email or both
	if request.Channel == string(dto.ChannelEmail) || request.Channel == string(dto.ChannelBoth) {
		emailSuccess, emailFailed := s.sendEmails(customers, request.Subject, message)
		result.EmailSent = emailSuccess
		result.EmailFailed = emailFailed
	}

	// Send push notifications if channel is push or both
	if request.Channel == string(dto.ChannelPush) || request.Channel == string(dto.ChannelBoth) {
		pushSuccess, pushFailed := s.sendPushNotifications(ctx, eventID, request.Subject, request.Message)
		result.PushSent = pushSuccess
		result.PushFailed = pushFailed
	}

	// Record notification in history
	notificationID, recordErr := s.recordNotificationHistory(ctx, eventID, senderID, request, result)
	if recordErr != nil {
		log.Printf("[EVENT-NOTIFICATION] Failed to record notification history: %v", recordErr)
		// Don't fail the whole operation, just log
	}
	result.NotificationID = notificationID

	log.Printf("[EVENT-NOTIFICATION] Notification sent for event %s: recipients=%d, email_sent=%d, email_failed=%d, push_sent=%d, push_failed=%d",
		eventID, result.RecipientCount, result.EmailSent, result.EmailFailed, result.PushSent, result.PushFailed)

	return result, nil
}

// sendEmails sends emails to customers
func (s *EventNotificationService) sendEmails(customers []EventCustomer, subject, message string) (success int, failed int) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, customer := range customers {
		if customer.Email == "" {
			mu.Lock()
			failed++
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(c EventCustomer) {
			defer wg.Done()

			body := email.EventNotificationBody(c.FirstName, subject, message)
			if err := email.SendEmail(c.Email, subject, body); err != nil {
				log.Printf("[EVENT-NOTIFICATION] Failed to send email to %s: %v", c.Email, err.Message)
				mu.Lock()
				failed++
				mu.Unlock()
			} else {
				mu.Lock()
				success++
				mu.Unlock()
			}
		}(customer)
	}

	wg.Wait()
	return success, failed
}

// sendPushNotifications sends push notifications to customers enrolled in the event
func (s *EventNotificationService) sendPushNotifications(ctx context.Context, eventID uuid.UUID, title, body string) (success int, failed int) {
	// Get push tokens for enrolled customers
	query := `
		SELECT
			pt.expo_push_token
		FROM events.customer_enrollment ce
		JOIN notifications.push_tokens pt ON pt.user_id = ce.customer_id
		WHERE ce.event_id = $1
		  AND ce.is_cancelled = false
		  AND ce.payment_status = 'paid'
	`

	rows, err := s.db.QueryContext(ctx, query, eventID)
	if err != nil {
		log.Printf("[EVENT-NOTIFICATION] Error getting push tokens: %v", err)
		return 0, 0
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			continue
		}
		tokens = append(tokens, token)
	}

	if len(tokens) == 0 {
		return 0, 0
	}

	// Prepare Expo messages
	type ExpoMessage struct {
		To    string `json:"to"`
		Title string `json:"title"`
		Body  string `json:"body"`
		Sound string `json:"sound"`
	}

	messages := make([]ExpoMessage, len(tokens))
	for i, token := range tokens {
		messages[i] = ExpoMessage{
			To:    token,
			Title: title,
			Body:  body,
			Sound: "default",
		}
	}

	// Send to Expo
	jsonData, err := json.Marshal(messages)
	if err != nil {
		log.Printf("[EVENT-NOTIFICATION] Failed to marshal push messages: %v", err)
		return 0, len(tokens)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "https://exp.host/--/api/v2/push/send", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[EVENT-NOTIFICATION] Failed to create Expo request: %v", err)
		return 0, len(tokens)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[EVENT-NOTIFICATION] Failed to send to Expo: %v", err)
		return 0, len(tokens)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[EVENT-NOTIFICATION] Expo returned status %d", resp.StatusCode)
		return 0, len(tokens)
	}

	// Parse response to count successes/failures
	type ExpoResult struct {
		Status string `json:"status"`
	}
	type ExpoResponse struct {
		Data []ExpoResult `json:"data"`
	}

	var expoResp ExpoResponse
	if err := json.NewDecoder(resp.Body).Decode(&expoResp); err != nil {
		log.Printf("[EVENT-NOTIFICATION] Failed to decode Expo response: %v", err)
		return len(tokens), 0 // Assume success if we got 200
	}

	for _, result := range expoResp.Data {
		if result.Status == "ok" {
			success++
		} else {
			failed++
		}
	}

	return success, failed
}

// recordNotificationHistory saves the notification to history
func (s *EventNotificationService) recordNotificationHistory(
	ctx context.Context,
	eventID uuid.UUID,
	senderID uuid.UUID,
	request dto.SendNotificationRequestDto,
	result *NotificationResult,
) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO events.notification_history (
			event_id, sent_by, channel, subject, message, include_event_details,
			recipient_count, email_success_count, email_failure_count, push_success_count, push_failure_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, eventID, senderID, request.Channel, request.Subject, request.Message, request.IncludeEventDetails,
		result.RecipientCount, result.EmailSent, result.EmailFailed, result.PushSent, result.PushFailed).Scan(&id)

	return id, err
}

// GetNotificationHistory retrieves notification history for an event
func (s *EventNotificationService) GetNotificationHistory(ctx context.Context, eventID uuid.UUID) ([]dto.NotificationHistoryDto, *errLib.CommonError) {
	query := `
		SELECT
			nh.id,
			nh.event_id,
			nh.sent_by,
			u.first_name || ' ' || u.last_name AS sent_by_name,
			nh.channel,
			nh.subject,
			nh.message,
			nh.include_event_details,
			nh.recipient_count,
			nh.email_success_count,
			nh.email_failure_count,
			nh.push_success_count,
			nh.push_failure_count,
			nh.created_at
		FROM events.notification_history nh
		JOIN users.users u ON nh.sent_by = u.id
		WHERE nh.event_id = $1
		ORDER BY nh.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, errLib.New("Failed to get notification history", http.StatusInternalServerError)
	}
	defer rows.Close()

	var history []dto.NotificationHistoryDto
	for rows.Next() {
		var h dto.NotificationHistoryDto
		var subject sql.NullString
		var createdAt time.Time

		if err := rows.Scan(
			&h.ID, &h.EventID, &h.SentBy, &h.SentByName, &h.Channel,
			&subject, &h.Message, &h.IncludeEventDetails, &h.RecipientCount,
			&h.EmailSuccessCount, &h.EmailFailureCount, &h.PushSuccessCount, &h.PushFailureCount,
			&createdAt,
		); err != nil {
			log.Printf("[EVENT-NOTIFICATION] Error scanning history row: %v", err)
			continue
		}

		if subject.Valid {
			h.Subject = subject.String
		}
		h.CreatedAt = createdAt.Format(time.RFC3339)
		history = append(history, h)
	}

	return history, nil
}

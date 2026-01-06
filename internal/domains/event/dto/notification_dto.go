package event

import (
	"net/http"

	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"

	"github.com/google/uuid"
)

// NotificationChannel represents the notification delivery channel
type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
	ChannelPush  NotificationChannel = "push"
	ChannelBoth  NotificationChannel = "both"
)

// SendNotificationRequestDto is the request body for sending notifications to event attendees
type SendNotificationRequestDto struct {
	Channel             string      `json:"channel" validate:"required,oneof=email push both" example:"both"`
	Subject             string      `json:"subject" example:"Event Update"`
	Message             string      `json:"message" validate:"required" example:"Important update about your upcoming event..."`
	IncludeEventDetails bool        `json:"include_event_details" example:"true"`
	CustomerIDs         []uuid.UUID `json:"customer_ids" example:"[\"f0e21457-75d4-4de6-b765-5ee13221fd72\"]"` // nil = all enrolled
}

// Validate validates the notification request
func (dto *SendNotificationRequestDto) Validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}

	// Subject is required for email
	if (dto.Channel == string(ChannelEmail) || dto.Channel == string(ChannelBoth)) && dto.Subject == "" {
		return errLib.New("subject is required for email notifications", http.StatusBadRequest)
	}

	return nil
}

// SendNotificationResponseDto is the response after sending notifications
type SendNotificationResponseDto struct {
	NotificationID uuid.UUID `json:"notification_id"`
	RecipientCount int       `json:"recipient_count"`
	EmailSent      int       `json:"email_sent"`
	EmailFailed    int       `json:"email_failed"`
	PushSent       int       `json:"push_sent"`
	PushFailed     int       `json:"push_failed"`
}

// NotificationHistoryDto represents a notification in the history
type NotificationHistoryDto struct {
	ID                  uuid.UUID `json:"id"`
	EventID             uuid.UUID `json:"event_id"`
	SentBy              uuid.UUID `json:"sent_by"`
	SentByName          string    `json:"sent_by_name"`
	Channel             string    `json:"channel"`
	Subject             string    `json:"subject,omitempty"`
	Message             string    `json:"message"`
	IncludeEventDetails bool      `json:"include_event_details"`
	RecipientCount      int       `json:"recipient_count"`
	EmailSuccessCount   int       `json:"email_success_count"`
	EmailFailureCount   int       `json:"email_failure_count"`
	PushSuccessCount    int       `json:"push_success_count"`
	PushFailureCount    int       `json:"push_failure_count"`
	CreatedAt           string    `json:"created_at"`
}

// EventCustomerDto represents an enrolled customer for an event
type EventCustomerDto struct {
	ID           uuid.UUID `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	HasPushToken bool      `json:"has_push_token"`
}

// EventCustomersResponseDto is the response for getting enrolled customers
type EventCustomersResponseDto struct {
	TotalCount int                `json:"total_count"`
	Customers  []EventCustomerDto `json:"customers"`
}

// NotificationHistoryResponseDto is the response for getting notification history
type NotificationHistoryResponseDto struct {
	Notifications []NotificationHistoryDto `json:"notifications"`
}

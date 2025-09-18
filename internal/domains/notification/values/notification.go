package notification

import "github.com/google/uuid"

type PushToken struct {
	ID            int       `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	ExpoPushToken string    `json:"expo_push_token"`
	DeviceType    string    `json:"device_type"`
}

type TeamNotification struct {
	Type   string                 `json:"type"`
	Title  string                 `json:"title"`
	Body   string                 `json:"body"`
	TeamID uuid.UUID              `json:"team_id"`
	Data   map[string]interface{} `json:"data"`
}
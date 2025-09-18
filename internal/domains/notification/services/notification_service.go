package notification

import (
	"api/internal/di"
	"api/internal/domains/notification/persistence/repositories"
	values "api/internal/domains/notification/values"
	errLib "api/internal/libs/errors"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type NotificationService struct {
	repo *repositories.PushTokenRepository
}

func NewNotificationService(container *di.Container) *NotificationService {
	return &NotificationService{
		repo: repositories.NewPushTokenRepository(container),
	}
}

func (s *NotificationService) RegisterPushToken(ctx context.Context, userID uuid.UUID, token, deviceType string) *errLib.CommonError {
	return s.repo.UpsertPushToken(ctx, userID, token, deviceType)
}

func (s *NotificationService) SendTeamNotification(ctx context.Context, teamID uuid.UUID, notification values.TeamNotification) *errLib.CommonError {
	// Get all push tokens for team members
	tokens, err := s.repo.GetPushTokensByTeamID(ctx, teamID)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil // No tokens to send to
	}

	// Prepare Expo push messages
	messages := make([]ExpoMessage, 0, len(tokens))
	for _, token := range tokens {
		messages = append(messages, ExpoMessage{
			To:    token.ExpoPushToken,
			Title: notification.Title,
			Body:  notification.Body,
			Data:  notification.Data,
			Sound: "default",
		})
	}

	// Send to Expo
	return s.sendToExpo(messages)
}

type ExpoMessage struct {
	To    string                 `json:"to"`
	Title string                 `json:"title"`
	Body  string                 `json:"body"`
	Data  map[string]interface{} `json:"data,omitempty"`
	Sound string                 `json:"sound"`
}

type ExpoResponse struct {
	Data []ExpoResult `json:"data"`
}

type ExpoResult struct {
	Status string `json:"status"`
	ID     string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

func (s *NotificationService) sendToExpo(messages []ExpoMessage) *errLib.CommonError {
	if len(messages) == 0 {
		return nil
	}

	jsonData, err := json.Marshal(messages)
	if err != nil {
		return errLib.New("Failed to marshal notification data", http.StatusInternalServerError)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "https://exp.host/--/api/v2/push/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return errLib.New("Failed to create Expo request", http.StatusInternalServerError)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return errLib.New(fmt.Sprintf("Failed to send notification to Expo: %v", err), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errLib.New(fmt.Sprintf("Expo API returned status %d", resp.StatusCode), http.StatusInternalServerError)
	}

	var expoResp ExpoResponse
	if err := json.NewDecoder(resp.Body).Decode(&expoResp); err != nil {
		return errLib.New("Failed to decode Expo response", http.StatusInternalServerError)
	}

	// Log any errors from Expo
	for _, result := range expoResp.Data {
		if result.Status == "error" {
			fmt.Printf("Expo push notification error: %s - %v\n", result.Message, result.Details)
		}
	}

	return nil
}
package notification

import (
	"api/internal/di"
	familyRepo "api/internal/domains/family/persistence"
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
	repo       *repositories.PushTokenRepository
	familyRepo *familyRepo.Repository
}

func NewNotificationService(container *di.Container) *NotificationService {
	return &NotificationService{
		repo:       repositories.NewPushTokenRepository(container),
		familyRepo: familyRepo.NewFamilyRepository(container),
	}
}

func (s *NotificationService) RegisterPushToken(ctx context.Context, userID uuid.UUID, token, deviceType string) *errLib.CommonError {
	return s.repo.UpsertPushToken(ctx, userID, token, deviceType)
}

func (s *NotificationService) SendTeamNotification(ctx context.Context, teamID uuid.UUID, notification values.TeamNotification) *errLib.CommonError {
	// Get all push tokens for team members
	tokens, err := s.repo.GetPushTokensByTeamID(ctx, teamID)
	if err != nil {
		fmt.Printf("[NOTIFICATION] Error getting push tokens for team %s: %v\n", teamID, err)
		return err
	}

	fmt.Printf("[NOTIFICATION] Sending %s notification to team %s, found %d push tokens\n", notification.Type, teamID, len(tokens))
	for i, token := range tokens {
		fmt.Printf("[NOTIFICATION]   Token %d: user_id=%s, device_type=%s, token_prefix=%s\n", i+1, token.UserID, token.DeviceType, token.ExpoPushToken[:min(len(token.ExpoPushToken), 30)])
	}

	if len(tokens) == 0 {
		fmt.Printf("[NOTIFICATION] No push tokens found for team %s\n", teamID)
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

	// Log results from Expo for each message
	for i, result := range expoResp.Data {
		tokenPrefix := ""
		if i < len(messages) {
			tokenPrefix = messages[i].To[:min(len(messages[i].To), 30)]
		}
		if result.Status == "error" {
			fmt.Printf("[NOTIFICATION] Expo ERROR for token %s: %s - %v\n", tokenPrefix, result.Message, result.Details)
		} else {
			fmt.Printf("[NOTIFICATION] Expo OK for token %s: ticket_id=%s\n", tokenPrefix, result.ID)
		}
	}

	return nil
}

// SendUserNotification sends a notification to a specific user
// This also automatically sends the notification to the user's parent if they have one
func (s *NotificationService) SendUserNotification(ctx context.Context, userID uuid.UUID, notification values.UserNotification) *errLib.CommonError {
	// Get push tokens for the user
	tokens, err := s.repo.GetPushTokensByUserID(ctx, userID)
	if err != nil {
		fmt.Printf("[NOTIFICATION] Error getting push tokens for user %s: %v\n", userID, err)
		return err
	}

	fmt.Printf("[NOTIFICATION] Sending notification to user %s, found %d push tokens\n", userID, len(tokens))

	// Check if user has a parent and get their tokens too
	user, userErr := s.familyRepo.GetUserById(ctx, userID)
	if userErr == nil && user.ParentID.Valid && user.ParentID.UUID != uuid.Nil {
		parentTokens, parentErr := s.repo.GetPushTokensByUserID(ctx, user.ParentID.UUID)
		if parentErr == nil && len(parentTokens) > 0 {
			fmt.Printf("[NOTIFICATION] User %s has parent %s, also sending to %d parent tokens\n", userID, user.ParentID.UUID, len(parentTokens))
			tokens = append(tokens, parentTokens...)
		}
	}

	if len(tokens) == 0 {
		fmt.Printf("[NOTIFICATION] No push tokens found for user %s (or parent)\n", userID)
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

// SendUserNotificationWithoutParent sends a notification to a specific user only, without duplicating to parent
func (s *NotificationService) SendUserNotificationWithoutParent(ctx context.Context, userID uuid.UUID, notification values.UserNotification) *errLib.CommonError {
	// Get push tokens for the user only
	tokens, err := s.repo.GetPushTokensByUserID(ctx, userID)
	if err != nil {
		fmt.Printf("[NOTIFICATION] Error getting push tokens for user %s: %v\n", userID, err)
		return err
	}

	fmt.Printf("[NOTIFICATION] Sending notification to user %s only, found %d push tokens\n", userID, len(tokens))

	if len(tokens) == 0 {
		fmt.Printf("[NOTIFICATION] No push tokens found for user %s\n", userID)
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
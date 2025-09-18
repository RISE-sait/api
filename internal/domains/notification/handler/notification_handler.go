package notification

import (
	"net/http"

	"api/internal/di"
	dto "api/internal/domains/notification/dto"
	service "api/internal/domains/notification/services"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(container *di.Container) *NotificationHandler {
	return &NotificationHandler{
		service: service.NewNotificationService(container),
	}
}

// RegisterPushToken registers an Expo push token for a user
// @Summary Register push notification token
// @Description Register an Expo push token for receiving notifications
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body dto.RegisterPushTokenRequestDto true "Push token registration details"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Token registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /secure/notifications/register [post]
func (h *NotificationHandler) RegisterPushToken(w http.ResponseWriter, r *http.Request) {
	var request dto.RegisterPushTokenRequestDto

	if err := validators.ParseJSON(r.Body, &request); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := request.Validate(); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Get user ID from JWT context
	userID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Register the push token
	if err := h.service.RegisterPushToken(r.Context(), userID, request.ExpoPushToken, request.DeviceType); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"message": "Push token registered successfully",
		"user_id": userID,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// SendTeamNotification sends a notification to all members of a team
// @Summary Send team notification
// @Description Send a push notification to all members of a specific team
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body dto.SendNotificationRequestDto true "Notification details"
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Notification sent successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /secure/notifications/send [post]
func (h *NotificationHandler) SendTeamNotification(w http.ResponseWriter, r *http.Request) {
	var request dto.SendNotificationRequestDto

	if err := validators.ParseJSON(r.Body, &request); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	notification, err := request.ToTeamNotification()
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Send the notification
	if err := h.service.SendTeamNotification(r.Context(), notification.TeamID, notification); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := map[string]interface{}{
		"message": "Notification sent successfully",
		"team_id": notification.TeamID,
		"type":    notification.Type,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}
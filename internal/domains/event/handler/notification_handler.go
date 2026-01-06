package event

import (
	"net/http"

	"api/internal/di"
	dto "api/internal/domains/event/dto"
	"api/internal/domains/event/service"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

// EventNotificationHandler provides HTTP handlers for event notifications
type EventNotificationHandler struct {
	notificationService *service.EventNotificationService
}

// NewEventNotificationHandler creates a new EventNotificationHandler
func NewEventNotificationHandler(container *di.Container) *EventNotificationHandler {
	return &EventNotificationHandler{
		notificationService: service.NewEventNotificationService(container),
	}
}

// GetEventCustomers retrieves all enrolled customers for an event
// @Summary Get enrolled customers for an event
// @Description Returns all customers enrolled in an event with their email and push notification status
// @Tags event-notifications
// @Accept json
// @Produce json
// @Param event_id path string true "Event ID" Format(uuid)
// @Success 200 {object} dto.EventCustomersResponseDto "List of enrolled customers"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid event ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{event_id}/customers [get]
// @Security Bearer
func (h *EventNotificationHandler) GetEventCustomers(w http.ResponseWriter, r *http.Request) {
	eventID, err := h.parseEventID(r)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Check coach authorization
	if authErr := h.checkCoachAuthorization(r, eventID); authErr != nil {
		responseHandlers.RespondWithError(w, authErr)
		return
	}

	customers, err := h.notificationService.GetEventCustomers(r.Context(), eventID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Convert to DTO
	customerDtos := make([]dto.EventCustomerDto, len(customers))
	for i, c := range customers {
		customerDtos[i] = dto.EventCustomerDto{
			ID:           c.ID,
			FirstName:    c.FirstName,
			LastName:     c.LastName,
			Email:        c.Email,
			HasPushToken: c.HasPushToken,
		}
	}

	response := dto.EventCustomersResponseDto{
		TotalCount: len(customerDtos),
		Customers:  customerDtos,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// SendNotification sends email and/or push notifications to event attendees
// @Summary Send notification to event attendees
// @Description Sends email and/or push notifications to customers enrolled in an event
// @Tags event-notifications
// @Accept json
// @Produce json
// @Param event_id path string true "Event ID" Format(uuid)
// @Param notification body dto.SendNotificationRequestDto true "Notification details"
// @Success 200 {object} dto.SendNotificationResponseDto "Notification send result"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 403 {object} map[string]interface{} "Forbidden: Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{event_id}/notifications [post]
// @Security Bearer
func (h *EventNotificationHandler) SendNotification(w http.ResponseWriter, r *http.Request) {
	eventID, err := h.parseEventID(r)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Check coach authorization
	if authErr := h.checkCoachAuthorization(r, eventID); authErr != nil {
		responseHandlers.RespondWithError(w, authErr)
		return
	}

	// Parse request body
	var requestDto dto.SendNotificationRequestDto
	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Validate request
	if err := requestDto.Validate(); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Get sender ID from context
	senderID, idErr := contextUtils.GetUserID(r.Context())
	if idErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("Failed to get user ID", http.StatusUnauthorized))
		return
	}

	// Send notification
	result, err := h.notificationService.SendNotification(r.Context(), eventID, senderID, requestDto)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.SendNotificationResponseDto{
		NotificationID: result.NotificationID,
		RecipientCount: result.RecipientCount,
		EmailSent:      result.EmailSent,
		EmailFailed:    result.EmailFailed,
		PushSent:       result.PushSent,
		PushFailed:     result.PushFailed,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetNotificationHistory retrieves notification history for an event
// @Summary Get notification history for an event
// @Description Returns all notifications that have been sent for an event
// @Tags event-notifications
// @Accept json
// @Produce json
// @Param event_id path string true "Event ID" Format(uuid)
// @Success 200 {object} dto.NotificationHistoryResponseDto "Notification history"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid event ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /events/{event_id}/notifications [get]
// @Security Bearer
func (h *EventNotificationHandler) GetNotificationHistory(w http.ResponseWriter, r *http.Request) {
	eventID, err := h.parseEventID(r)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Check coach authorization
	if authErr := h.checkCoachAuthorization(r, eventID); authErr != nil {
		responseHandlers.RespondWithError(w, authErr)
		return
	}

	history, err := h.notificationService.GetNotificationHistory(r.Context(), eventID)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if history == nil {
		history = []dto.NotificationHistoryDto{}
	}

	response := dto.NotificationHistoryResponseDto{
		Notifications: history,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// parseEventID extracts and validates the event ID from the URL
func (h *EventNotificationHandler) parseEventID(r *http.Request) (uuid.UUID, *errLib.CommonError) {
	eventIDStr := chi.URLParam(r, "event_id")
	if eventIDStr == "" {
		return uuid.Nil, errLib.New("Event ID is required", http.StatusBadRequest)
	}

	eventID, err := validators.ParseUUID(eventIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	return eventID, nil
}

// checkCoachAuthorization verifies that coaches can only access their own events
func (h *EventNotificationHandler) checkCoachAuthorization(r *http.Request, eventID uuid.UUID) *errLib.CommonError {
	role, err := contextUtils.GetUserRole(r.Context())
	if err != nil {
		return errLib.New("Failed to get user role", http.StatusUnauthorized)
	}

	// Admins, SuperAdmins, IT, and Receptionists have full access
	if role == contextUtils.RoleAdmin || role == contextUtils.RoleSuperAdmin ||
		role == contextUtils.RoleIT || role == contextUtils.RoleReceptionist {
		return nil
	}

	// Coaches need to be verified for access
	if role == contextUtils.RoleCoach {
		staffID, idErr := contextUtils.GetUserID(r.Context())
		if idErr != nil {
			return errLib.New("Failed to get user ID", http.StatusUnauthorized)
		}

		hasAccess, accessErr := h.notificationService.CheckCoachHasAccessToEvent(r.Context(), eventID, staffID)
		if accessErr != nil {
			return accessErr
		}

		if !hasAccess {
			return errLib.New("You do not have access to this event", http.StatusForbidden)
		}
	}

	return nil
}

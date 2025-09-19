package haircut_event

import (
	"api/internal/di"
	dto "api/internal/domains/haircut/event/dto"
	repository "api/internal/domains/haircut/event/persistence"
	db "api/internal/domains/haircut/event/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
	"time"

	contextUtils "api/utils/context"
)

// EventsHandler provides HTTP handlers for managing events.
type EventsHandler struct {
	Repo *repository.Repository
}

func NewEventsHandler(container *di.Container) *EventsHandler {
	return &EventsHandler{Repo: repository.NewEventsRepository(container)}
}

// GetEvents retrieves all haircut events based on filter criteria.
// @Summary Get all haircut events
// @Description Retrieve all haircut events, with optional filters by barber ID and customer ID.
// @Tags haircuts
// @Accept json
// @Produce json
// @Param after query string false "Start date of the events range (YYYY-MM-DD)" example("2025-03-01")
// @Param before query string false "End date of the events range (YYYY-MM-DD)" example("2025-03-31")
// @Param barber_id query string false "Filter by barber ID"
// @Param customer_id query string false "Filter by customer ID"
// @Success 200 {array} dto.EventResponseDto "List of haircut events retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/events [get]
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()

	var (
		barberID, customerID uuid.UUID
		before, after        time.Time
	)

	if afterStr := query.Get("after"); afterStr != "" {
		if afterDate, formatErr := time.Parse("2006-01-02", afterStr); formatErr != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'after' date format, expected YYYY-MM-DD", http.StatusBadRequest))
			return
		} else {
			after = afterDate
		}
	}

	if beforeStr := query.Get("before"); beforeStr != "" {
		if beforeDate, formatErr := time.Parse("2006-01-02", beforeStr); formatErr != nil {
			responseHandlers.RespondWithError(w, errLib.New("invalid 'before' date format, expected YYYY-MM-DD", http.StatusBadRequest))
			return
		} else {
			before = beforeDate
		}
	}

	if barberIdStr := query.Get("barber_id"); barberIdStr != "" {
		if id, err := validators.ParseUUID(barberIdStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			barberID = id
		}
	}

	if customerIdStr := query.Get("customer_id"); customerIdStr != "" {
		if id, err := validators.ParseUUID(customerIdStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			customerID = id
		}
	}

	if (after.IsZero() || before.IsZero()) && (barberID == uuid.Nil && customerID == uuid.Nil) {
		responseHandlers.RespondWithError(w, errLib.New("at least one of (before and after) or one of (barber_id, customer_id) must be provided", http.StatusBadRequest))
		return
	}

	events, err := h.Repo.GetEvents(r.Context(), barberID, customerID, before, after)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.EventResponseDto, len(events))

	for i, event := range events {
		result[i] = dto.NewEventResponse(event)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// CreateEvent creates a new haircut event.
// @Description Registers a new haircut event with the provided details from request body.
// @Description Requires an Authorization header containing the customer's JWT, ensuring only logged-in customers can make the request.
// @Tags haircuts
// @Accept json
// @Produce json
// @Security Bearer
// @Param event body dto.RequestDto true "Haircut event details"
// @Success 201 {object} dto.EventResponseDto "Haircut event created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/events [post]
func (h *EventsHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {

	customerID, ctxErr := contextUtils.GetUserID(r.Context())

	if ctxErr != nil {
		responseHandlers.RespondWithError(w, ctxErr)
		return
	}

	var targetBody dto.RequestDto

	if err := validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	eventCreateValues, err := targetBody.ToCreateEventValue(customerID)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	createdEvent, err := h.Repo.CreateEvent(r.Context(), eventCreateValues)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.NewEventResponse(createdEvent)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
}

// DeleteEvent deletes a haircut event by ID.
// @Description Deletes a haircut event by its ID.
// @Tags haircuts
// @Accept json
// @Produce json
// @Param id path string true "Haircut event ID"
// @Success 204 "No Content: Haircut event deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Haircut event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/events/{id} [delete]
func (h *EventsHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
	}

	if err = h.Repo.DeleteEvent(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// GetEvent retrieves information about a specific haircut event.
// @Description Retrieves details of a specific haircut event based on its ID.
// @Tags haircuts
// @Accept json
// @Produce json
// @Param id path string true "Haircut event ID"
// @Success 200 {object} dto.EventResponseDto "Haircut event details retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Haircut event not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/events/{id} [get]
func (h *EventsHandler) GetEvent(w http.ResponseWriter, r *http.Request) {

	var eventId uuid.UUID

	if eventIdStr := chi.URLParam(r, "id"); eventIdStr != "" {
		id, err := validators.ParseUUID(eventIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		eventId = id
	}

	event, err := h.Repo.GetEvent(r.Context(), eventId)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.NewEventResponse(event)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusOK)
}

// GetAvailableTimeSlots retrieves available time slots for a barber on a specific date.
// @Summary Get available time slots for a barber
// @Description Get available booking slots for a specific barber on a given date, considering their working hours and existing bookings.
// @Tags haircuts
// @Accept json
// @Produce json
// @Param barber_id path string true "Barber ID"
// @Param date query string true "Date in YYYY-MM-DD format" example("2025-09-20")
// @Param service_duration query int false "Service duration in minutes (default: 30)" example(30)
// @Success 200 {object} map[string]interface{} "Available time slots"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/barbers/{barber_id}/availability [get]
func (h *EventsHandler) GetAvailableTimeSlots(w http.ResponseWriter, r *http.Request) {
	barberIDStr := chi.URLParam(r, "barber_id")
	barberID, err := validators.ParseUUID(barberIDStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		responseHandlers.RespondWithError(w, errLib.New("date parameter is required", http.StatusBadRequest))
		return
	}

	date, parseErr := time.Parse("2006-01-02", dateStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, errLib.New("invalid date format, expected YYYY-MM-DD", http.StatusBadRequest))
		return
	}

	// Validate date is not in the past
	today := time.Now().Truncate(24 * time.Hour)
	if date.Before(today) {
		responseHandlers.RespondWithError(w, errLib.New("cannot get availability for past dates", http.StatusBadRequest))
		return
	}

	// Get service duration (default 30 minutes)
	serviceDuration := 30
	if durationStr := r.URL.Query().Get("service_duration"); durationStr != "" {
		if duration, parseErr := time.ParseDuration(durationStr + "m"); parseErr == nil {
			serviceDuration = int(duration.Minutes())
		}
	}

	availableSlots, repoErr := h.Repo.GetAvailableTimeSlots(r.Context(), barberID, date, serviceDuration)
	if repoErr != nil {
		responseHandlers.RespondWithError(w, repoErr)
		return
	}

	result := map[string]interface{}{
		"barber_id":        barberID,
		"date":            date.Format("2006-01-02"),
		"service_duration": serviceDuration,
		"available_slots":  availableSlots,
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// ===== BARBER AVAILABILITY MANAGEMENT ENDPOINTS =====

// GetMyAvailability retrieves the current barber's availability schedule.
// @Summary Get my availability schedule  
// @Description Get all availability records for the authenticated barber
// @Tags barber-availability
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} dto.WeeklyAvailabilityResponseDto "Barber availability schedule"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/barbers/me/availability [get]
func (h *EventsHandler) GetMyAvailability(w http.ResponseWriter, r *http.Request) {
	barberID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	availability, repoErr := h.Repo.GetBarberFullAvailability(r.Context(), barberID)
	if repoErr != nil {
		responseHandlers.RespondWithError(w, repoErr)
		return
	}

	// Convert to response DTOs
	availabilityDtos := make([]dto.AvailabilityResponseDto, len(availability))
	for i, avail := range availability {
		availabilityDtos[i] = dto.NewAvailabilityResponse(
			avail.ID, avail.DayOfWeek, avail.StartTime, avail.EndTime,
			avail.IsActive, avail.CreatedAt, avail.UpdatedAt,
		)
	}

	result := dto.WeeklyAvailabilityResponseDto{
		BarberID:     barberID,
		Availability: availabilityDtos,
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// SetMyAvailability sets availability for a specific day.
// @Summary Set availability for a day
// @Description Set working hours for a specific day of the week
// @Tags barber-availability
// @Accept json
// @Produce json
// @Security Bearer
// @Param availability body dto.SetAvailabilityDto true "Availability details"
// @Success 201 {object} dto.AvailabilityResponseDto "Availability created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 409 {object} map[string]interface{} "Conflict: Availability already exists"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/barbers/me/availability [post]
func (h *EventsHandler) SetMyAvailability(w http.ResponseWriter, r *http.Request) {
	barberID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var requestDto dto.SetAvailabilityDto
	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := requestDto.Validate(); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Convert to repository parameters
	startTime, _ := time.Parse("15:04", requestDto.StartTime)
	endTime, _ := time.Parse("15:04", requestDto.EndTime)
	
	isActive := true
	if requestDto.IsActive != nil {
		isActive = *requestDto.IsActive
	}

	params := db.InsertBarberAvailabilityParams{
		BarberID:  barberID,
		DayOfWeek: int32(requestDto.DayOfWeek),
		StartTime: startTime,
		EndTime:   endTime,
		IsActive:  isActive,
	}

	createdAvailability, repoErr := h.Repo.CreateBarberAvailability(r.Context(), params)
	if repoErr != nil {
		responseHandlers.RespondWithError(w, repoErr)
		return
	}

	result := dto.NewAvailabilityResponse(
		createdAvailability.ID, createdAvailability.DayOfWeek,
		createdAvailability.StartTime, createdAvailability.EndTime,
		createdAvailability.IsActive, createdAvailability.CreatedAt, createdAvailability.UpdatedAt,
	)

	responseHandlers.RespondWithSuccess(w, result, http.StatusCreated)
}

// BulkSetMyAvailability sets availability for multiple days at once.
// @Summary Set availability for multiple days
// @Description Set working hours for multiple days of the week in one request
// @Tags barber-availability
// @Accept json
// @Produce json
// @Security Bearer
// @Param availability body dto.BulkSetAvailabilityDto true "Multiple availability records"
// @Success 200 {object} dto.WeeklyAvailabilityResponseDto "All availability set successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/barbers/me/availability/bulk [post]
func (h *EventsHandler) BulkSetMyAvailability(w http.ResponseWriter, r *http.Request) {
	barberID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var requestDto dto.BulkSetAvailabilityDto
	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := requestDto.Validate(); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Use upsert for bulk operations to avoid conflicts
	var createdAvailability []db.HaircutBarberAvailability
	for _, avail := range requestDto.Availability {
		startTime, _ := time.Parse("15:04", avail.StartTime)
		endTime, _ := time.Parse("15:04", avail.EndTime)
		
		isActive := true
		if avail.IsActive != nil {
			isActive = *avail.IsActive
		}

		params := db.UpsertBarberAvailabilityParams{
			BarberID:  barberID,
			DayOfWeek: int32(avail.DayOfWeek),
			StartTime: startTime,
			EndTime:   endTime,
			IsActive:  isActive,
		}

		result, repoErr := h.Repo.UpsertBarberAvailability(r.Context(), params)
		if repoErr != nil {
			responseHandlers.RespondWithError(w, repoErr)
			return
		}
		createdAvailability = append(createdAvailability, result)
	}

	// Convert to response DTOs
	availabilityDtos := make([]dto.AvailabilityResponseDto, len(createdAvailability))
	for i, avail := range createdAvailability {
		availabilityDtos[i] = dto.NewAvailabilityResponse(
			avail.ID, avail.DayOfWeek, avail.StartTime, avail.EndTime,
			avail.IsActive, avail.CreatedAt, avail.UpdatedAt,
		)
	}

	result := dto.WeeklyAvailabilityResponseDto{
		BarberID:     barberID,
		Availability: availabilityDtos,
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateMyAvailability updates a specific availability record.
// @Summary Update availability record
// @Description Update an existing availability record by ID
// @Tags barber-availability
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Availability ID"
// @Param availability body dto.UpdateAvailabilityDto true "Updated availability details"
// @Success 200 {object} dto.AvailabilityResponseDto "Availability updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Not Found: Availability record not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/barbers/me/availability/{id} [put]
func (h *EventsHandler) UpdateMyAvailability(w http.ResponseWriter, r *http.Request) {
	barberID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var requestDto dto.UpdateAvailabilityDto
	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := requestDto.Validate(); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Verify ownership
	existing, repoErr := h.Repo.GetBarberAvailabilityByID(r.Context(), id)
	if repoErr != nil {
		responseHandlers.RespondWithError(w, repoErr)
		return
	}

	if existing.BarberID != barberID {
		responseHandlers.RespondWithError(w, errLib.New("You can only update your own availability", http.StatusForbidden))
		return
	}

	// Convert to repository parameters
	startTime, _ := time.Parse("15:04", requestDto.StartTime)
	endTime, _ := time.Parse("15:04", requestDto.EndTime)
	
	isActive := existing.IsActive
	if requestDto.IsActive != nil {
		isActive = *requestDto.IsActive
	}

	params := db.UpdateBarberAvailabilityParams{
		ID:        id,
		StartTime: startTime,
		EndTime:   endTime,
		IsActive:  isActive,
	}

	updatedAvailability, repoErr := h.Repo.UpdateBarberAvailability(r.Context(), params)
	if repoErr != nil {
		responseHandlers.RespondWithError(w, repoErr)
		return
	}

	result := dto.NewAvailabilityResponse(
		updatedAvailability.ID, updatedAvailability.DayOfWeek,
		updatedAvailability.StartTime, updatedAvailability.EndTime,
		updatedAvailability.IsActive, updatedAvailability.CreatedAt, updatedAvailability.UpdatedAt,
	)

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// DeleteMyAvailability deletes a specific availability record.
// @Summary Delete availability record
// @Description Delete an existing availability record by ID
// @Tags barber-availability
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Availability ID"
// @Success 204 "No Content: Availability deleted successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Not Found: Availability record not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /haircuts/barbers/me/availability/{id} [delete]
func (h *EventsHandler) DeleteMyAvailability(w http.ResponseWriter, r *http.Request) {
	barberID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Verify ownership
	existing, repoErr := h.Repo.GetBarberAvailabilityByID(r.Context(), id)
	if repoErr != nil {
		responseHandlers.RespondWithError(w, repoErr)
		return
	}

	if existing.BarberID != barberID {
		responseHandlers.RespondWithError(w, errLib.New("You can only delete your own availability", http.StatusForbidden))
		return
	}

	if repoErr := h.Repo.DeleteBarberAvailability(r.Context(), id); repoErr != nil {
		responseHandlers.RespondWithError(w, repoErr)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

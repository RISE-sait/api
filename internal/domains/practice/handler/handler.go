package practice

import (
	"api/internal/di"
	dto "api/internal/domains/practice/dto"
	service "api/internal/domains/practice/services"
	values "api/internal/domains/practice/values"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	Service *service.Service
}

func NewHandler(container *di.Container) *Handler {
	return &Handler{Service: service.NewService(container)}
}

// CreatePractice creates a new practice.
// @Summary Create a practice
// @Description Creates a new practice session.
// @Tags practices
// @Accept json
// @Produce json
// @Param practice body dto.RequestDto true "Practice details"
// @Security Bearer
// @Success 201 {object} nil "Practice created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices [post]
func (h *Handler) CreatePractice(w http.ResponseWriter, r *http.Request) {
	var req dto.RequestDto
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	val, err := req.ToCreateValue()
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	if err = h.Service.CreatePractice(r.Context(), val); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetPractice retrieves a practice by ID.
// @Summary Get a practice by ID
// @Description Fetches a single practice session using its UUID.
// @Tags practices
// @Produce json
// @Param id path string true "Practice ID"
// @Success 200 {object} dto.ResponseDto "Practice retrieved"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Practice not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices/{id} [get]
func (h *Handler) GetPractice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	p, err := h.Service.GetPractice(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, dto.NewResponse(p), http.StatusOK)
}

// GetPractices returns a paginated list of practices.
// @Summary List practices
// @Description Retrieves a list of practices, optionally filtered by team.
// @Tags practices
// @Produce json
// @Param team_id query string false "Team UUID"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page (max 100)"
// @Success 200 {array} dto.ResponseDto "List of practices"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices [get]
func (h *Handler) GetPractices(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit
	var teamID uuid.UUID
	if val := query.Get("team_id"); val != "" {
		id, err := validators.ParseUUID(val)
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
		teamID = id
	}
	res, err := h.Service.GetPractices(r.Context(), teamID, int32(limit), int32(offset))
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	dtoRes := make([]dto.ResponseDto, len(res))
	for i, p := range res {
		dtoRes[i] = dto.NewResponse(p)
	}
	responseHandlers.RespondWithSuccess(w, dtoRes, http.StatusOK)
}

// UpdatePractice modifies an existing practice.
// @Summary Update a practice
// @Description Updates the details of a specific practice.
// @Tags practices
// @Accept json
// @Produce json
// @Param id path string true "Practice ID"
// @Param practice body dto.RequestDto true "Updated practice details"
// @Security Bearer
// @Success 204 "Practice updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Practice not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices/{id} [put]
func (h *Handler) UpdatePractice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var req dto.RequestDto
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	val, err := req.ToUpdateValue(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	if err = h.Service.UpdatePractice(r.Context(), val); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeletePractice removes a practice by ID.
// @Summary Delete a practice
// @Description Deletes a specific practice session.
// @Tags practices
// @Produce json
// @Param id path string true "Practice ID"
// @Security Bearer
// @Success 204 "Practice deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Practice not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices/{id} [delete]
func (h *Handler) DeletePractice(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	if err = h.Service.DeletePractice(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// CreateRecurringPractices creates recurring practice sessions.
// @Summary Create recurring practices
// @Description Creates multiple recurring practices using recurrence rules.
// @Tags practices
// @Accept json
// @Produce json
// @Param recurrence body dto.RecurrenceRequestDto true "Recurring practice details"
// @Security Bearer
// @Success 201 {object} nil "Recurring practices created"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /practices/recurring [post]
func (h *Handler) CreateRecurringPractices(w http.ResponseWriter, r *http.Request) {
	var req dto.RecurrenceRequestDto
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	rec, err := req.ToRecurrenceValues()
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	base := values.CreatePracticeValue{
		TeamID:     req.TeamID,
		LocationID: req.LocationID,
		CourtID:    req.CourtID,
		Status:     req.Status,
	}
	if err = h.Service.CreateRecurringPractices(r.Context(), rec, base); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}
	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

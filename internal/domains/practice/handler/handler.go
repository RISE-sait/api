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
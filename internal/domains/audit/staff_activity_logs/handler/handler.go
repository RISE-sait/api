package staff_activity_logs

import (
	"api/internal/di"
	dto "api/internal/domains/audit/staff_activity_logs/dto"
	service "api/internal/domains/audit/staff_activity_logs/service"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/google/uuid"
	"net/http"
	"strconv"
)

// Handler provides HTTP handlers for managing events.
type Handler struct {
	EventsService *service.Service
}

func NewHandler(container *di.Container) *Handler {
	return &Handler{EventsService: service.NewService(container)}
}

// GetStaffActivityLogs retrieves all staff activity logs based on filter criteria.
// @Tags staff_activity_logs
// @Summary Get staff activity logs
// @Description Retrieves a paginated list of staff activity logs with optional filtering
// @Param staff_id query string false "Filter by staff member ID (UUID format)" example("550e8400-e29b-41d4-a716-446655440000")
// @Param search_description query string false "Search term to filter activity descriptions (case-insensitive partial match)"
// @Param limit query int false "Number of records to return (default: 10)" example(10)
// @Param offset query int false "Number of records to skip for pagination (default: 0)" example(0)
// @Produce json
// @Success 200 {array} dto.StaffActivityLogResponse "List of staff activity logs retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input format"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /staffs/logs [get]
func (h *Handler) GetStaffActivityLogs(w http.ResponseWriter, r *http.Request) {

	var (
		staffID           uuid.UUID
		searchDescription string
		limit, offset     int32
	)

	query := r.URL.Query()

	searchDescription = query.Get("search_description")

	limitStr := query.Get("limit")

	if limitStr != "" {
		limitB64, _ := strconv.ParseInt(limitStr, 10, 32)
		limit = int32(limitB64)
	} else {
		limit = 10
	}

	offsetStr := query.Get("offset")
	if offsetStr != "" {
		offsetB64, _ := strconv.ParseInt(offsetStr, 10, 32)
		offset = int32(offsetB64)
	} else {
		offset = 0
	}

	if staffIDStr := query.Get("staff_id"); staffIDStr != "" {
		if id, err := validators.ParseUUID(staffIDStr); err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		} else {
			staffID = id
		}
	}

	logs, err := h.EventsService.GetStaffActivityLogs(r.Context(), staffID, searchDescription, limit, offset)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseDto := make([]dto.StaffActivityLogResponse, len(logs))

	for i, log := range logs {

		responseDto[i] = dto.NewStaffActivityLogResponse(log)
	}

	responseHandlers.RespondWithSuccess(w, responseDto, http.StatusOK)
}

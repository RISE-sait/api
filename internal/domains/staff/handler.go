package staff

import (
	"api/internal/di"
	dto "api/internal/domains/staff/dto"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handlers struct {
	Service *Service
}

func NewStaffHandlers(container *di.Container) *Handlers {
	return &Handlers{Service: NewStaffService(container)}
}

// GetStaffs retrieves a list of staff members.
// @Summary Get a list of staff members
// @Description Get a list of staff members
// @Tags staff
// @Accept json
// @Produce json
// @Param role query string false "RoleName HubSpotId to filter staff" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Success 200 {array} staff.ResponseDto "GetMemberships of staff members retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/staffs [get]
func (h *Handlers) GetStaffs(w http.ResponseWriter, r *http.Request) {

	var rolePtr *uuid.UUID

	roleIdStr := r.URL.Query().Get("role")

	if roleIdStr != "" {
		id, err := validators.ParseUUID(roleIdStr)
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
		rolePtr = &id

	}

	staffs, err := h.Service.GetStaffs(r.Context(), rolePtr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.ResponseDto, len(staffs))
	for i, staff := range staffs {
		result[i] = dto.NewStaffResponse(staff)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateStaff updates an existing staff member.
// @Summary Update a staff member
// @Description Update a staff member
// @Tags staff
// @Accept json
// @Produce json
// @Param id path string true "Staff HubSpotId" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Param staff body dto.RequestDto true "Staff details"
// @Success 204 "No Content: Staff updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Staff not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/staffs/{id} [put]
// @Security Bearer
func (h *Handlers) UpdateStaff(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	staffUpdateFields, err := requestDto.ToEntity(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	staff, err := h.Service.UpdateStaff(r.Context(), staffUpdateFields)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.NewStaffResponse(*staff)

	responseHandlers.RespondWithSuccess(w, response, http.StatusNoContent)
}

// DeleteStaff deletes a staff member by HubSpotId.
// @Summary Delete a staff member
// @Description Delete a staff member by HubSpotId
// @Tags staff
// @Accept json
// @Produce json
// @Param id path string true "Staff HubSpotId" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Success 204 "No Content: Staff deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid HubSpotId"
// @Failure 404 {object} map[string]interface{} "Not Found: Staff not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/staffs/{id} [delete]
// @Security Bearer
func (h *Handlers) DeleteStaff(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.DeleteStaff(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

package staff

import (
	"api/internal/di"
	dto "api/internal/domains/staff/dto"
	values "api/internal/domains/staff/values"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type StaffController struct {
	Service *StaffService
}

func NewStaffController(container *di.Container) *StaffController {
	return &StaffController{Service: NewStaffService(container)}
}

// GetStaffs retrieves a list of staff members.
// @Summary Get a list of staff members
// @Description Get a list of staff members
// @Tags staff
// @Accept json
// @Produce json
// @Param role query string false "Role ID to filter staff" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Success 200 {array} dto.StaffResponseDto "List of staff members retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/staffs [get]
func (h *StaffController) GetStaffs(w http.ResponseWriter, r *http.Request) {

	var rolePtr *uuid.UUID

	roleIdStr := r.URL.Query().Get("role")

	if roleIdStr != "" {
		id, err := validators.ParseUUID(roleIdStr)
		if err != nil {
			response_handlers.RespondWithError(w, err)
			return
		}
		rolePtr = &id

	}

	staffs, err := h.Service.GetStaffs(r.Context(), rolePtr)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.StaffResponseDto, len(staffs))
	for i, staff := range staffs {
		result[i] = mapEntityToResponse(&staff)
	}

	response_handlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateStaff updates an existing staff member.
// @Summary Update a staff member
// @Description Update a staff member
// @Tags staff
// @Accept json
// @Produce json
// @Param id path string true "Staff ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Param staff body dto.StaffRequestDto true "Staff details"
// @Success 204 "No Content: Staff updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Staff not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/staffs/{id} [put]
// @Security Bearer
func (h *StaffController) UpdateStaff(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var dto dto.StaffRequestDto

	if err := validators.ParseJSON(r.Body, &dto); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	staffUpdateFields, err := dto.ToUpdateValueObjects(idStr)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	staffAllFields, err := h.Service.UpdateStaff(r.Context(), staffUpdateFields)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response := mapEntityToResponse(staffAllFields)

	response_handlers.RespondWithSuccess(w, response, http.StatusNoContent)
}

// DeleteStaff deletes a staff member by ID.
// @Summary Delete a staff member
// @Description Delete a staff member by ID
// @Tags staff
// @Accept json
// @Produce json
// @Param id path string true "Staff ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Success 204 "No Content: Staff deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Staff not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/staffs/{id} [delete]
// @Security Bearer
func (h *StaffController) DeleteStaff(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	if err := h.Service.DeleteStaff(r.Context(), id); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

func mapEntityToResponse(staff *values.StaffAllFields) dto.StaffResponseDto {
	return dto.StaffResponseDto{
		Id:        staff.ID,
		IsActive:  staff.IsActive,
		CreatedAt: staff.CreatedAt,
		UpdatedAt: staff.UpdatedAt,
		RoleID:    staff.RoleID,
		RoleName:  staff.RoleName,
	}
}

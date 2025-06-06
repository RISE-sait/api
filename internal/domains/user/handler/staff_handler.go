package user

import (
	"api/internal/di"
	dto "api/internal/domains/user/dto/staff"
	repo "api/internal/domains/user/persistence/repository"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"api/internal/services/hubspot"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

type StaffHandler struct {
	StaffRepo      *repo.StaffRepository
	HubspotService *hubspot.Service
}

func NewStaffHandlers(container *di.Container) *StaffHandler {
	return &StaffHandler{
		StaffRepo:      repo.NewStaffRepository(container),
		HubspotService: container.HubspotService}
}

// GetStaffs retrieves a list of staff members based on optional role filter.
// @Tags staff
// @Accept json
// @Produce json
// @Param role query string false "Role name to filter staff" example("Coach")
// @Success 200 {array} dto.ResponseDto "List of staff members retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /staffs [get]
func (h *StaffHandler) GetStaffs(w http.ResponseWriter, r *http.Request) {

	role := r.URL.Query().Get("role")

	staffs, err := h.StaffRepo.List(r.Context(), strings.ToLower(role))
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.ResponseDto, len(staffs))
	for i, staff := range staffs {
		staffResponse := dto.NewStaffResponse(staff)

		if staff.CoachStatsReadValues != nil {
			staffResponse.CoachStats = &dto.CoachStatsResponseDto{
				Wins:   staff.CoachStatsReadValues.Wins,
				Losses: staff.CoachStatsReadValues.Losses,
			}
		}

		result[i] = staffResponse
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateStaff updates an existing staff member.
// @Tags staff
// @Accept json
// @Produce json
// @Param id path string true "Staff ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Param staff body dto.RequestDto true "Staff details"
// @Success 204 "No Content: Staff updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Staff not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /staffs/{id} [put]
// @Security Bearer
func (h *StaffHandler) UpdateStaff(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	staffUpdateFields, err := requestDto.ToUpdateRequestValues(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var availableRoles []string

	if roles, err := h.StaffRepo.GetAvailableStaffRoles(r.Context()); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	} else {
		for _, role := range roles {
			availableRoles = append(availableRoles, role.RoleName)
		}
	}

	// Check if the role exists
	roleExists := false

	for _, role := range availableRoles {
		if role == staffUpdateFields.RoleName {
			roleExists = true
			break
		}
	}

	if !roleExists {
		responseHandlers.RespondWithError(w, errLib.New(fmt.Sprintf("Role not found. Available roles: %v", availableRoles), http.StatusNotFound))
		return
	}

	if err = h.StaffRepo.Update(r.Context(), staffUpdateFields); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteStaff deletes a staff member by ID.
// @Tags staff
// @Accept json
// @Produce json
// @Param id path string true "Staff ID" example("f47ac10b-58cc-4372-a567-0e02b2c3d479")
// @Success 204 "No Content: Staff deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Staff not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /staffs/{id} [delete]
// @Security Bearer
func (h *StaffHandler) DeleteStaff(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.StaffRepo.Delete(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

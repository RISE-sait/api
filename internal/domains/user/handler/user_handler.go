package user

import (
	"api/internal/di"
	dto "api/internal/domains/user/dto"
	repo "api/internal/domains/user/persistence/repository"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"api/internal/services/hubspot"
	contextUtils "api/utils/context"
	"github.com/go-chi/chi"
	"net/http"
)

type UsersHandler struct {
	UsersRepo      *repo.UsersRepository
	HubspotService *hubspot.Service
}

func NewUsersHandlers(container *di.Container) *UsersHandler {
	return &UsersHandler{
		UsersRepo:      repo.NewUsersRepository(container),
		HubspotService: container.HubspotService}
}

// UpdateUser updates an existing user by ID.
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Param user body dto.UpdateRequestDto true "Updated user details"
// @Success 204 {object} map[string]interface{} "No Content: User updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 403 {object} map[string]interface{} "Forbidden: Unauthorized access"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id} [put]
func (h *UsersHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {

	// Extract and validate input

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var targetBody dto.UpdateRequestDto

	if err = validators.ParseJSON(r.Body, &targetBody); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	userUpdateFields, err := targetBody.ToUpdateValue(id)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Get auth context

	loggedInUserId, err := contextUtils.GetUserID(r.Context())

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	loggedInUserRole := r.Context().Value(contextUtils.RoleKey).(contextUtils.CtxRole)
	isLoggedInUserSameAsUpdateTarget := id == loggedInUserId

	isAdmin := loggedInUserRole == contextUtils.RoleAdmin || loggedInUserRole == contextUtils.RoleSuperAdmin || loggedInUserRole == contextUtils.RoleIT

	if !isAdmin && !isLoggedInUserSameAsUpdateTarget {
		responseHandlers.RespondWithError(w, errLib.New("You do not have permission to access this resource", http.StatusForbidden))
		return
	}

	if err = h.UsersRepo.UpdateUser(r.Context(), userUpdateFields); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)

}

package team

import (
	"api/internal/di"
	dto "api/internal/domains/team/dto"
	repository "api/internal/domains/team/persistence"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	Repo *repository.Repository
}

func NewHandler(container *di.Container) *Handler {
	return &Handler{Repo: repository.NewTeamRepository(container.Queries.TeamDb)}
}

// CreateTeam creates a new team.
// @Tags teams
// @Accept json
// @Produce json
// @Param team body dto.RequestDto true "Team details"
// @Security Bearer
// @Success 201 {object} map[string]interface{} "Team created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /teams [post]
func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	teamCreate, err := requestDto.ToCreateValueObjects()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.Create(r.Context(), teamCreate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetTeams retrieves all teams.
// @Tags teams
// @Accept json
// @Produce json
// @Success 200 {array} dto.Response "Teams retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /teams [get]
func (h *Handler) GetTeams(w http.ResponseWriter, r *http.Request) {

	teams, err := h.Repo.List(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.Response, len(teams))

	for i, team := range teams {
		result[i] = dto.Response{
			ID:        team.ID,
			Name:      team.TeamDetails.Name,
			Capacity:  team.TeamDetails.Capacity,
			CoachID:   team.TeamDetails.CoachID,
			CreatedAt: team.CreatedAt,
			UpdatedAt: team.UpdatedAt,
		}
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateTeam updates an existing team.
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Param team body dto.RequestDto true "Team details"
// @Security Bearer
// @Success 204 "No Content: Team updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Team not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /teams/{id} [put]
func (h *Handler) UpdateTeam(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	details, err := requestDto.ToUpdateValueObjects(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.Update(r.Context(), details); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteTeam deletes a team by ID.
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Security Bearer
// @Success 204 "No Content: Team deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Team not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /teams/{id} [delete]
func (h *Handler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.Delete(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

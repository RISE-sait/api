package team

import (
	"api/internal/di"
	dto "api/internal/domains/team/dto"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/google/uuid"

	"github.com/go-chi/chi"
)

type Handler struct {
	Service *Service
}

func NewHandler(container *di.Container) *Handler {
	return &Handler{
		Service: NewService(container),
	}
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

	if err = h.Service.Create(r.Context(), teamCreate); err != nil {
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

	teams, err := h.Service.GetTeams(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.Response, len(teams))

	for i, team := range teams {
		response := dto.Response{
			ID:        team.ID,
			Name:      team.TeamDetails.Name,
			Capacity:  team.TeamDetails.Capacity,
			CreatedAt: team.CreatedAt,
			UpdatedAt: team.UpdatedAt,
		}

		if team.TeamDetails.CoachID != uuid.Nil {
			response.Coach = &dto.Coach{
				ID:    team.TeamDetails.CoachID,
				Name:  team.TeamDetails.CoachName,
				Email: team.TeamDetails.CoachEmail,
			}
		}

		result[i] = response
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// GetTeamByID retrieves team by ID.
// @Tags teams
// @Accept json
// @Produce json
// @Param id path string true "Team ID"
// @Success 200 {object} dto.Response "Team retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Team not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /teams/{id} [get]
func (h *Handler) GetTeamByID(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	team, err := h.Service.GetTeamByID(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.Response{
		ID:        team.ID,
		Name:      team.TeamDetails.Name,
		Capacity:  team.TeamDetails.Capacity,
		CreatedAt: team.CreatedAt,
		UpdatedAt: team.UpdatedAt,
	}

	if team.TeamDetails.CoachID != uuid.Nil {
		response.Coach = &dto.Coach{
			ID:    team.TeamDetails.CoachID,
			Name:  team.TeamDetails.CoachName,
			Email: team.TeamDetails.CoachEmail,
		}
	}

	roster := make([]dto.RosterMemberInfo, len(team.Roster))

	for i, member := range team.Roster {

		roster[i] = dto.RosterMemberInfo{
			ID:       member.ID,
			Name:     member.Name,
			Email:    member.Email,
			Country:  member.Country,
			Points:   member.Points,
			Wins:     member.Wins,
			Losses:   member.Losses,
			Assists:  member.Assists,
			Rebounds: member.Rebounds,
			Steals:   member.Steals,
		}
	}

	response.Roster = &roster

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
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

	if err = h.Service.UpdateTeam(r.Context(), details); err != nil {
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

	if err = h.Service.DeleteTeam(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

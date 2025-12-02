package team

import (
	"api/internal/di"
	dto "api/internal/domains/team/dto"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	contextUtils "api/utils/context"
	"net/http"
	"strconv"

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
// @Param team body dto.RequestDto true "Team details including name, capacity, coach_id, and optional logo_url"
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
			ID:         team.ID,
			Name:       team.TeamDetails.Name,
			Capacity:   team.TeamDetails.Capacity,
			LogoURL:    team.TeamDetails.LogoURL,
			IsExternal: team.IsExternal,
			CreatedAt:  team.CreatedAt,
			UpdatedAt:  team.UpdatedAt,
		}

		if team.TeamDetails.CoachID != uuid.Nil {
			response.Coach = &dto.Coach{
				ID:    team.TeamDetails.CoachID,
				Name:  team.TeamDetails.CoachName,
				Email: team.TeamDetails.CoachEmail,
			}
		}

		roster := make([]dto.RosterMemberInfo, len(team.Roster))
		for j, member := range team.Roster {
			roster[j] = dto.RosterMemberInfo{
				ID:       member.ID,
				Name:     member.Name,
				Email:    member.Email,
				Country:  member.Country,
				PhotoURL: member.PhotoURL,
				Points:   member.Points,
				Wins:     member.Wins,
				Losses:   member.Losses,
				Assists:  member.Assists,
				Rebounds: member.Rebounds,
				Steals:   member.Steals,
			}
		}
		response.Roster = &roster

		result[i] = response
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// GetMyTeams retrieves teams based on user role - coaches see only their teams.
// @Summary Get my teams (role-based)
// @Description Retrieves teams based on user role. Coaches see only teams they coach, admins see all teams.
// @Tags teams
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {array} dto.Response "Teams retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden: Role not supported"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /secure/teams [get]
func (h *Handler) GetMyTeams(w http.ResponseWriter, r *http.Request) {
	role, err := contextUtils.GetUserRole(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Admins can view all teams
	if role == contextUtils.RoleAdmin || role == contextUtils.RoleSuperAdmin || role == contextUtils.RoleIT {
		h.GetTeams(w, r)
		return
	}

	// Coaches can view only their teams
	if role == contextUtils.RoleCoach {
		userID, err := contextUtils.GetUserID(r.Context())
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		teams, err := h.Service.GetTeamsByCoach(r.Context(), userID)
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		result := make([]dto.Response, len(teams))

		for i, team := range teams {
			response := dto.Response{
				ID:         team.ID,
				Name:       team.TeamDetails.Name,
				Capacity:   team.TeamDetails.Capacity,
				LogoURL:    team.TeamDetails.LogoURL,
				IsExternal: team.IsExternal,
				CreatedAt:  team.CreatedAt,
				UpdatedAt:  team.UpdatedAt,
			}

			if team.TeamDetails.CoachID != uuid.Nil {
				response.Coach = &dto.Coach{
					ID:    team.TeamDetails.CoachID,
					Name:  team.TeamDetails.CoachName,
					Email: team.TeamDetails.CoachEmail,
				}
			}

			roster := make([]dto.RosterMemberInfo, len(team.Roster))
			for j, member := range team.Roster {
				roster[j] = dto.RosterMemberInfo{
					ID:       member.ID,
					Name:     member.Name,
					Email:    member.Email,
					Country:  member.Country,
					PhotoURL: member.PhotoURL,
					Points:   member.Points,
					Wins:     member.Wins,
					Losses:   member.Losses,
					Assists:  member.Assists,
					Rebounds: member.Rebounds,
					Steals:   member.Steals,
				}
			}
			response.Roster = &roster

			result[i] = response
		}

		responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
		return
	}

	responseHandlers.RespondWithError(w, errLib.New("Role not supported for team access", http.StatusForbidden))
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
		ID:         team.ID,
		Name:       team.TeamDetails.Name,
		Capacity:   team.TeamDetails.Capacity,
		LogoURL:    team.TeamDetails.LogoURL,
		IsExternal: team.IsExternal,
		CreatedAt:  team.CreatedAt,
		UpdatedAt:  team.UpdatedAt,
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
			PhotoURL: member.PhotoURL,
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
// @Param team body dto.RequestDto true "Team details including name, capacity, coach_id, and optional logo_url"
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

// CreateExternalTeam creates a new external/opponent team.
// @Summary Create external team
// @Description Creates a new external team (opponent team not belonging to RISE). These teams are shared across all coaches.
// @Tags teams
// @Accept json
// @Produce json
// @Param team body dto.ExternalTeamRequestDto true "External team details (name, capacity, optional logo_url)"
// @Security Bearer
// @Success 201 {object} map[string]interface{} "External team created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 409 {object} map[string]interface{} "Conflict: Team name already exists"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /teams/external [post]
func (h *Handler) CreateExternalTeam(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.ExternalTeamRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	teamCreate, err := requestDto.ToCreateValueObjects()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.CreateExternalTeam(r.Context(), teamCreate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, map[string]string{
		"message": "External team created successfully. You can now use this team when scheduling games.",
	}, http.StatusCreated)
}

// GetExternalTeams retrieves all external/opponent teams.
// @Summary Get external teams
// @Description Retrieves all external teams (opponent teams). Useful for selecting opponents when creating games.
// @Tags teams
// @Accept json
// @Produce json
// @Success 200 {array} dto.Response "External teams retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /teams/external [get]
func (h *Handler) GetExternalTeams(w http.ResponseWriter, r *http.Request) {

	teams, err := h.Service.GetExternalTeams(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.Response, len(teams))

	for i, team := range teams {
		response := dto.Response{
			ID:         team.ID,
			Name:       team.TeamDetails.Name,
			Capacity:   team.TeamDetails.Capacity,
			LogoURL:    team.TeamDetails.LogoURL,
			IsExternal: true,
			CreatedAt:  team.CreatedAt,
			UpdatedAt:  team.UpdatedAt,
		}

		result[i] = response
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// SearchTeams searches for teams by name.
// @Summary Search teams
// @Description Searches for teams (both RISE and external) by name. Useful for autocomplete when creating games.
// @Tags teams
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Max results (default 20, max 50)"
// @Success 200 {array} dto.Response "Teams matching search query"
// @Failure 400 {object} map[string]interface{} "Bad Request: Missing query parameter"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /teams/search [get]
func (h *Handler) SearchTeams(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		responseHandlers.RespondWithError(w, errLib.New("Search query 'q' is required", http.StatusBadRequest))
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := int32(20) // default
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(parsedLimit)
		}
	}

	teams, err := h.Service.SearchTeamsByName(r.Context(), query, limit)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.Response, len(teams))

	for i, team := range teams {
		response := dto.Response{
			ID:         team.ID,
			Name:       team.TeamDetails.Name,
			Capacity:   team.TeamDetails.Capacity,
			LogoURL:    team.TeamDetails.LogoURL,
			IsExternal: team.IsExternal,
			CreatedAt:  team.CreatedAt,
			UpdatedAt:  team.UpdatedAt,
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

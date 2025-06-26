package game

import (
	"api/internal/di"
	dto "api/internal/domains/game/dto"
	service "api/internal/domains/game/services"
	"api/internal/domains/game/values"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"
	"strconv"

	contextUtils "api/utils/context"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	Service *service.Service
}

func NewHandler(container *di.Container) *Handler {
	return &Handler{Service: service.NewService(container)}
}

// CreateGame creates a new game record.
// @Summary Create a new game
// @Description Creates a new game entry in the system.
// @Tags games
// @Accept json
// @Produce json
// @Param game body dto.RequestDto true "Game details"
// @Security Bearer
// @Success 201 {object} nil "Game created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /games [post]
func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	value, err := requestDto.ToCreateGameValue()
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.CreateGame(r.Context(), value); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetGameById fetches a single game by ID.
// @Summary Get game by ID
// @Description Retrieves a single game using its UUID.
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID"
// @Success 200 {object} dto.ResponseDto "Game retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Game not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /games/{id} [get]
func (h *Handler) GetGameById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	game, err := h.Service.GetGameById(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.NewGameResponse(game)
	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetGames returns games, optionally filtered by 'upcoming' or 'past'.
// @Summary List games (all, upcoming, or past)
// @Description Retrieves a list of games with optional time-based filtering.
// @Tags games
// @Accept json
// @Produce json
// @Param filter query string false "Filter by time: upcoming or past"
// @Success 200 {array} dto.ResponseDto "List of games"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /games [get]
func (h *Handler) GetGames(w http.ResponseWriter, r *http.Request) {
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

	filter := query.Get("filter")

	var courtID, locationID uuid.UUID
	if val := query.Get("court_id"); val != "" {
		id, err := validators.ParseUUID(val)
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
		courtID = id
	}
	if val := query.Get("location_id"); val != "" {
		id, err := validators.ParseUUID(val)
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
		locationID = id
	}
	var games []values.ReadGameValue

	var err *errLib.CommonError
		gameFilter := values.GetGamesFilter{
		CourtID:    courtID,
		LocationID: locationID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	}
	switch filter {
	case "upcoming":
		games, err = h.Service.GetUpcomingGames(r.Context(), int32(limit), int32(offset))
	case "past":
		games, err = h.Service.GetPastGames(r.Context(), int32(limit), int32(offset))
	default:
		games, err = h.Service.GetGames(r.Context(), gameFilter)
	}

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.ResponseDto, len(games))
	for i, game := range games {
		result[i] = dto.NewGameResponse(game)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// GetMyGames returns games associated with the authenticated user's team.
// Only coaches and athletes are supported. The user's team is derived from
// their role and used to filter games.
// @Tags games
// @Security Bearer
// @Produce json
// @Success 200 {array} dto.ResponseDto "List of games for the current user"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /secure/games [get]
// GetMyGames returns games associated with the authenticated user's team.
// Only coaches and athletes are supported. The user's team is derived from
// their role and used to filter games.
// @Tags games
// @Security Bearer
// @Produce json
// @Success 200 {array} dto.ResponseDto "List of games for the current user"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /secure/games [get]
func (h *Handler) GetRoleGames(w http.ResponseWriter, r *http.Request) {
	role, err := contextUtils.GetUserRole(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Admins can view all games 
	if role == contextUtils.RoleAdmin || role == contextUtils.RoleSuperAdmin {
		h.GetGames(w, r)
		return
	}
	// Coaches and athletes can view games related to their teams
	userID, err := contextUtils.GetUserID(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit
	// Fetch games based on user role
	// Coaches see games for teams they coach, athletes see games for their team
	games, errC := h.Service.GetUserGames(r.Context(), userID, role, int32(limit), int32(offset))
	if errC != nil {
		responseHandlers.RespondWithError(w, errC)
		return
	}

	result := make([]dto.ResponseDto, len(games))
	for i, game := range games {
		result[i] = dto.NewGameResponse(game)
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// UpdateGame updates an existing game.
// @Summary Update a game
// @Description Updates a game by its ID.
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID"
// @Param game body dto.RequestDto true "Updated game details"
// @Security Bearer
// @Success 204 "Game updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Game not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /games/{id} [put]
func (h *Handler) UpdateGame(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	var requestDto dto.RequestDto
	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	value, err := requestDto.ToUpdateGameValue(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.UpdateGame(r.Context(), value); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteGame removes a game by ID.
// @Summary Delete a game
// @Description Deletes a game by its ID.
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID"
// @Security Bearer
// @Success 204 "Game deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Game not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /games/{id} [delete]
func (h *Handler) DeleteGame(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.DeleteGame(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

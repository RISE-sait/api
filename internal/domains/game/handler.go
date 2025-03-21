package game

import (
	dto "api/internal/domains/game/dto"
	repository "api/internal/domains/game/persistence"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	Repo *repository.Repository
}

func NewHandler(repo *repository.Repository) *Handler {
	return &Handler{Repo: repo}
}

// CreateGame creates a new game.
// @Summary Create a new game
// @Description Create a new game
// @Tags games
// @Accept json
// @Produce json
// @Param game body dto.RequestDto true "Game details"
// @Security Bearer
// @Success 201 {object} game.ResponseDto "Game created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /games [post]
func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	name, err := requestDto.ToCreateGameName()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	createdGame, err := h.Repo.CreateGame(r.Context(), name)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.NewGameResponse(createdGame)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusCreated)
}

// GetGameById retrieves a game by ID.
// @Summary Get a game by ID
// @Description Get a game by ID
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID"
// @Success 200 {object} game.ResponseDto "Game retrieved successfully"
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

	game, err := h.Repo.GetGameById(r.Context(), id)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.NewGameResponse(game)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetGames retrieves a list of games.
// @Summary Get a list of games
// @Description Get a list of games
// @Tags games
// @Accept json
// @Produce json
// @Param name query string false "Filter by game name"
// @Success 200 {array} game.ResponseDto "List of games retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /games [get]
func (h *Handler) GetGames(w http.ResponseWriter, r *http.Request) {

	games, err := h.Repo.GetGames(r.Context())
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

// UpdateGame updates an existing game.
// @Summary Update a game
// @Description Update a game
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID"
// @Param game body dto.RequestDto true "Game details"
// @Security Bearer
// @Success 200 {object} game.ResponseDto "Game updated successfully"
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

	gameToUpdate, err := requestDto.ToUpdateGameValue(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	updatedGame, err := h.Repo.UpdateGame(r.Context(), gameToUpdate)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := dto.NewGameResponse(updatedGame)

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusNoContent)
}

// DeleteGame deletes a game by ID.
// @Summary Delete a game
// @Description Delete a game by ID
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID"
// @Security Bearer
// @Success 204 "No Content: Game deleted successfully"
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

	if err = h.Repo.DeleteGame(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

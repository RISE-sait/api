package game

import (
	"api/internal/di"
	dto "api/internal/domains/game/dto"
	service "api/internal/domains/game/services"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	Service *service.Service
}

func NewHandler(container *di.Container) *Handler {
	return &Handler{Service: service.NewService(container)}
}

// CreateGame creates a new game.
// @Tags games
// @Accept json
// @Produce json
// @Param game body dto.RequestDto true "Game details"
// @Security Bearer
// @Success 201 "Game created successfully"
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

	if err = h.Service.CreateGame(r.Context(), name); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetGameById retrieves a game by ID.
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

	game, err := h.Service.GetGameById(r.Context(), id)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.NewGameResponse(game)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// GetGames retrieves a list of games.
// @Tags games
// @Accept json
// @Produce json
// @Success 200 {array} game.ResponseDto "List of games retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /games [get]
func (h *Handler) GetGames(w http.ResponseWriter, r *http.Request) {

	games, err := h.Service.GetGames(r.Context())
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
// @Tags games
// @Accept json
// @Produce json
// @Param id path string true "Game ID"
// @Param game body dto.RequestDto true "Game details"
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

	gameToUpdate, err := requestDto.ToUpdateGameValue(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Service.UpdateGame(r.Context(), gameToUpdate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteGame deletes a game by ID.
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

	if err = h.Service.DeleteGame(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

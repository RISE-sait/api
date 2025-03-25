package practice

import (
	dto "api/internal/domains/program/dto"
	repository "api/internal/domains/program/persistence"
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

// CreateProgram creates a new program.
// @Summary Create a new program
// @Description Create a new program
// @Tags programs
// @Accept json
// @Produce json
// @Param program body dto.RequestDto true "Program details"
// @Security Bearer
// @Success 201 {object} map[string]interface{} "Program created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /programs [post]
func (h *Handler) CreateProgram(w http.ResponseWriter, r *http.Request) {
	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	programCreate, err := requestDto.ToCreateValueObjects()

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.Create(r.Context(), programCreate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetPrograms retrieves a list of programs.
// @Summary Get a list of programs
// @Description Get a list of programs
// @Tags programs
// @Param type query string false "Program Type (game, practice, course, others)"
// @Accept json
// @Produce json
// @Success 200 {array} dto.Response "Programs retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /programs [get]
func (h *Handler) GetPrograms(w http.ResponseWriter, r *http.Request) {

	programs, err := h.Repo.List(r.Context(), r.URL.Query().Get("type"))

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.Response, len(programs))

	for i, program := range programs {
		result[i] = dto.Response{
			ID:          program.ID,
			Name:        program.ProgramDetails.Name,
			Description: program.ProgramDetails.Description,
			Level:       program.ProgramDetails.Level,
			Type:        program.ProgramDetails.Type,
			CreatedAt:   program.CreatedAt,
			UpdatedAt:   program.UpdatedAt,
		}
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// GetProgramLevels retrieves available program levels.
// @Description Retrieves a list of available program levels.
// @Tags programs
// @Accept json
// @Produce json
// @Success 200 {array} dto.LevelsResponse "Get program levels retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /programs/levels [get]
func (h *Handler) GetProgramLevels(w http.ResponseWriter, _ *http.Request) {
	levels := h.Repo.GetProgramLevels()

	response := dto.LevelsResponse{Name: levels}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// UpdateProgram updates an existing program.
// @Summary Update a program
// @Description Update a program
// @Tags programs
// @Accept json
// @Produce json
// @Param id path string true "Program ID"
// @Param program body dto.RequestDto true "Program details"
// @Success 204 "No Content: Program updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Program not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /programs/{id} [put]
func (h *Handler) UpdateProgram(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	var requestDto dto.RequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	programUpdate, err := requestDto.ToUpdateValueObjects(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.Repo.Update(r.Context(), programUpdate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

// DeleteProgram deletes a program by ID.
// @Tags programs
// @Accept json
// @Produce json
// @Param id path string true "Program ID"
// @Security Bearer
// @Success 204 "No Content: Program deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Program not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /programs/{id} [delete]
func (h *Handler) DeleteProgram(w http.ResponseWriter, r *http.Request) {
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

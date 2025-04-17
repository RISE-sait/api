package program

import (
	"api/internal/di"
	dto "api/internal/domains/program/dto"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	Service *Service
}

func NewHandler(container *di.Container) *Handler {
	return &Handler{Service: NewProgramService(container)}
}

// CreateProgram creates a new program.
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

	if err = h.Service.CreateProgram(r.Context(), programCreate); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

// GetPrograms retrieves a list of programs.
// @Tags programs
// @Param type query string false "Program Type (game, practice, course, others)"
// @Accept json
// @Produce json
// @Success 200 {array} dto.Response "Programs retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /programs [get]
func (h *Handler) GetPrograms(w http.ResponseWriter, r *http.Request) {

	programs, err := h.Service.GetPrograms(r.Context(), r.URL.Query().Get("type"))

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.Response, len(programs))

	for i, program := range programs {
		response := dto.Response{
			ID:          program.ID,
			Name:        program.ProgramDetails.Name,
			Description: program.ProgramDetails.Description,
			Level:       program.ProgramDetails.Level,
			Type:        program.ProgramDetails.Type,
			CreatedAt:   program.CreatedAt,
			UpdatedAt:   program.UpdatedAt,
		}

		if program.ProgramDetails.Capacity != nil {
			response.Capacity = program.ProgramDetails.Capacity
		}

		result[i] = response
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// GetProgram retrieves a program by ID.
// @Tags programs
// @Param id path string true "Program ID"
// @Accept json
// @Produce json
// @Success 200 {array} dto.Response "Programs retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Program not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /programs/{id} [get]
func (h *Handler) GetProgram(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	program, err := h.Service.GetProgram(r.Context(), id)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := dto.Response{
		ID:          program.ID,
		Name:        program.ProgramDetails.Name,
		Description: program.ProgramDetails.Description,
		Level:       program.ProgramDetails.Level,
		Type:        program.ProgramDetails.Type,
		CreatedAt:   program.CreatedAt,
		UpdatedAt:   program.UpdatedAt,
	}

	if program.ProgramDetails.Capacity != nil {
		result.Capacity = program.ProgramDetails.Capacity
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
}

// GetProgramLevels retrieves available program levels.
// @Tags programs
// @Accept json
// @Produce json
// @Success 200 {array} dto.LevelsResponse "Get program levels retrieved successfully"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /programs/levels [get]
func (h *Handler) GetProgramLevels(w http.ResponseWriter, _ *http.Request) {
	levels := h.Service.GetProgramLevels()

	response := dto.LevelsResponse{Name: levels}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// UpdateProgram updates an existing program.
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

	if err = h.Service.UpdateProgram(r.Context(), programUpdate); err != nil {
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

	if err = h.Service.DeleteProgram(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

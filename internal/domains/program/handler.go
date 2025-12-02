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
// @Success 201 {object} dto.Response "Program created successfully"
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

	program, err := h.Service.CreateProgram(r.Context(), programCreate)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := dto.Response{
		ID:          program.ID,
		Name:        program.ProgramDetails.Name,
		Description: program.ProgramDetails.Description,
		Type:        program.ProgramDetails.Type,
		CreatedAt:   program.CreatedAt,
		UpdatedAt:   program.UpdatedAt,
	}

	if program.ProgramDetails.Capacity != nil {
		result.Capacity = program.ProgramDetails.Capacity
	}

	if program.ProgramDetails.PhotoURL != nil {
		result.PhotoURL = program.ProgramDetails.PhotoURL
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusCreated)
}

// GetPrograms retrieves a list of programs.
// @Tags programs
// @Param type query string false "Program Type (practice, course, game, other, others, tournament, event, tryouts)"
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
			Type:        program.ProgramDetails.Type,
			CreatedAt:   program.CreatedAt,
			UpdatedAt:   program.UpdatedAt,
		}

		if program.ProgramDetails.Capacity != nil {
			response.Capacity = program.ProgramDetails.Capacity
		}

		if program.ProgramDetails.PhotoURL != nil {
			response.PhotoURL = program.ProgramDetails.PhotoURL
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
		Type:        program.ProgramDetails.Type,
		CreatedAt:   program.CreatedAt,
		UpdatedAt:   program.UpdatedAt,
	}

	if program.ProgramDetails.Capacity != nil {
		result.Capacity = program.ProgramDetails.Capacity
	}

	if program.ProgramDetails.PhotoURL != nil {
		result.PhotoURL = program.ProgramDetails.PhotoURL
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)
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

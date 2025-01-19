package controllers

// import (
// 	dto "api/internal/dtos/schedule"
// 	"api/internal/repositories"
// 	"api/internal/utils"
// 	"api/internal/utils/validators"
// 	"net/http"

// 	"github.com/go-chi/chi"
// )

// // SchedulesController provides HTTP handlers for managing schedules.
// type SchedulesController struct {
// 	SchedulesRepository *repositories.SchedulesRepository
// }

// // NewController creates a new instance of SchedulesController.
// func NewSchedulesController(SchedulesRepository *repositories.SchedulesRepository) *SchedulesController {
// 	return &SchedulesController{SchedulesRepository: SchedulesRepository}
// }

// // GetAllSchedules retrieves all schedules from the database.
// func (c *SchedulesController) GetAllSchedules(w http.ResponseWriter, r *http.Request) {
// 	schedules, err := c.SchedulesRepository.GetAllSchedules(r.Context())
// 	if err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	utils.RespondWithSuccess(w, schedules, http.StatusOK)
// }

// // GetScheduleByID retrieves a single schedule by its ID.
// func (c *SchedulesController) GetScheduleByID(w http.ResponseWriter, r *http.Request) {

// 	idStr := chi.URLParam(r, "id")

// 	id, err := validators.ParseUUID(idStr)

// 	if err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	schedule, err := c.SchedulesRepository.GetSchedule(r.Context(), id)

// 	if err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	utils.RespondWithSuccess(w, schedule, http.StatusOK)
// }

// func (c *SchedulesController) CreateSchedule(w http.ResponseWriter, r *http.Request) {

// 	var targetBody dto.CreateScheduleRequest

// 	if err := validators.DecodeRequestBody(r.Body, &targetBody); err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	if err := validators.ValidateDto(&targetBody); err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	err, params := targetBody.ToDBParams()

// 	if err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	if err := c.SchedulesRepository.CreateSchedule(r.Context(), params); err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	utils.RespondWithSuccess(w, nil, http.StatusCreated)
// }

// func (c *SchedulesController) UpdateSchedule(w http.ResponseWriter, r *http.Request) {

// 	var targetBody dto.UpdateScheduleRequest

// 	if err := validators.DecodeRequestBody(r.Body, &targetBody); err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	if err := validators.ValidateDto(&targetBody); err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	err, params := targetBody.ToDBParams()

// 	if err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	if err := c.SchedulesRepository.UpdateSchedule(r.Context(), params); err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
// }

// func (c *SchedulesController) DeleteSchedule(w http.ResponseWriter, r *http.Request) {

// 	idStr := chi.URLParam(r, "id")

// 	id, err := validators.ParseUUID(idStr)

// 	if err != nil {
// 		utils.RespondWithError(w, err)
// 	}

// 	if err = c.SchedulesRepository.DeleteSchedule(r.Context(), id); err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
// }

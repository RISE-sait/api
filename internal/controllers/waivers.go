package controllers

import (
	dto "api/internal/dtos/waiver"
	"api/internal/repositories"
	"api/internal/utils"
	"api/internal/utils/validators"
	db "api/sqlc"
	"net/http"
)

// SchedulesController provides HTTP handlers for managing schedules.
type WaiversController struct {
	WaiverRepository *repositories.WaiverRepository
}

// NewController creates a new instance of SchedulesController.
func NewWaiversController(waiverRepository *repositories.WaiverRepository) *WaiversController {
	return &WaiversController{WaiverRepository: waiverRepository}
}

func (c *WaiversController) GetAllUniqueWaivers(w http.ResponseWriter, r *http.Request) {
	waivers, err := c.WaiverRepository.GetAllUniqueWaivers(r.Context())
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, waivers, http.StatusOK)
}

func (c *WaiversController) GetWaiverByEmailAndDocLink(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	docLink := r.URL.Query().Get("doc_link")

	if email == "" || docLink == "" {
		http.Error(w, "Email and document link are required", http.StatusBadRequest)
		return
	}

	params := db.GetWaiverByEmailAndDocLinkParams{
		Email:        email,
		DocumentLink: docLink,
	}

	waiver, err := c.WaiverRepository.GetWaiverByEmailAndDocLink(r.Context(), &params)
	if err != nil {
		utils.RespondWithError(w, err)
	}

	utils.RespondWithSuccess(w, waiver, http.StatusOK)
}

// UpdateWaiverStatus updates the signed status of a waiver by email.
func (c *WaiversController) UpdateWaiverStatus(w http.ResponseWriter, r *http.Request) {

	var targetBody dto.UpdateWaiverRequest

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	params := targetBody.ToDBParams()

	if err := c.WaiverRepository.UpdateWaiverStatus(r.Context(), params); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, nil, http.StatusNoContent)
}

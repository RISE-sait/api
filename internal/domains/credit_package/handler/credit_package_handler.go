package handler

import (
	"net/http"

	"api/internal/di"
	dto "api/internal/domains/credit_package/dto"
	service "api/internal/domains/credit_package/service"
	paymentDto "api/internal/domains/payment/dto"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"

	"github.com/go-chi/chi"
)

type CreditPackageHandler struct {
	Service *service.CreditPackageService
}

func NewCreditPackageHandler(container *di.Container) *CreditPackageHandler {
	return &CreditPackageHandler{
		Service: service.NewCreditPackageService(container),
	}
}

// GetAllCreditPackages returns all available credit packages
// @Tags credit-packages
// @Accept json
// @Produce json
// @Success 200 {array} dto.CreditPackageResponse "List of credit packages"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /credit_packages [get]
func (h *CreditPackageHandler) GetAllCreditPackages(w http.ResponseWriter, r *http.Request) {
	packages, err := h.Service.GetAllPackages(r.Context())
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, packages, http.StatusOK)
}

// GetCreditPackageByID returns a specific credit package
// @Tags credit-packages
// @Accept json
// @Produce json
// @Param id path string true "Credit Package ID" Format(uuid)
// @Success 200 {object} dto.CreditPackageResponse "Credit package details"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Package not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /credit_packages/{id} [get]
func (h *CreditPackageHandler) GetCreditPackageByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := validators.ParseUUID(idStr)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	pkg, err := h.Service.GetPackageByID(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, pkg, http.StatusOK)
}

// CheckoutCreditPackage creates a Stripe checkout session for purchasing a credit package
// @Tags payments
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Credit Package ID" Format(uuid)
// @Success 200 {object} paymentDto.CheckoutResponseDto "Checkout URL generated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID or customer has remaining credits"
// @Failure 404 {object} map[string]interface{} "Not Found: Package not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /checkout/credit_packages/{id} [post]
func (h *CreditPackageHandler) CheckoutCreditPackage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, parseErr := validators.ParseUUID(idStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	checkoutURL, err := h.Service.CheckoutCreditPackage(r.Context(), id)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := paymentDto.CheckoutResponseDto{
		PaymentURL: checkoutURL,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)
}

// Admin CRUD handlers

// CreateCreditPackage creates a new credit package
// @Tags admin-credit-packages
// @Accept json
// @Produce json
// @Security Bearer
// @Param package body dto.CreateCreditPackageRequest true "Credit package details"
// @Success 201 {object} dto.CreditPackageResponse "Credit package created successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /credit_packages [post]
func (h *CreditPackageHandler) CreateCreditPackage(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateCreditPackageRequest

	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := validators.ValidateDto(&req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	pkg, err := h.Service.CreatePackage(r.Context(), req)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, pkg, http.StatusCreated)
}

// UpdateCreditPackage updates an existing credit package
// @Tags admin-credit-packages
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Credit Package ID" Format(uuid)
// @Param package body dto.UpdateCreditPackageRequest true "Updated credit package details"
// @Success 200 {object} dto.CreditPackageResponse "Credit package updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input"
// @Failure 404 {object} map[string]interface{} "Not Found: Package not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /credit_packages/{id} [put]
func (h *CreditPackageHandler) UpdateCreditPackage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, parseErr := validators.ParseUUID(idStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	var req dto.UpdateCreditPackageRequest
	if err := validators.ParseJSON(r.Body, &req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err := validators.ValidateDto(&req); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	pkg, err := h.Service.UpdatePackage(r.Context(), id, req)
	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, pkg, http.StatusOK)
}

// DeleteCreditPackage deletes a credit package
// @Tags admin-credit-packages
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Credit Package ID" Format(uuid)
// @Success 204 "Credit package deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Package not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /credit_packages/{id} [delete]
func (h *CreditPackageHandler) DeleteCreditPackage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, parseErr := validators.ParseUUID(idStr)
	if parseErr != nil {
		responseHandlers.RespondWithError(w, parseErr)
		return
	}

	if err := h.Service.DeletePackage(r.Context(), id); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)
}

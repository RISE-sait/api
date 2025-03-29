package user

import (
	"api/internal/di"
	dto "api/internal/domains/user/dto/customer"
	customerRepo "api/internal/domains/user/persistence/repository"
	errLib "api/internal/libs/errors"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
	"strconv"
)

type CustomersHandler struct {
	CustomerRepo *customerRepo.CustomerRepository
}

func NewCustomersHandler(container *di.Container) *CustomersHandler {
	return &CustomersHandler{
		CustomerRepo: customerRepo.NewCustomerRepository(container.Queries.UserDb),
	}
}

// UpdateCustomerStats updates customer statistics based on the provided customer ID.
// @Tags customers
// @Accept json
// @Produce json
// @Security Bearer
// @Param customer_id path string true "Customer ID" // Customer ID to update stats for
// @Param update_body body customer.StatsUpdateRequestDto true "Customer stats update data"
// @Success 204 {object} map[string]interface{} "Customer stats updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{customer_id}/athlete [patch]
func (h *CustomersHandler) UpdateCustomerStats(w http.ResponseWriter, r *http.Request) {

	customerIdStr := chi.URLParam(r, "customer_id")

	var requestDto dto.StatsUpdateRequestDto

	if err := validators.ParseJSON(r.Body, &requestDto); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	details, err := requestDto.ToUpdateValue(customerIdStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if err = h.CustomerRepo.UpdateStats(r.Context(), details); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)

}

// GetCustomers retrieves a list of customers with optional filtering and pagination.
// @Summary Get customers
// @Description Retrieves a list of customers, optionally filtered by HubSpot IDs, with pagination support.
// @Tags customers
// @Accept json
// @Produce json
// @Param limit query int false "Number of customers to retrieve (default: 20)"
// @Param offset query int false "Number of customers to skip (default: 0)"
// @Param parent_id query string false "Parent ID to filter customers (example: 123e4567-e89b-12d3-a456-426614174000)"
// @Success 200 {array} customer.Response "List of customers"
// @Failure 400 "Bad Request: Invalid parameters"
// @Failure 500 "Internal Server Error"
// @Router /customers [get]
func (h *CustomersHandler) GetCustomers(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()

	maxLimit, offset := 20, 0

	var parentID uuid.UUID

	if limitStr := query.Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New(fmt.Sprintf("Error encountered parsing limit: %s", err.Error()), http.StatusBadRequest))
			return
		}
		if parsedLimit <= 0 {
			responseHandlers.RespondWithError(w, errLib.New("Limit must be greater than 0", http.StatusBadRequest))
			return
		}
		if parsedLimit > maxLimit {
			responseHandlers.RespondWithError(w, errLib.New(fmt.Sprintf("max limit is %d", maxLimit), http.StatusBadRequest))
			return
		}
		maxLimit = parsedLimit
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			responseHandlers.RespondWithError(w, errLib.New(fmt.Sprintf("Error encountered parsing offset: %s", err.Error()), http.StatusBadRequest))
			return
		}
		if parsedOffset < 0 {
			responseHandlers.RespondWithError(w, errLib.New("Offset must be at least 0", http.StatusBadRequest))
			return
		}
		offset = parsedOffset
	}

	if parentIdStr := query.Get("parent_id"); parentIdStr != "" {
		id, err := validators.ParseUUID(parentIdStr)
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
		parentID = id
	}

	dbCustomers, err := h.CustomerRepo.GetCustomers(r.Context(), int32(maxLimit), int32(offset), parentID)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.Response, len(dbCustomers))

	for i, customer := range dbCustomers {
		response := dto.UserReadValueToResponse(customer)

		result[i] = response
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)

}

// GetCustomerByID retrieves a customer by ID.
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} customer.Response "The customer"
// @Failure 400 "Bad Request: Invalid parameters"
// @Failure 500 "Internal Server Error"
// @Router /customers/id/{id} [get]
func (h *CustomersHandler) GetCustomerByID(w http.ResponseWriter, r *http.Request) {

	var id uuid.UUID

	if idStr := chi.URLParam(r, "id"); idStr != "" {
		tempId, err := validators.ParseUUID(idStr)
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}
		id = tempId
	}

	customer, err := h.CustomerRepo.GetCustomer(r.Context(), id, "")

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.UserReadValueToResponse(customer)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)

}

// GetCustomerByEmail retrieves a customer by email
// @Tags customers
// @Accept json
// @Produce json
// @Param email path string true "Customer Email"
// @Success 200 {object} customer.Response "The customer"
// @Failure 400 "Bad Request: Invalid parameters"
// @Failure 500 "Internal Server Error"
// @Router /customers/email/{email} [get]
func (h *CustomersHandler) GetCustomerByEmail(w http.ResponseWriter, r *http.Request) {

	email := chi.URLParam(r, "email")

	customer, err := h.CustomerRepo.GetCustomer(r.Context(), uuid.Nil, email)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.UserReadValueToResponse(customer)

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)

}

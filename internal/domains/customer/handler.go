package user

import (
	"api/internal/di"
	dto "api/internal/domains/customer/dto"
	repository "api/internal/domains/customer/persistence/repository"
	enrollmentRepo "api/internal/domains/enrollment/persistence/repository/enrollment"
	eventCapacityRepo "api/internal/domains/enrollment/persistence/repository/event_capacity"
	enrollmentService "api/internal/domains/enrollment/service"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"api/internal/services/hubspot"
	"github.com/go-chi/chi"
	"net/http"
	"strings"
)

type CustomersHandler struct {
	HubSpotService    *hubspot.Service
	CustomerRepo      repository.RepositoryInterface
	EnrollmentService *enrollmentService.EnrollmentService
}

func NewCustomersHandler(container *di.Container) *CustomersHandler {
	return &CustomersHandler{
		HubSpotService: container.HubspotService,
		EnrollmentService: enrollmentService.NewEnrollmentService(
			enrollmentRepo.NewEnrollmentRepository(container.Queries.EnrollmentDb),
			eventCapacityRepo.NewEventCapacityRepository(container.Queries.EnrollmentDb),
		),
		CustomerRepo: repository.NewCustomerRepository(container.Queries.CustomerDb),
	}
}

// UpdateCustomerStats updates customer statistics based on the provided customer ID.
// @Summary Update customer statistics
// @Description Updates customer statistics (wins, losses, etc.) for the specified customer ID
// @Tags customers
// @Accept json
// @Produce json
// @Param customer_id path string true "Customer ID" // Customer ID to update stats for
// @Param update_body body dto.StatsUpdateRequestDto true "Customer stats update data"
// @Success 204 {object} map[string]interface{} "Customer stats updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{customer_id}/stats [patch]
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

	if err := h.CustomerRepo.UpdateStats(r.Context(), details); err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseHandlers.RespondWithSuccess(w, nil, http.StatusNoContent)

}

// GetCustomers retrieves a list of customers with optional filtering by HubSpot IDs.
// @Summary Get customers
// @Description Retrieves a list of customers, optionally filtered by HubSpot IDs. Returns user details from the database and HubSpot.
// @Tags customers
// @Accept json
// @Produce json
// @Param hubspot_ids query string false "Comma-separated list of HubSpot IDs to filter customers"
// @Success 200 {array} dto.Response "List of customers"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers [get]
func (h *CustomersHandler) GetCustomers(w http.ResponseWriter, r *http.Request) {

	// get hubspot ids

	hubspotIdsStr := r.URL.Query().Get("hubspot_ids")

	var hubspotIds []string = nil

	if hubspotIdsStr != "" {
		hubspotIds = strings.Split(hubspotIdsStr, ",")
	}

	dbCustomers, err := h.CustomerRepo.GetCustomers(r.Context(), hubspotIds)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	if hubspotIds == nil {
		ids := make([]string, len(dbCustomers))

		for i, dbCustomer := range dbCustomers {
			ids[i] = dbCustomer.HubspotID
		}

		hubspotIds = ids
	}

	// get customers using hubspot ids

	hubspotCustomers, err := h.HubSpotService.GetUsersByIds(hubspotIds)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	responseBody := make([]dto.Response, len(dbCustomers))

	for i, dbCustomer := range dbCustomers {

		var hubspotCustomer hubspot.UserResponse

		for _, c := range hubspotCustomers {
			if c.HubSpotId == dbCustomer.HubspotID {
				hubspotCustomer = c
			}
		}

		customer := dto.Response{
			UserID:     dbCustomer.ID,
			HubspotId:  dbCustomer.HubspotID,
			ProfilePic: dbCustomer.ProfilePicUrl,
			FirstName:  hubspotCustomer.Properties.FirstName,
			LastName:   hubspotCustomer.Properties.LastName,
		}

		if hubspotCustomer.Properties.Email != "" {
			customer.Email = &hubspotCustomer.Properties.Email
		}

		responseBody[i] = customer
	}

	responseHandlers.RespondWithSuccess(w, responseBody, http.StatusOK)

}

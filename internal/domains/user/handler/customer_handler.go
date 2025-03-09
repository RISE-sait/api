package user

import (
	"api/internal/di"
	enrollmentRepo "api/internal/domains/enrollment/persistence"
	enrollmentService "api/internal/domains/enrollment/service"
	"api/internal/domains/user/dto/customer"
	customerRepo "api/internal/domains/user/persistence/repository/customer"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"api/internal/services/hubspot"
	"github.com/go-chi/chi"
	"net/http"
	"strings"
)

type CustomersHandler struct {
	HubspotService    *hubspot.Service
	CustomerRepo      *customerRepo.Repository
	EnrollmentService *enrollmentService.Service
}

func NewCustomersHandler(container *di.Container) *CustomersHandler {
	return &CustomersHandler{
		HubspotService: container.HubspotService,
		EnrollmentService: enrollmentService.NewEnrollmentService(
			enrollmentRepo.NewEnrollmentRepository(container.Queries.EnrollmentDb),
		),
		CustomerRepo: customerRepo.NewCustomerRepository(container.Queries.UserDb),
	}
}

// UpdateCustomerStats updates customer statistics based on the provided customer ID.
// @Summary Update customer statistics
// @Description Updates customer statistics (wins, losses, etc.) for the specified customer ID
// @Tags customers
// @Accept json
// @Produce json
// @Param customer_id path string true "Customer ID" // Customer ID to update stats for
// @Param update_body body customer.StatsUpdateRequestDto true "Customer stats update data"
// @Success 204 {object} map[string]interface{} "Customer stats updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{customer_id}/stats [patch]
func (h *CustomersHandler) UpdateCustomerStats(w http.ResponseWriter, r *http.Request) {

	customerIdStr := chi.URLParam(r, "customer_id")

	var requestDto customer.StatsUpdateRequestDto

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
// @Description Retrieves a list of customers, optionally filtered by HubSpot IDs.
// @Tags customers
// @Accept json
// @Produce json
// @Param hubspot_ids query string false "Comma-separated list of HubSpot IDs to filter customers"
// @Success 200 {array} customer.Response "List of customers"
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

	hubspotUsers, err := h.HubspotService.GetUsersByIds(hubspotIds)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]customer.Response, len(hubspotUsers))

	for i, staff := range dbCustomers {
		response := customer.Response{
			UserID:     staff.ID,
			HubspotId:  staff.HubspotID,
			ProfilePic: staff.ProfilePicUrl,
		}

		for _, user := range hubspotUsers {
			if user.HubSpotId == staff.HubspotID {
				response.FirstName = user.Properties.FirstName
				response.LastName = user.Properties.LastName

				if user.Properties.Email != "" {
					response.Email = &user.Properties.Email
				}
			}
		}

		result[i] = response
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)

}

// GetMembershipPlansByCustomer retrieves a list of membership plans for a specific customer.
// @Summary Get membership plans by customer
// @Description Retrieves a list of membership plans associated with a specific customer, using the customer ID as a required parameter.
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" // Customer ID is required as part of the URL path
// @Success 200 {array} customer.MembershipPlansResponseDto "List of membership plans for the customer"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid customer ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Customer not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{id}/membership-plans [get]
func (h *CustomersHandler) GetMembershipPlansByCustomer(w http.ResponseWriter, r *http.Request) {

	customerIDStr := chi.URLParam(r, "id")

	customerID, err := validators.ParseUUID(customerIDStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	dbPlans, err := h.CustomerRepo.GetMembershipPlansByCustomer(r.Context(), customerID)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]customer.MembershipPlansResponseDto, len(dbPlans))

	for i, dbPlan := range dbPlans {
		response := customer.MembershipPlansResponseDto{
			ID:               dbPlan.ID,
			CustomerID:       dbPlan.CustomerID,
			MembershipPlanID: dbPlan.MembershipPlanID,
			StartDate:        dbPlan.StartDate,
			RenewalDate:      dbPlan.RenewalDate,
			Status:           dbPlan.Status,
			CreatedAt:        dbPlan.CreatedAt,
			UpdatedAt:        dbPlan.UpdatedAt,
			MembershipName:   dbPlan.MembershipName,
		}

		result[i] = response
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)

}

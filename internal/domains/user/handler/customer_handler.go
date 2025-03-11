package user

import (
	"api/internal/di"
	enrollmentRepo "api/internal/domains/enrollment/persistence"
	enrollmentService "api/internal/domains/enrollment/service"
	dto "api/internal/domains/user/dto/customer"
	customerRepo "api/internal/domains/user/persistence/repository"
	responseHandlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"github.com/go-chi/chi"
	"net/http"
)

type CustomersHandler struct {
	CustomerRepo      *customerRepo.CustomerRepository
	EnrollmentService *enrollmentService.Service
}

func NewCustomersHandler(container *di.Container) *CustomersHandler {
	return &CustomersHandler{
		EnrollmentService: enrollmentService.NewEnrollmentService(
			enrollmentRepo.NewEnrollmentRepository(container.Queries.EnrollmentDb),
		),
		CustomerRepo: customerRepo.NewCustomerRepository(container.Queries.UserDb),
	}
}

// GetAthleteInfo retrieves customer statistics based on the provided customer ID.
// @Summary Get customer statistics
// @Description Fetches customer statistics (wins, losses, etc.) for the specified customer ID.
// @Tags customers
// @Accept json
// @Produce json
// @Param customer_id path string true "Customer ID" // Customer ID to fetch stats for
// @Success 200 {object} dto.AthleteResponseDto "Customer stats retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid parameters"
// @Failure 404 {object} map[string]interface{} "Not Found: Customer does not exist"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{customer_id}/athlete [get]
func (h *CustomersHandler) GetAthleteInfo(w http.ResponseWriter, r *http.Request) {

	customerIdStr := chi.URLParam(r, "customer_id")

	customerId, err := validators.ParseUUID(customerIdStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	info, err := h.CustomerRepo.GetAthleteInfo(r.Context(), customerId)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := dto.AthleteResponseDto{
		ID:         info.ID,
		ProfilePic: info.ProfilePicUrl,
		Wins:       info.Wins,
		Losses:     info.Losses,
		Points:     info.Points,
		Steals:     info.Steals,
		Assists:    info.Assists,
		Rebounds:   info.Rebounds,
		CreatedAt:  info.CreatedAt,
		UpdatedAt:  info.UpdatedAt,
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusNoContent)

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

// GetCustomers retrieves a list of customers with optional filtering by HubSpot IDs.
// @Summary Get customers
// @Description Retrieves a list of customers, optionally filtered by HubSpot IDs.
// @Tags customers
// @Accept json
// @Produce json
// @Success 200 {array} customer.Response "List of customers"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers [get]
func (h *CustomersHandler) GetCustomers(w http.ResponseWriter, r *http.Request) {

	dbCustomers, err := h.CustomerRepo.GetCustomers(r.Context())

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.Response, len(dbCustomers))

	for i, customer := range dbCustomers {
		response := dto.Response{
			UserID:      customer.ID,
			FirstName:   customer.FirstName,
			LastName:    customer.LastName,
			Email:       customer.Email,
			Phone:       customer.Phone,
			CountryCode: customer.CountryCode,
			HubspotId:   customer.HubspotID,
			ProfilePic:  customer.ProfilePicUrl,
		}

		result[i] = response
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)

}

// GetChildrenByParentID retrieves a repository's children using the parent's ID.
// @Summary Get a repository's children by parent ID
// @Description Retrieves a repository's children using the parent's ID
// @Tags customers
// @Accept json
// @Produce json
// @Param email path string true "Parent ID"
// @Success 200 {array} hubspot.UserResponse "Customer's children retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid ID"
// @Failure 404 {object} map[string]interface{} "Not Found: Parent or children not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{id}/children [get]
func (h *CustomersHandler) GetChildrenByParentID(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := validators.ParseUUID(idStr)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	// Fetch repository's children from HubSpot
	children, err := h.CustomerRepo.GetChildrenByCustomerID(r.Context(), id)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var childrenResponse []dto.Response

	for _, child := range children {

		response := dto.Response{
			UserID:      child.ID,
			FirstName:   child.FirstName,
			LastName:    child.LastName,
			Email:       child.Email,
			Age:         child.Age,
			Phone:       child.Phone,
			HubspotId:   child.HubspotID,
			CountryCode: child.CountryCode,
			ProfilePic:  child.ProfilePicUrl,
		}

		childrenResponse = append(childrenResponse, response)
	}

	responseHandlers.RespondWithSuccess(w, children, http.StatusOK)
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

	plans, err := h.CustomerRepo.GetMembershipPlansByCustomer(r.Context(), customerID)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	result := make([]dto.MembershipPlansResponseDto, len(plans))

	for i, plan := range plans {
		response := dto.MembershipPlansResponseDto{
			ID:               plan.ID,
			CustomerID:       plan.CustomerID,
			MembershipPlanID: plan.MembershipPlanID,
			StartDate:        plan.StartDate,
			RenewalDate:      plan.RenewalDate,
			Status:           plan.Status,
			CreatedAt:        plan.CreatedAt,
			UpdatedAt:        plan.UpdatedAt,
			MembershipName:   plan.MembershipName,
		}

		result[i] = response
	}

	responseHandlers.RespondWithSuccess(w, result, http.StatusOK)

}

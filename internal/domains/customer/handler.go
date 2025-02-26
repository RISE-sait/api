package customer

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
	"github.com/google/uuid"
	"net/http"
)

type CustomersHandler struct {
	HubSpotService    *hubspot.Service
	UserRepo          repository.RepositoryInterface
	EnrollmentService *enrollmentService.EnrollmentService
}

func NewCustomersHandler(container *di.Container) *CustomersHandler {
	return &CustomersHandler{
		HubSpotService: container.HubspotService,
		EnrollmentService: enrollmentService.NewEnrollmentService(
			enrollmentRepo.NewEnrollmentRepository(container.Queries.EnrollmentDb),
			eventCapacityRepo.NewEventCapacityRepository(container.Queries.EnrollmentDb),
		),
		UserRepo: repository.NewUserRepository(container.Queries.CustomerDb),
	}
}

// GetCustomerByEmail retrieves a repository by email.
// @Summary Get a repository by email
// @Description Retrieves a repository using their email address
// @Tags customers
// @Accept json
// @Produce json
// @Param email path string true "Email"
// @Success 200 {object} hubspot.UserResponse "Customer retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid Email"
// @Failure 404 {object} map[string]interface{} "Not Found: Customer not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{email} [get]
func (h *CustomersHandler) GetCustomerByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	user, err := h.HubSpotService.GetUserByEmail(email)

	if err != nil && err.HTTPCode == http.StatusInternalServerError {
		responseHandlers.RespondWithError(w, err)
		return
	}

	var customerInfo dto.Response

	if user != nil {

		userId, err := h.UserRepo.GetUserIDByHubSpotId(r.Context(), user.HubSpotId)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
		}

		customerInfo = dto.Response{
			HubspotId: user.HubSpotId,
			Email:     user.Properties.Email,
			FirstName: user.Properties.FirstName,
			LastName:  user.Properties.LastName,
			UserID:    *userId,
		}
	}

	responseHandlers.RespondWithSuccess(w, customerInfo, http.StatusOK)
}

// GetChildrenByParentEmail retrieves a repository's children using the parent's email.
// @Summary Get a repository's children by parent email
// @Description Retrieves a repository's children using the parent's email address
// @Tags customers
// @Accept json
// @Produce json
// @Param email path string true "Parent Email"
// @Success 200 {array} hubspot.UserResponse "Customer's children retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid Email"
// @Failure 404 {object} map[string]interface{} "Not Found: Parent or children not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers/{email}/children [get]
func (h *CustomersHandler) GetChildrenByParentEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	// Fetch repository's children from HubSpot
	customer, err := h.HubSpotService.GetUserByEmail(email)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	contacts := customer.Associations.Contact.Result

	var childrenIds []string

	// Map HubSpot response to DTO
	for _, contact := range contacts {

		if contact.Type == "child_parent" {

			childrenIds = append(childrenIds, contact.ID)
		}
	}

	children, err := h.HubSpotService.GetUsersByIds(childrenIds)

	if err != nil {
		responseHandlers.RespondWithError(w, err)
	}

	responseHandlers.RespondWithSuccess(w, children, http.StatusOK)
}

// GetCustomers retrieves customers by event HubSpotId or fetches all customers.
// @Summary Get customers
// @Description Retrieves customers based on an event HubSpotId or returns all customers
// @Tags customers
// @Accept json
// @Produce json
// @Param event_id query string false "Event HubSpotId (if specified, fetches customers for the event)"
// @Success 200 {array} hubspot.UserResponse "Customers retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /customers [get]
func (h *CustomersHandler) GetCustomers(w http.ResponseWriter, r *http.Request) {

	eventIdStr := r.URL.Query().Get("event_id")

	if eventIdStr != "" {
		var eventId uuid.UUID

		eventId, err := validators.ParseUUID(eventIdStr)

		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		enrolledCustomers, err := h.EnrollmentService.GetEnrollments(r.Context(), &eventId, nil)

		customerIds := make([]string, len(enrolledCustomers))
		for i, c := range enrolledCustomers {
			customerIds[i] = c.ID.String()
		}

		hubspotCustomers, err := h.HubSpotService.GetUsersByIds(customerIds)
		if err != nil {
			responseHandlers.RespondWithError(w, err)
			return
		}

		responseHandlers.RespondWithSuccess(w, hubspotCustomers, http.StatusOK)
		return
	}

	customers, err := h.HubSpotService.GetUsers("")

	if err != nil {
		responseHandlers.RespondWithError(w, err)
		return
	}

	response := make([]dto.Response, len(customers))

	for i, customer := range customers {

		response[i] = dto.Response{
			HubspotId: customer.HubSpotId,
			Email:     customer.Properties.Email,
			FirstName: customer.Properties.FirstName,
			LastName:  customer.Properties.LastName,
		}
	}

	responseHandlers.RespondWithSuccess(w, response, http.StatusOK)

}

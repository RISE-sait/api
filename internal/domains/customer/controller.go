package customer

import (
	"api/internal/di"
	"api/internal/domains/customer/dto"
	errLib "api/internal/libs/errors"
	response_handlers "api/internal/libs/responses"
	"api/internal/libs/validators"
	"api/internal/services/hubspot"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type CustomersController struct {
	HubSpotService   *hubspot.HubSpotService
	CustomersService *CustomersService
}

func NewCustomersController(container *di.Container) *CustomersController {
	return &CustomersController{HubSpotService: container.HubspotService, CustomersService: NewCustomersService(container)}
}

func (h *CustomersController) GetCustomerById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	customer, err := h.HubSpotService.GetCustomerById(id)
	if err != nil {
		response_handlers.RespondWithError(w, err)
	} else {
		response_handlers.RespondWithSuccess(w, customer, http.StatusOK)
	}
}

func (h *CustomersController) GetCustomerByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	customer, err := h.HubSpotService.GetCustomerByEmail(email)

	if err != nil {
		fmt.Println("Error: ", *err)
		response_handlers.RespondWithError(w, err)
	} else {
		response_handlers.RespondWithSuccess(w, *customer, http.StatusOK)
	}
}

func (h *CustomersController) GetCustomers(w http.ResponseWriter, r *http.Request) {

	eventIdStr := r.URL.Query().Get("event_id")

	var eventId uuid.UUID

	if eventIdStr != "" {
		id, err := validators.ParseUUID(eventIdStr)

		if err != nil {
			response_handlers.RespondWithError(w, err)
			return
		}

		eventId = id
	}

	customers, err := h.CustomersService.GetCustomers(r.Context(), eventId)

	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response := make([]dto.CustomerResponse, len(customers))

	for i, customer := range customers {

		response[i] = dto.CustomerResponse{
			CustomerID:            customer.CustomerID,
			Name:                  *customer.Name,
			Email:                 customer.Email,
			Membership:            customer.MembershipName,
			Attendance:            customer.Attendance,
			MembershipRenewalDate: customer.MembershipRenewalDate.Format(time.RFC3339),
		}
	}

	response_handlers.RespondWithSuccess(w, response, http.StatusOK)
}

func (h *CustomersController) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer hubspot.HubSpotCustomerCreateBody
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {

		newErr := errLib.New(err.Error(), http.StatusBadRequest)
		response_handlers.RespondWithError(w, newErr)
		return
	}

	// if err := h.HubSpotService.CreateCustomer(customer); err != nil {
	// 	response_CustomersControllers.RespondWithError(w, err)
	// 	return
	// }

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

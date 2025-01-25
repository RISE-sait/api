package customer

import (
	errLib "api/internal/libs/errors"
	response_handlers "api/internal/libs/responses"
	"api/internal/services/hubspot"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	HubSpotService *hubspot.HubSpotService
}

func NewHandler(HubSpotService *hubspot.HubSpotService) *Handler {
	return &Handler{HubSpotService: HubSpotService}
}

func (h *Handler) GetCustomerById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	customer, err := h.HubSpotService.GetCustomerById(id)
	if err != nil {
		response_handlers.RespondWithError(w, err)
	} else {
		response_handlers.RespondWithSuccess(w, customer, http.StatusOK)
	}
}

func (h *Handler) GetCustomerByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	customer, err := h.HubSpotService.GetCustomerByEmail(email)

	if err != nil {
		fmt.Println("Error: ", *err)
		response_handlers.RespondWithError(w, err)
	} else {
		response_handlers.RespondWithSuccess(w, *customer, http.StatusOK)
	}
}

func (h *Handler) GetCustomers(w http.ResponseWriter, _ *http.Request) {
	customers, err := h.HubSpotService.GetCustomers("")
	if err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}
	response_handlers.RespondWithSuccess(w, customers, http.StatusOK)
}

func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer hubspot.HubSpotCustomerCreateBody
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {

		newErr := errLib.New(err.Error(), http.StatusBadRequest)
		response_handlers.RespondWithError(w, newErr)
		return
	}

	if err := h.HubSpotService.CreateCustomer(customer); err != nil {
		response_handlers.RespondWithError(w, err)
		return
	}

	response_handlers.RespondWithSuccess(w, nil, http.StatusCreated)
}

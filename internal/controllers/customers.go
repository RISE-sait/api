package controllers

import (
	"api/internal/repositories"
	"api/internal/types/hubspot"
	"api/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type CustomersController struct {
	CustomerRepository *repositories.CustomerRepository
}

func NewCustomersController(customerRepository *repositories.CustomerRepository) *CustomersController {
	return &CustomersController{CustomerRepository: customerRepository}
}

func (c *CustomersController) GetCustomerById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	customer, err := c.CustomerRepository.GetCustomerById(id)
	if err != nil {
		utils.RespondWithError(w, err)
	} else {
		utils.RespondWithSuccess(w, customer, http.StatusOK)
	}
}

func (c *CustomersController) GetCustomerByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	customer, err := c.CustomerRepository.GetCustomerByEmail(email)

	if err != nil {
		fmt.Println("Error: ", *err)
		utils.RespondWithError(w, err)
	} else {
		utils.RespondWithSuccess(w, *customer, http.StatusOK)
	}
}

func (c *CustomersController) GetCustomers(w http.ResponseWriter, _ *http.Request) {
	customers, err := c.CustomerRepository.GetCustomers("")
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}
	utils.RespondWithSuccess(w, customers, http.StatusOK)
}

func (c *CustomersController) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer hubspot.HubSpotCustomerCreateBody
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		utils.RespondWithError(w, utils.CreateHTTPError(err.Error(), http.StatusBadRequest))
		return
	}

	createdCustomer, err := c.CustomerRepository.CreateCustomer(customer)
	if err != nil {
		utils.RespondWithError(w, err)
		return
	}

	utils.RespondWithSuccess(w, createdCustomer, http.StatusCreated)
}

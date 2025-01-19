package repositories

import (
	"api/internal/services"
	"api/internal/types/hubspot"
	"api/internal/utils"
	db "api/sqlc"
	"fmt"
	"net/http"
)

type CustomerRepository struct {
	HubSpotService *services.HubSpotService
	Queries        *db.Queries
}

func (r *CustomerRepository) GetCustomerById(id string) (*hubspot.HubSpotCustomerResponse, *utils.HTTPError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&properties=firstName,lastName,family_role,email", r.HubSpotService.BaseURL, id)
	return utils.ExecuteHubSpotRequest[hubspot.HubSpotCustomerResponse](r.HubSpotService, http.MethodGet, url, nil)
}

func (r *CustomerRepository) GetCustomerByEmail(email string) (*hubspot.HubSpotCustomerResponse, *utils.HTTPError) {

	fmt.Println("Email: ", email)
	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&idProperty=email&properties=firstName,lastName,family_role", r.HubSpotService.BaseURL, email)
	return utils.ExecuteHubSpotRequest[hubspot.HubSpotCustomerResponse](r.HubSpotService, http.MethodGet, url, nil)
}

func (r *CustomerRepository) CreateCustomer(customer hubspot.HubSpotCustomerCreateBody) (*hubspot.HubSpotCustomerCreateBody, *utils.HTTPError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts", r.HubSpotService.BaseURL)
	return utils.ExecuteHubSpotRequest[hubspot.HubSpotCustomerCreateBody](r.HubSpotService, http.MethodPost, url, customer)
}

func (r *CustomerRepository) GetCustomers(after string) ([]hubspot.HubSpotCustomerResponse, *utils.HTTPError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts?limit=10", r.HubSpotService.BaseURL)
	if after != "" {
		url += fmt.Sprintf("&after=%s", after)
	}

	hubSpotResponse, err := utils.ExecuteHubSpotRequest[hubspot.HubSpotCustomersResponse](r.HubSpotService, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return hubSpotResponse.Results, nil
}

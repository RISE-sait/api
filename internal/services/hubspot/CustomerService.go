package hubspot

import (
	"api/internal/types"
	"fmt"
	"net/http"
)

type CustomerService struct {
	HubSpotService *HubSpotService
}

func (r *CustomerService) GetCustomerById(id string) (*HubSpotCustomerResponse, *types.HTTPError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&properties=firstName,lastName,family_role,email", r.HubSpotService.BaseURL, id)
	return ExecuteHubSpotRequest[HubSpotCustomerResponse](r.HubSpotService, http.MethodGet, url, nil)
}

func (r *CustomerService) GetCustomerByEmail(email string) (*HubSpotCustomerResponse, *types.HTTPError) {

	fmt.Println("Email: ", email)
	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&idProperty=email&properties=firstName,lastName,family_role", r.HubSpotService.BaseURL, email)
	return ExecuteHubSpotRequest[HubSpotCustomerResponse](r.HubSpotService, http.MethodGet, url, nil)
}

func (r *CustomerService) CreateCustomer(customer HubSpotCustomerCreateBody) (*HubSpotCustomerCreateBody, *types.HTTPError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts", r.HubSpotService.BaseURL)
	return ExecuteHubSpotRequest[HubSpotCustomerCreateBody](r.HubSpotService, http.MethodPost, url, customer)
}

func (r *CustomerService) GetCustomers(after string) ([]HubSpotCustomerResponse, *types.HTTPError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts?limit=10", r.HubSpotService.BaseURL)
	if after != "" {
		url += fmt.Sprintf("&after=%s", after)
	}

	hubSpotResponse, err := ExecuteHubSpotRequest[HubSpotCustomersResponse](r.HubSpotService, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return hubSpotResponse.Results, nil
}

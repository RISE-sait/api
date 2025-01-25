package hubspot

import (
	errLib "api/internal/libs/errors"
	"api/internal/libs/hubspot"
	"api/internal/services"
	"fmt"
	"net/http"
)

type HubSpotCustomersService struct {
	HubSpotService *services.HubSpotService
}

func (s *HubSpotCustomersService) GetCustomerById(id string) (*HubSpotCustomerResponse, *errLib.CommonError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&properties=firstName,lastName,family_role,email", s.HubSpotService.BaseURL, id)
	return hubspot.ExecuteHubSpotRequest[HubSpotCustomerResponse](s.HubSpotService, http.MethodGet, url, nil)
}

func (s *HubSpotCustomersService) GetCustomerByEmail(email string) (*HubSpotCustomerResponse, *errLib.CommonError) {

	fmt.Println("Email: ", email)
	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&idProperty=email&properties=firstName,lastName,family_role", s.HubSpotService.BaseURL, email)
	return hubspot.ExecuteHubSpotRequest[HubSpotCustomerResponse](s.HubSpotService, http.MethodGet, url, nil)
}

func (s *HubSpotCustomersService) CreateCustomer(customer HubSpotCustomerCreateBody) (*HubSpotCustomerCreateBody, *errLib.CommonError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts", s.HubSpotService.BaseURL)
	return hubspot.ExecuteHubSpotRequest[HubSpotCustomerCreateBody](s.HubSpotService, http.MethodPost, url, customer)
}

func (s *HubSpotCustomersService) GetCustomers(after string) ([]HubSpotCustomerResponse, *errLib.CommonError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts?limit=10", s.HubSpotService.BaseURL)
	if after != "" {
		url += fmt.Sprintf("&after=%s", after)
	}

	hubSpotResponse, err := hubspot.ExecuteHubSpotRequest[HubSpotCustomersResponse](s.HubSpotService, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return hubSpotResponse.Results, nil
}

package hubspot

import (
	"api/config"
	"api/internal/types"
	"fmt"
	"net/http"
	"time"
)

type HubSpotService struct {
	Client  *http.Client
	ApiKey  string
	BaseURL string
}

func GetHubSpotService() *HubSpotService {
	apiKey := config.Envs.HubSpotApiKey

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &HubSpotService{
		Client:  httpClient,
		ApiKey:  apiKey,
		BaseURL: "https://api.hubapi.com/",
	}
}

func (r *HubSpotService) GetCustomerById(id string) (*HubSpotCustomerResponse, *types.HTTPError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&properties=firstName,lastName,family_role,email", r.BaseURL, id)
	return ExecuteHubSpotRequest[HubSpotCustomerResponse](r, http.MethodGet, url, nil)
}

func (r *HubSpotService) GetCustomerByEmail(email string) (*HubSpotCustomerResponse, *types.HTTPError) {

	fmt.Println("Email: ", email)
	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&idProperty=email&properties=firstName,lastName,family_role", r.BaseURL, email)
	return ExecuteHubSpotRequest[HubSpotCustomerResponse](r, http.MethodGet, url, nil)
}

func (r *HubSpotService) CreateCustomer(customer HubSpotCustomerCreateBody) (*HubSpotCustomerCreateBody, *types.HTTPError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts", r.BaseURL)
	return ExecuteHubSpotRequest[HubSpotCustomerCreateBody](r, http.MethodPost, url, customer)
}

func (r *HubSpotService) GetCustomers(after string) ([]HubSpotCustomerResponse, *types.HTTPError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts?limit=10", r.BaseURL)
	if after != "" {
		url += fmt.Sprintf("&after=%s", after)
	}

	hubSpotResponse, err := ExecuteHubSpotRequest[HubSpotCustomersResponse](r, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return hubSpotResponse.Results, nil
}

package hubspot

import (
	"api/config"
	errLib "api/internal/libs/errors"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// HubSpotService handles integration with HubSpot API.
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

func (s *HubSpotService) GetCustomerById(id string) (*HubSpotCustomerResponse, *errLib.CommonError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&properties=firstName,lastName,family_role,email", s.BaseURL, id)
	return executeHubSpotRequest[HubSpotCustomerResponse](s, http.MethodGet, url, nil)
}

func (s *HubSpotService) GetCustomerByEmail(email string) (*HubSpotCustomerResponse, *errLib.CommonError) {

	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&idProperty=email&properties=firstName,lastName,family_role", s.BaseURL, email)
	response, err := executeHubSpotRequest[HubSpotCustomerResponse](s, http.MethodGet, url, nil)

	return response, err
}

func (s *HubSpotService) CreateCustomer(customer HubSpotCustomerCreateBody) *errLib.CommonError {
	url := fmt.Sprintf("%scrm/v3/objects/contacts", s.BaseURL)
	if _, err := executeHubSpotRequest[HubSpotCustomerCreateBody](s, http.MethodPost, url, customer); err != nil {
		if err.HTTPCode == http.StatusConflict {
			return errLib.New("Customer already exists", http.StatusConflict)
		}

		log.Println("error creating customer on hubspot: ", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return nil
}

func (s *HubSpotService) AssociateChildAndParent(parentId, childId string) *errLib.CommonError {
	url := fmt.Sprintf("%scrm/v4/objects/contacts/%s/associations/contacts/%s",
		s.BaseURL,
		parentId,
		childId,
	)

	request := []AssociationInput{
		{
			AssociationCategory: "USER_DEFINED",
			AssociationTypeId:   5,
		},
	}

	_, err := executeHubSpotRequest[any](s, http.MethodPut, url, request)
	return err
}

// func (s *HubSpotService) GetCustomers(after string) ([]HubSpotCustomerResponse, *errLib.CommonError) {
// 	url := fmt.Sprintf("%scrm/v3/objects/contacts?limit=10", s.BaseURL)
// 	if after != "" {
// 		url += fmt.Sprintf("&after=%s", after)
// 	}

// 	hubSpotResponse, err := executeHubSpotRequest[HubSpotCustomersResponse](s, http.MethodGet, url, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return hubSpotResponse.Results, nil
// }

func executeHubSpotRequest[T any](s *HubSpotService, method, url string, body any) (*T, *errLib.CommonError) {
	var requestBody []byte
	var err error

	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, errLib.New(err.Error(), http.StatusBadRequest)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, errLib.New(err.Error(), http.StatusInternalServerError)
	}

	req.Header.Set("Authorization", "Bearer "+s.ApiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.Client.Do(req)

	if err != nil {
		return nil, errLib.New(err.Error(), http.StatusInternalServerError)
	}

	defer resp.Body.Close()

	statusCode := resp.StatusCode
	bytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, errLib.New(err.Error(), http.StatusInternalServerError)
	}
	if statusCode >= 400 {
		return nil, errLib.New(string(bytes), statusCode)
	}

	var result T

	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, errLib.New(err.Error(), http.StatusInternalServerError)
	}

	// success

	return &result, nil
}

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

type Service struct {
	Client  *http.Client
	ApiKey  string
	BaseURL string
}

// GetHubSpotService initializes and returns a new HubSpot service instance.
//
// Returns:
//   - *Service: A pointer to the initialized HubSpot service.
func GetHubSpotService(apiKeyPtr *string) *Service {

	var apiKey string

	if apiKeyPtr != nil {
		apiKey = *apiKeyPtr
	} else {
		apiKey = config.Envs.HubSpotApiKey
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &Service{
		Client:  httpClient,
		ApiKey:  apiKey,
		BaseURL: "https://api.hubapi.com/",
	}
}

// GetUsersByIds retrieves multiple users from HubSpot based on their IDs.
//
// Parameters:
//   - ids: A slice of user IDs.
//
// Returns:
//   - [] UserResponse: A list of user responses from HubSpot.
//   - *errLib.CommonError: An error if the request fails.
func (s *Service) GetUsersByIds(ids []string) ([]UserResponse, *errLib.CommonError) {
	if len(ids) == 0 {
		return nil, errLib.New("No customer IDs provided", http.StatusBadRequest)
	}

	url := fmt.Sprintf("%scrm/v3/objects/contacts/batch/read", s.BaseURL)

	// Construct the request body for batch retrieval
	requestBody := map[string]interface{}{
		"properties": []string{"firstName", "lastName", "email", "family_role", "phone", "age", "hs_country_region_code", "has_sms_consent", "has_marketing_email_consent"},
		"inputs":     make([]map[string]string, len(ids)),
	}

	for i, id := range ids {
		requestBody["inputs"].([]map[string]string)[i] = map[string]string{"id": id}
	}

	// Execute the batch request
	response, err := executeHubSpotRequest[UsersResponse](s, http.MethodPost, url, requestBody)
	if err != nil {
		log.Println("GetUsersByIds err:", err)
		return nil, errLib.New("Error fetching multiple customers", http.StatusInternalServerError)
	}

	return response.Results, nil
}

// GetUserById retrieves a user from HubSpot using their Hubspot ID.
//
// Parameters:
//   - id: The HubSpot user ID.
//
// Returns:
//   - *UserResponse: The user data from HubSpot.
//   - *errLib.CommonError: An error if retrieval fails.
func (s *Service) GetUserById(id string) (*UserResponse, *errLib.CommonError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&properties=firstName,lastName,family_role,email,phone,age,hs_country_region_code,has_sms_consent,has_marketing_email_consent", s.BaseURL, id)
	response, err := executeHubSpotRequest[UserResponse](s, http.MethodGet, url, nil)

	if err != nil {
		log.Println("GetUserById err:", err)
		return nil, errLib.New(fmt.Sprintf("Error getting customer with id %v", id), http.StatusInternalServerError)
	}

	return response, nil
}

// GetUserByEmail retrieves a user from HubSpot by email.
//
// Parameters:
//   - email: The user's email address.
//
// Returns:
//   - *UserResponse: The user data from HubSpot.
//   - *errLib.CommonError: An error if retrieval fails.
func (s *Service) GetUserByEmail(email string) (*UserResponse, *errLib.CommonError) {

	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s?associations=contacts&idProperty=email&properties=firstName,lastName,phone,age,hs_country_region_code,has_sms_consent,has_marketing_email_consent", s.BaseURL, email)
	response, err := executeHubSpotRequest[UserResponse](s, http.MethodGet, url, nil)

	if err != nil {
		log.Println("GetUserByEmail err:", err)
		return nil, errLib.New(fmt.Sprintf("Error getting user with email %s", email), http.StatusInternalServerError)
	}

	if response == nil {
		return nil, errLib.New(fmt.Sprintf("No user found with associated email %s", email), http.StatusNotFound)
	}

	return response, nil
}

// CreateUser creates a new user in HubSpot.
//
// Parameters:
//   - props: The user properties to create.
//
// Returns:
//   - string: The HubSpot user ID of the created user.
//   - *errLib.CommonError: An error if creation fails.
func (s *Service) CreateUser(props UserProps) (string, *errLib.CommonError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts", s.BaseURL)

	body := UserCreationBody{
		Properties: props,
	}

	response, err := executeHubSpotRequest[UserResponse](s, http.MethodPost, url, body)

	if err != nil {
		if err.HTTPCode == http.StatusConflict {
			return "", errLib.New("Customer already exists", http.StatusConflict)
		}

		log.Println("error creating props on hubspot: ", err)
		return "", errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return response.HubSpotId, nil
}

// AssociateChildAndParent associates a child account with a parent account in HubSpot.
//
// Parameters:
//   - parentId: The parent user's HubSpot ID.
//   - childId: The child user's HubSpot ID.
//
// Returns:
//   - *errLib.CommonError: An error if the association fails.
func (s *Service) AssociateChildAndParent(parentId, childId string) *errLib.CommonError {
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

	if err != nil {
		log.Println("error associating customer on hubspot: ", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return nil
}

// GetUsers retrieves a paginated list of users from HubSpot.
//
// Parameters:
//   - after: The pagination cursor for retrieving the next batch.
//
// Returns:
//   - [] UserResponse: A list of users from HubSpot.
//   - *errLib.CommonError: An error if retrieval fails.
func (s *Service) GetUsers(after string) ([]UserResponse, *errLib.CommonError) {
	url := fmt.Sprintf("%scrm/v3/objects/contacts?limit=10", s.BaseURL)
	if after != "" {
		url += fmt.Sprintf("&after=%s", after)
	}

	hubSpotResponse, err := executeHubSpotRequest[UsersResponse](s, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return hubSpotResponse.Results, nil
}

// DeleteUser deletes a user from HubSpot.
//
// Parameters:
//   - userId: The user's HubSpot ID.
//
// Returns:
//   - *errLib.CommonError: An error if deletion fails.
func (s *Service) DeleteUser(userId string) *errLib.CommonError {
	url := fmt.Sprintf("%scrm/v3/objects/contacts/%s", s.BaseURL, userId)

	_, err := executeHubSpotRequest[any](s, http.MethodDelete, url, nil)
	if err != nil {
		log.Println("Error deleting customer from HubSpot:", err)
		return errLib.New("Failed to delete customer", http.StatusInternalServerError)
	}

	return nil
}

// executeHubSpotRequest makes an HTTP request to the HubSpot API and unmarshal the response.
//
// Parameters:
//   - s: The HubSpot service instance.
//   - method: The HTTP method (GET, POST, PUT, DELETE).
//   - url: The endpoint URL.
//   - body: The request body (if applicable).
//
// Returns:
//   - *T: The unmarshalled response.
//   - *errLib.CommonError: An error if the request fails.
func executeHubSpotRequest[T any](s *Service, method, url string, body any) (*T, *errLib.CommonError) {
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
	responseBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, errLib.New(err.Error(), http.StatusInternalServerError)
	}

	if statusCode >= 400 {
		return nil, errLib.New(string(responseBytes), statusCode)
	}

	var result T

	if err = json.Unmarshal(responseBytes, &result); err != nil {
		return nil, errLib.New(err.Error(), http.StatusInternalServerError)
	}

	// success

	return &result, nil
}

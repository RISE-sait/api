package hubspot

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestGetUserByID(t *testing.T) {

	loadEnv(t)

	apiKey := os.Getenv("HUBSPOT_API_KEY")
	require.NotEmpty(t, apiKey, "HUBSPOT_API_KEY must be set")

	hubspotService := GetHubSpotService(&apiKey)

	user, hubspotErr := hubspotService.GetUserById("101448883244")

	require.Nil(t, hubspotErr)

	email := user.Properties.Email

	// Assert course data
	require.Equal(t, "klintlee1@gmail.com", email)
}

func TestGetUserByNonExistentID(t *testing.T) {

	loadEnv(t)

	apiKey := os.Getenv("HUBSPOT_API_KEY")
	require.NotEmpty(t, apiKey, "HUBSPOT_API_KEY must be set")

	hubspotService := GetHubSpotService(&apiKey)

	user, hubspotErr := hubspotService.GetUserById("abcdnwfwefwefewee")

	require.NotNil(t, hubspotErr)

	require.Contains(t, hubspotErr.Message, "Error getting customer with id abcdnwfwefwefewee")

	require.Nil(t, user)

}

func loadEnv(t *testing.T) {
	if err := godotenv.Load("../../../config/.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
}

//func TestCreateCustomer(t *testing.T) {
//
//	tests := []struct {
//		name           string
//		options        CreateCustomerOptions
//		serverResponse int
//		// responseBody   string
//		expectedError  bool
//	}{
//		{
//			name: "successful creation",
//			options: CreateCustomerOptions{
//				Properties: HubSpotCustomerProps{
//					Email:     "weffffffffffff@test.com",
//					FirstName: "Test",
//					LastName:  "User",
//				},
//				Associations: []HubSpotAssociationCreateRequest{
//					{
//						ToId:            "94523719596",
//						AssociationType: AssociationTypeChildParent,
//					},
//				},
//			},
//			serverResponse: http.StatusOK,
//			// responseBody:   `{"id": "123"}`,
//			expectedError:  false,
//		},
//		// {
//		// 	name: "server error",
//		// 	customer: HubSpotCustomerCreateBody{
//		// 		Properties: HubSpotCustomerProps{
//		// 			Email: "test@test.com",
//		// 		},
//		// 	},
//		// 	serverResponse: http.StatusInternalServerError,
//		// 	responseBody:   `{"error": "server error"}`,
//		// 	expectedError:  true,
//		// },
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//				// Verify request
//				if r.Method != http.MethodPost {
//					t.Errorf("expected POST request, got %s", r.Method)
//				}
//				if r.Header.Get("Content-Type") != "application/json" {
//					t.Errorf("expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
//				}
//
//				// Verify request body
//				var receivedCustomer HubSpotCustomerCreateBody
//				if err := json.NewDecoder(r.Body).Decode(&receivedCustomer); err != nil {
//					t.Fatalf("failed to decode request body: %v", err)
//				}
//
//				w.WriteHeader(tt.serverResponse)
//				w.Write([]byte(tt.responseBody))
//			}))
//			defer server.Close()
//
//			service := GetHubSpotService()
//
//			err := service.RegisterCustomer(tt.customer)
//
//			if tt.expectedError && err == nil {
//				t.Error("expected error but got none")
//			}
//			if !tt.expectedError && err != nil {
//				t.Errorf("unexpected error: %v", err)
//			}
//		})
//	}
//}

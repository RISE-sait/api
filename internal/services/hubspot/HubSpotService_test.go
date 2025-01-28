package hubspot

// import (
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// func TestCreateCustomer(t *testing.T) {

// 	tests := []struct {
// 		name           string
// 		options        CreateCustomerOptions
// 		serverResponse int
// 		// responseBody   string
// 		expectedError  bool
// 	}{
// 		{
// 			name: "successful creation",
// 			options: CreateCustomerOptions{
// 				Properties: HubSpotCustomerProps{
// 					Email:     "weffffffffffff@test.com",
// 					FirstName: "Test",
// 					LastName:  "User",
// 				},
// 				Associations: []HubSpotAssociationCreateRequest{
// 					{
// 						ToId:            "94523719596",
// 						AssociationType: AssociationTypeChildParent,
// 					},
// 				},
// 			},
// 			serverResponse: http.StatusOK,
// 			// responseBody:   `{"id": "123"}`,
// 			expectedError:  false,
// 		},
// 		// {
// 		// 	name: "server error",
// 		// 	customer: HubSpotCustomerCreateBody{
// 		// 		Properties: HubSpotCustomerProps{
// 		// 			Email: "test@test.com",
// 		// 		},
// 		// 	},
// 		// 	serverResponse: http.StatusInternalServerError,
// 		// 	responseBody:   `{"error": "server error"}`,
// 		// 	expectedError:  true,
// 		// },
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				// Verify request
// 				if r.Method != http.MethodPost {
// 					t.Errorf("expected POST request, got %s", r.Method)
// 				}
// 				if r.Header.Get("Content-Type") != "application/json" {
// 					t.Errorf("expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
// 				}

// 				// Verify request body
// 				var receivedCustomer HubSpotCustomerCreateBody
// 				if err := json.NewDecoder(r.Body).Decode(&receivedCustomer); err != nil {
// 					t.Fatalf("failed to decode request body: %v", err)
// 				}

// 				w.WriteHeader(tt.serverResponse)
// 				w.Write([]byte(tt.responseBody))
// 			}))
// 			defer server.Close()

// 			service := GetHubSpotService()

// 			err := service.CreateCustomer(tt.customer)

// 			if tt.expectedError && err == nil {
// 				t.Error("expected error but got none")
// 			}
// 			if !tt.expectedError && err != nil {
// 				t.Errorf("unexpected error: %v", err)
// 			}
// 		})
// 	}
// }

package customer

import (
	"api/internal/services/hubspot"

	"github.com/go-chi/chi"
)

func RegisterCustomerRoutes(r chi.Router, HubSpotService *hubspot.HubSpotService) {
	customerHandler := NewHandler(HubSpotService)

	r.Route("/customers", func(auth chi.Router) {
		auth.Get("/", customerHandler.GetCustomers)
		auth.Get("/{email}", customerHandler.GetCustomerByEmail)
		auth.Post("/", customerHandler.CreateCustomer)
	})
}

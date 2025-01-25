package customer

import (
	"api/internal/domains/customer/hubspot"

	"github.com/go-chi/chi"
)

func RegisterCustomerRoutes(r chi.Router, HubSpotService *hubspot.HubSpotCustomersService) {
	customerHandler := NewHandler(HubSpotService)

	r.Route("/customers", func(auth chi.Router) {
		auth.Get("/", customerHandler.GetCustomers)
		auth.Get("/{email}", customerHandler.GetCustomerByEmail)
		auth.Post("/", customerHandler.CreateCustomer)
	})
}

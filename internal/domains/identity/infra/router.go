package identity

import (
	"api/cmd/server/di"
	"api/internal/domains/identity/authentication"
	"api/internal/domains/identity/infra/persistence/repository"
	"api/internal/domains/identity/oauth"
	"api/internal/domains/identity/registration"
	registration_handler "api/internal/domains/identity/registration/infra/http"
	waiver_repository "api/internal/domains/identity/registration/infra/persistence"

	"github.com/go-chi/chi"
)

func RegisterIdentityRoutes(r chi.Router, container *di.Container) {

	staffRepository := repository.NewStaffRepository(container.Queries.IdentityDb)
	usersRepository := repository.NewUserRepository(container.Queries.IdentityDb)
	usersCredentialsRepository := repository.NewUserCredentialsRepository(container.Queries.IdentityDb)
	waiverRepository := waiver_repository.NewWaiverRepository(container.Queries.WaiversDb)

	oauthService := oauth.NewService(staffRepository)

	authHandler := authentication.NewHandler(authentication.NewService(
		usersRepository, staffRepository,
	))

	registrationService := registration.NewAccountRegistrationService(usersRepository, usersCredentialsRepository, staffRepository, waiverRepository, container.DB, container.HubspotService)

	registrationHandler := registration_handler.NewHandler(registrationService)

	oauthHandler := oauth.NewHandler(oauthService)

	r.Route("/auth", func(auth chi.Router) {
		auth.Post("/traditional", authHandler.Login)
		auth.Post("/oauth/google", oauthHandler.HandleOAuthCallback)
	})

	r.Route("/register", func(registration chi.Router) {
		registration.Post("/", registrationHandler.CreateTraditionalAccount)
	})
}

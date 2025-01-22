package identity

import (
	"api/internal/domains/identity/authentication/infra/repository"
	"api/internal/domains/identity/authentication/infra/sqlc/generated"
	"api/internal/domains/identity/authentication/oauth"
	"api/internal/domains/identity/authentication/traditional"
	"github.com/go-chi/chi"
)

func RegisterIdentityRoutes(r chi.Router, queries *db.Queries) {
	authHandler := traditional.NewHandler(traditional.NewService(
		repository.NewUserRepository(queries),
		repository.NewStaffRepository(queries),
	))

	oauthHandler := oauth.NewHandler(oauth.NewService(
		repository.NewStaffRepository(queries)))

	r.Route("/auth", func(auth chi.Router) {
		auth.Post("/traditional", authHandler.Login)
		auth.Post("/oauth/google", oauthHandler.HandleOAuthCallback)
	})
}

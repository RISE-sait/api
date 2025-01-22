package identity

import (
	repository2 "api/internal/domains/identity/authentication/infra/repository"
	"api/internal/domains/identity/authentication/infra/sqlc/generated"
	oauth2 "api/internal/domains/identity/authentication/oauth"
	traditional2 "api/internal/domains/identity/authentication/traditional"
	"github.com/go-chi/chi"
)

func RegisterIdentityRoutes(r chi.Router, queries *db.Queries) {
	authHandler := traditional2.NewHandler(traditional2.NewService(
		repository2.NewUserRepository(queries),
		repository2.NewStaffRepository(queries),
	))

	oauthHandler := oauth2.NewHandler(oauth2.NewService(
		repository2.NewStaffRepository(queries)))

	r.Route("/auth", func(auth chi.Router) {
		auth.Post("/traditional", authHandler.Login)
		auth.Post("/oauth/google", oauthHandler.HandleOAuthCallback)
	})
}

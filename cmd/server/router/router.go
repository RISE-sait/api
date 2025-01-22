package router

import (
	"api/internal/domains/identity"
	db "api/internal/domains/identity/authentication/infra/sqlc/generated"
	"github.com/go-chi/chi"
)

func RegisterRoutes(r *chi.Mux, queries *db.Queries) {

	identity.RegisterIdentityRoutes(r, queries)

}

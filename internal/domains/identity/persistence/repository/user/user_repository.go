package user

import (
	"api/internal/di"
	dbIdentity "api/internal/domains/identity/persistence/sqlc/generated"
	dbOutbox "api/internal/services/outbox/generated"
)

// UsersRepository provides methods to interact with the user data in the database.
type UsersRepository struct {
	IdentityQueries *dbIdentity.Queries
	OutboxQueries   *dbOutbox.Queries
}

// NewUserRepository creates a new instance of UserRepository with the provided dependency injection container.
func NewUserRepository(container *di.Container) *UsersRepository {
	return &UsersRepository{
		IdentityQueries: container.Queries.IdentityDb,
		OutboxQueries:   container.Queries.OutboxDb,
	}
}

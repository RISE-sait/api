package identity

import (
	dbIdentity "api/internal/domains/identity/persistence/sqlc/generated"
	dbOutbox "api/internal/services/outbox/generated"
)

// UsersRepository provides methods to interact with the user data in the database.
type UsersRepository struct {
	IdentityQueries *dbIdentity.Queries
	OutboxQueries   *dbOutbox.Queries
}

// NewUserRepository creates a new instance of UserRepository with the provided dependency injection container.
func NewUserRepository(identityDb *dbIdentity.Queries, outboxDb *dbOutbox.Queries) *UsersRepository {
	return &UsersRepository{
		IdentityQueries: identityDb,
		OutboxQueries:   outboxDb,
	}
}

package repository

import (
	"api/cmd/server/di"
	database_errors "api/internal/constants"
	"api/internal/domains/identity/entities"
	db "api/internal/domains/identity/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/lib/pq"
)

type PendingChildAccountRepository struct {
	Queries *db.Queries
}

func NewPendingChildAcountRepository(container *di.Container) *PendingChildAccountRepository {
	return &PendingChildAccountRepository{
		Queries: container.Queries.IdentityDb,
	}
}

func (r *PendingChildAccountRepository) GetPendingChildAccountByChildEmail(ctx context.Context, email string) (*entities.PendingChildAccount, *errLib.CommonError) {
	account, err := r.Queries.GetPendingChildAccountByChildEmail(ctx, email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Child not found", http.StatusNotFound)
		}
		log.Printf("Error fetching child account by email: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &entities.PendingChildAccount{
		UserEmail:   account.UserEmail,
		ParentEmail: account.ParentEmail,
		Password:    account.Password.String,
		CreatedAt:   account.CreatedAt,
	}, nil
}

func (r *PendingChildAccountRepository) CreatePendingChildAccountTx(ctx context.Context, tx *sql.Tx, email, parentEmail string, password *string) *errLib.CommonError {

	passwordStr := ""

	if password != nil {
		passwordStr = *password
	}

	params := db.CreatePendingChildAccountParams{
		UserEmail:   email,
		ParentEmail: parentEmail,
		Password: sql.NullString{
			String: passwordStr,
			Valid:  passwordStr != "",
		},
	}

	rows, err := r.Queries.WithTx(tx).CreatePendingChildAccount(ctx, params)

	log.Println(params)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle PostgreSQL unique violation errors (e.g., duplicate staff emails)
			if pqErr.Code == database_errors.UniqueViolation { // Unique violation
				return errLib.New("Account with this email already exists", http.StatusConflict)
			}
		}
		log.Printf("Error creating account: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if rows != 1 {
		log.Println("Error creating account ", err)
		return errLib.New("Failed to create account", http.StatusInternalServerError)
	}

	return nil
}

func (r *PendingChildAccountRepository) DeleteAccount(ctx context.Context, tx *sql.Tx, email string) *errLib.CommonError {

	rows, err := r.Queries.WithTx(tx).DeletePendingChildAccount(ctx, email)

	if err != nil {
		log.Printf("Error deleting account: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if rows != 1 {
		log.Println("Error deleting account ", err)
		return errLib.New("Failed to delete account", 500)
	}

	return nil
}

package repository

import (
	database_errors "api/internal/constants"
	"api/internal/di"
	entity "api/internal/domains/identity/entities"
	db "api/internal/domains/identity/persistence/sqlc/generated"
	"api/internal/domains/identity/values"
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

func (r *PendingChildAccountRepository) GetPendingChildAccountByChildEmail(ctx context.Context, email string) (*entity.PendingChildAccount, *errLib.CommonError) {
	account, err := r.Queries.GetPendingChildAccountByChildEmail(ctx, email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Child not found", http.StatusNotFound)
		}
		log.Printf("Error fetching child account by email: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	pendingChildAccount := entity.PendingChildAccount{
		ID:          account.ID,
		UserEmail:   account.UserEmail,
		ParentEmail: account.ParentEmail,
		CreatedAt:   account.CreatedAt,
	}

	if account.FirstName.Valid {
		pendingChildAccount.FirstName = &account.FirstName.String
	}

	if account.LastName.Valid {
		pendingChildAccount.LastName = &account.LastName.String
	}

	if account.Password.Valid {
		pendingChildAccount.Password = &account.Password.String
	}

	return &pendingChildAccount, nil
}

func (r *PendingChildAccountRepository) CreatePendingChildAccountTx(ctx context.Context, tx *sql.Tx, childAccountCreate *values.CreatePendingChildAccountValueObject) (*entity.PendingChildAccount, *errLib.CommonError) {

	params := db.CreatePendingChildAccountParams{
		UserEmail:   childAccountCreate.Email,
		ParentEmail: childAccountCreate.ParentEmail,
	}

	if childAccountCreate.FirstName != nil {
		params.FirstName = sql.NullString{
			String: *childAccountCreate.FirstName,
			Valid:  true,
		}
	}

	if childAccountCreate.LastName != nil {
		params.LastName = sql.NullString{
			String: *childAccountCreate.LastName,
			Valid:  true,
		}
	}

	if childAccountCreate.Password != nil {
		params.Password = sql.NullString{
			String: *childAccountCreate.Password,
			Valid:  true,
		}
	}

	child, err := r.Queries.WithTx(tx).CreatePendingChildAccount(ctx, params)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Handle PostgreSQL unique violation errors (e.g., duplicate staff emails)
			if pqErr.Code == database_errors.UniqueViolation { // Unique violation
				return nil, errLib.New("Account with this email already exists", http.StatusConflict)
			}
		}
		log.Printf("Error creating account: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	pendingChildAccount := entity.PendingChildAccount{
		ID:          child.ID,
		UserEmail:   child.UserEmail,
		ParentEmail: child.ParentEmail,
		CreatedAt:   child.CreatedAt,
	}

	if child.Password.Valid {
		pendingChildAccount.Password = &child.Password.String
	}

	return &pendingChildAccount, nil

}

func (r *PendingChildAccountRepository) DeleteAccount(ctx context.Context, tx *sql.Tx, email string) *errLib.CommonError {

	rows, err := r.Queries.WithTx(tx).DeletePendingChildAccount(ctx, email)

	if err != nil {
		log.Printf("Error deleting account: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if rows == 0 {
		log.Println("Error deleting account ", err)
		return errLib.New("Failed to delete account", 500)
	}

	return nil
}

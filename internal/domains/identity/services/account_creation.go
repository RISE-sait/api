package identity

import (
	"api/internal/di"
	entity "api/internal/domains/identity/entities"
	repo "api/internal/domains/identity/persistence/repository"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"net/http"
)

type AccountCreationService struct {
	UsersRepository *repo.UserRepository
	DB              *sql.DB
}

func NewAccountCreationService(
	container *di.Container,
) *AccountCreationService {
	return &AccountCreationService{
		UsersRepository: repo.NewUserRepository(container),
		DB:              container.DB,
	}
}

func (s *AccountCreationService) CreateAccount(
	ctx context.Context,
	tx *sql.Tx,
	email string,
	shouldCommit bool,
) (*sql.Tx, *entity.UserInfo, *errLib.CommonError) {

	// Begin transaction
	if tx == nil {
		newTx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
		if txErr != nil {
			return nil, nil, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
		}

		tx = newTx
	}

	// Ensure rollback if something goes wrong
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Insert into users table
	if err := s.UsersRepository.CreateUserTx(ctx, tx, email); err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// Commit the transaction

	if shouldCommit {
		if err := tx.Commit(); err != nil {
			return nil, nil, errLib.New("Failed to commit transaction", http.StatusInternalServerError)
		}
	}

	// Generate JWT for the new user
	userInfo := entity.UserInfo{
		Email: email,
	}

	return tx, &userInfo, nil
}

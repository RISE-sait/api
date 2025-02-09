package identity

import (
	"api/internal/di"
	"api/internal/domains/identity/entities"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"net/http"
	"strings"
)

type AccountRegistrationService struct {
	UsersRepository       *repo.UserRepository
	CredentialsRepository *repo.UserCredentialsRepository
	DB                    *sql.DB
}

func NewAccountRegistrationService(
	container *di.Container,
) *AccountRegistrationService {
	return &AccountRegistrationService{
		UsersRepository:       repo.NewUserRepository(container),
		CredentialsRepository: repo.NewUserCredentialsRepository(container),
		DB:                    container.DB,
	}
}

func (s *AccountRegistrationService) CreateAccount(
	ctx context.Context,
	registerCredentials *values.RegisterCredentials,
) (*sql.Tx, *entities.UserInfo, *errLib.CommonError) {

	emailStr := registerCredentials.Email
	password := registerCredentials.Password

	// Begin transaction
	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		return nil, nil, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	// Ensure rollback if something goes wrong
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Insert into users table
	if err := s.UsersRepository.CreateUserTx(ctx, tx, emailStr); err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// Insert into credentials (password).

	if password != nil {

		if err := s.CredentialsRepository.CreatePasswordTx(ctx, tx, emailStr, *password); err != nil {
			tx.Rollback()
			return nil, nil, err
		}
	}

	// // Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, nil, errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	// Generate JWT for the new user
	userInfo := entities.UserInfo{
		Name:  strings.Split(emailStr, "@")[0],
		Email: emailStr,
	}

	return tx, &userInfo, nil
}

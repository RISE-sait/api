package identity

import (
	"api/cmd/server/di"
	dto "api/internal/domains/identity/dto"
	"api/internal/domains/identity/entities"
	repo "api/internal/domains/identity/persistence/repository"
	errLib "api/internal/libs/errors"
	"api/internal/services/hubspot"
	"context"
	"database/sql"
	"net/http"
	"strings"
)

type AccountRegistrationService struct {
	UsersRepository         *repo.UserRepository
	CredentialsRepository   *repo.UserCredentialsRepository
	HubSpotService          *hubspot.HubSpotService
	WaiverSigningRepository *repo.WaiverSigningRepository
	DB                      *sql.DB
}

func NewAccountRegistrationService(
	container *di.Container,
) *AccountRegistrationService {
	return &AccountRegistrationService{
		UsersRepository:         repo.NewUserRepository(container),
		CredentialsRepository:   repo.NewUserCredentialsRepository(container),
		WaiverSigningRepository: repo.NewWaiverSigningRepository(container),
		DB:                      container.DB,
		HubSpotService:          container.HubspotService,
	}
}

func (s *AccountRegistrationService) CreateCustomer(
	ctx context.Context,
	customerCreate *dto.CustomerWaiverCreateDto,
	credentialsDto *dto.Credentials,
) (*entities.UserInfo, *errLib.CommonError) {

	// Begin transaction
	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		return nil, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	// Ensure rollback if something goes wrong
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := credentialsDto.Validate(); err != nil {
		return nil, err
	}

	email := credentialsDto.Email
	password := credentialsDto.Password

	// Insert into users table
	if err := s.UsersRepository.CreateUserTx(ctx, tx, email); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Insert into credentials (password).

	if err := s.CredentialsRepository.CreatePasswordTx(ctx, tx, email, password); err != nil {
		tx.Rollback()
		return nil, err
	}

	// User is Customer

	if err := customerCreate.Validate(); err != nil {
		return nil, err
	}

	if !customerCreate.IsWaiverSigned {
		tx.Rollback()
		return nil, errLib.New("Waiver is not signed", http.StatusBadRequest)
	}

	if err := s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, credentialsDto.Email, customerCreate.WaiverUrl, customerCreate.IsWaiverSigned); err != nil {
		tx.Rollback()
		return nil, err
	}

	customer := hubspot.HubSpotCustomerCreateBody{
		Properties: hubspot.HubSpotCustomerProps{
			FirstName: strings.Split(email, "@")[0],
			Email:     email,
			LastName:  "",
		},
	}

	// Insert into customers via hubspot

	if err := s.HubSpotService.CreateCustomer(customer); err != nil {
		tx.Rollback()
		return nil, err
	}

	// // Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	// Generate JWT for the new user
	userInfo := entities.UserInfo{
		Name:      strings.Split(email, "@")[0],
		Email:     email,
		StaffInfo: nil,
	}

	return &userInfo, nil
}

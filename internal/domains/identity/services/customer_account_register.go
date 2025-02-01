package identity

import (
	"api/cmd/server/di"
	"api/internal/domains/identity/entities"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
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
	customerCreate *values.CustomerRegistrationValueObject,
) (*entities.UserInfo, *errLib.CommonError) {

	emailStr := customerCreate.Email
	password := customerCreate.Password

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

	// Insert into users table
	if err := s.UsersRepository.CreateUserTx(ctx, tx, emailStr); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Insert into credentials (password).

	if password != nil {

		if err := s.CredentialsRepository.CreatePasswordTx(ctx, tx, emailStr, *password); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	for _, waiver := range customerCreate.Waivers {
		if !waiver.IsWaiverSigned {
			return nil, errLib.New("Waiver is not signed", http.StatusBadRequest)
		}
	}

	for _, waiver := range customerCreate.Waivers {

		if err := s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, emailStr, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	customer := hubspot.HubSpotCustomerCreateBody{
		Properties: hubspot.HubSpotCustomerProps{
			FirstName: strings.Split(emailStr, "@")[0],
			Email:     emailStr,
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
		Name:      strings.Split(emailStr, "@")[0],
		Email:     emailStr,
		StaffInfo: nil,
	}

	return &userInfo, nil
}

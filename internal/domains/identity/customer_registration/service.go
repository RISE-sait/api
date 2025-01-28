package customer_registration

import (
	"api/cmd/server/di"
	waiver_repository "api/internal/domains/identity/customer_registration/infra/persistence"
	"api/internal/domains/identity/customer_registration/values"
	"api/internal/domains/identity/entities"
	"api/internal/domains/identity/infra/persistence/repository"
	errLib "api/internal/libs/errors"
	"api/internal/services/hubspot"
	"context"
	"database/sql"
	"net/http"
	"strings"
)

type AccountRegistrationService struct {
	UsersRepository        *repository.UserRepository
	UserPasswordRepository *repository.UserCredentialsRepository
	HubSpotService         *hubspot.HubSpotService
	WaiverRepository       *waiver_repository.WaiverRepository
	DB                     *sql.DB
}

func NewAccountRegistrationService(
	container *di.Container,
) *AccountRegistrationService {
	return &AccountRegistrationService{
		UsersRepository:        repository.NewUserRepository(container.Queries.IdentityDb),
		UserPasswordRepository: repository.NewUserCredentialsRepository(container.Queries.IdentityDb),
		WaiverRepository:       waiver_repository.NewWaiverRepository(container.Queries.WaiversDb),
		DB:                     container.DB,
		HubSpotService:         container.HubspotService,
	}
}

func (s *AccountRegistrationService) CreateCustomer(
	ctx context.Context,
	userPasswordCreate *values.UserPasswordCreate,
	waiverCreate *values.WaiverCreate,
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

	if err := userPasswordCreate.Validate(); err != nil {
		return nil, err
	}

	email := userPasswordCreate.Email
	password := userPasswordCreate.Password

	// Insert into users table
	userId, err := s.UsersRepository.CreateUserTx(ctx, tx, email)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Insert into optional info (email/password).
	if err := s.UserPasswordRepository.CreatePasswordTx(ctx, tx, email, password); err != nil {
		tx.Rollback()
		return nil, err
	}

	// User is Customer

	if err := waiverCreate.Validate(); err != nil {
		return nil, err
	}

	_, err = s.WaiverRepository.GetWaiver(ctx, waiverCreate.WaiverUrl)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if !waiverCreate.IsSigned {
		tx.Rollback()
		return nil, errLib.New("Waiver is not signed", http.StatusBadRequest)
	}

	if err := s.WaiverRepository.CreateWaiverRecordTx(ctx, tx, userId, waiverCreate.WaiverUrl, waiverCreate.IsSigned); err != nil {
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

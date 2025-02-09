package identity

import (
	"api/internal/di"
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

type CustomerAccountRegistrationService struct {
	UsersRepository         *repo.UserRepository
	CredentialsRepository   *repo.UserCredentialsRepository
	HubSpotService          *hubspot.HubSpotService
	WaiverSigningRepository *repo.WaiverSigningRepository
	DB                      *sql.DB
}

func NewCustomerAccountRegistrationService(
	container *di.Container,
) *CustomerAccountRegistrationService {
	return &CustomerAccountRegistrationService{
		UsersRepository:         repo.NewUserRepository(container),
		CredentialsRepository:   repo.NewUserCredentialsRepository(container),
		WaiverSigningRepository: repo.NewWaiverSigningRepository(container),
		DB:                      container.DB,
		HubSpotService:          container.HubspotService,
	}
}

func (s *CustomerAccountRegistrationService) CreateCustomer(
	ctx context.Context,
	tx *sql.Tx,
	customerCreate *values.CustomerRegistrationValueObject,
) (*entities.UserInfo, *errLib.CommonError) {

	emailStr := customerCreate.Email

	// Begin transaction

	if tx == nil {
		newTx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
		if txErr != nil {
			return nil, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
		}

		tx = newTx
	}

	// Ensure rollback if something goes wrong
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

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
		Name:  strings.Split(emailStr, "@")[0],
		Email: emailStr,
	}

	return &userInfo, nil
}

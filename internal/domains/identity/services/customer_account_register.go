package identity

import (
	"api/internal/di"
	entity "api/internal/domains/identity/entities"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/services/hubspot"
	"context"
	"database/sql"
	"net/http"
	"strings"
)

type CustomerRegistrationService struct {
	AccountService          *AccountCreationService
	UserOptionalInfoService *UserOptionalInfoService
	WaiverSigningRepository *repo.WaiverSigningRepository
	DB                      *sql.DB
	HubSpotService          *hubspot.HubSpotService
}

func NewCustomerRegistrationService(container *di.Container) *CustomerRegistrationService {
	return &CustomerRegistrationService{
		AccountService:          NewAccountCreationService(container),
		UserOptionalInfoService: NewUserOptionalInfoService(container),
		WaiverSigningRepository: repo.NewWaiverSigningRepository(container),
		DB:                      container.DB,
		HubSpotService:          hubspot.GetHubSpotService(),
	}
}

func (s *CustomerRegistrationService) RegisterCustomer(
	ctx context.Context,
	customer *values.CustomerRegistrationInfo,
) (*entity.UserInfo, *errLib.CommonError) {

	email := customer.Email
	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		return nil, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	_, _, err := s.AccountService.CreateAccount(ctx, tx, customer.Email, false)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx, err = s.UserOptionalInfoService.CreateUserOptionalInfoTx(ctx, tx, customer.UserInfo, customer.Password)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, waiver := range customer.Waivers {
		if !waiver.IsWaiverSigned {
			tx.Rollback()
			return nil, errLib.New("Waiver is not signed", http.StatusBadRequest)
		}

		if err := s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, email, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	hubspotCustomer := hubspot.HubSpotCustomerCreateBody{
		Properties: hubspot.HubSpotCustomerProps{
			FirstName: strings.Split(email, "@")[0],
			Email:     email,
			LastName:  "",
		},
	}

	// Insert into customers via hubspot

	if err := s.HubSpotService.CreateCustomer(hubspotCustomer); err != nil {
		tx.Rollback()
		return nil, err
	}

	// // Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	// Generate JWT for the new user
	userInfo := entity.UserInfo{
		FirstName: customer.UserInfo.FirstName,
		LastName:  customer.UserInfo.LastName,
		Email:     email,
	}

	return &userInfo, nil
}

package registration

import (
	"api/internal/di"
	identityRepo "api/internal/domains/identity/persistence/repository"
	waiverSigningRepo "api/internal/domains/identity/persistence/repository/waiver_signing"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"net/http"
)

// CustomerRegistrationService handles customer registration and related operations.
type CustomerRegistrationService struct {
	UserRepo                *identityRepo.UsersRepository
	UserInfoTempRepo        *identityRepo.PendingUsersRepo
	WaiverSigningRepository *waiverSigningRepo.PendingUserWaiverSigningRepository
	DB                      *sql.DB
}

// NewCustomerRegistrationService initializes a new CustomerRegistrationService instance.
func NewCustomerRegistrationService(container *di.Container) *CustomerRegistrationService {
	return &CustomerRegistrationService{
		UserRepo:                identityRepo.NewUserRepository(container),
		WaiverSigningRepository: waiverSigningRepo.NewPendingUserWaiverSigningRepository(container),
		UserInfoTempRepo:        identityRepo.NewPendingUserInfoRepository(container),
		DB:                      container.DB,
	}
}

// RegisterCustomer registers a new customer, ensuring all waivers are signed, creating user and temp info records.
//
// Parameters:
// - ctx: Context for request lifecycle management.
// - customer: *identity.RegularCustomerRegistrationRequestInfo Customer registration data including name, email, and waiver signings.
//
// Returns:
// - *errLib.CommonError: Error if registration fails.
func (s *CustomerRegistrationService) RegisterCustomer(
	ctx context.Context,
	customer *identity.RegularCustomerRegistrationRequestInfo,
) *errLib.CommonError {

	for _, waiver := range customer.Waivers {
		if !waiver.IsWaiverSigned {
			return errLib.New("Waiver is not signed", http.StatusBadRequest)
		}
	}

	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		return errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	userId, err := s.UserInfoTempRepo.CreatePendingUserInfoTx(ctx, tx, customer.FirstName, customer.LastName, &customer.Email, nil, customer.Age)

	if err != nil {
		tx.Rollback()
		return err
	}

	for _, waiver := range customer.Waivers {
		if err := s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, userId, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	return nil
}

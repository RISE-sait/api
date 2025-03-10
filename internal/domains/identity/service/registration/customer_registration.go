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

// RegisterAthlete registers a new athlete customer, ensuring all waivers are signed and creates user and temporary info records.
//
// Parameters:
// - ctx: Context for managing the request lifecycle and cancellation.
// - customer: Contains the registration data, including name, email, phone number, age, and waivers signed.
//
// Returns:
// - *errLib.CommonError: Returns an error if registration fails or if waivers are not signed.
func (s *CustomerRegistrationService) RegisterAthlete(
	ctx context.Context,
	customer identity.AthleteRegistrationRequestInfo,
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

	userId, err := s.UserInfoTempRepo.CreatePendingUserInfoTx(ctx, tx, customer.FirstName, customer.LastName, customer.HasConsentToSms, customer.HasConsentToEmailMarketing, false, &customer.CountryCode, &customer.Phone, &customer.Email, nil, customer.Age)

	if err != nil {
		tx.Rollback()
		return err
	}

	for _, waiver := range customer.Waivers {
		if err = s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, userId, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		tx.Rollback()
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	return nil
}

// RegisterParent registers a new parent customer and creates user and temporary info records.
//
// Parameters:
// - ctx: Context for managing the request lifecycle and cancellation.
// - customer: Contains the registration data, including name, email, phone number, age, and consent information.
//
// Returns:
// - *errLib.CommonError: Returns an error if registration fails.
func (s *CustomerRegistrationService) RegisterParent(
	ctx context.Context,
	customer identity.AdultCustomerRegistrationRequestInfo,
) *errLib.CommonError {

	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		return errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()

		}
	}()

	_, err := s.UserInfoTempRepo.CreatePendingUserInfoTx(ctx, tx, customer.FirstName, customer.LastName, customer.HasConsentToSms, customer.HasConsentToEmailMarketing, true, nil, &customer.Phone, &customer.Email, nil, customer.Age)

	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		tx.Rollback()
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	return nil
}

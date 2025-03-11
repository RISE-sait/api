package registration

import (
	"api/internal/di"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"net/http"
)

// CustomerRegistrationService handles customer registration and related operations.
type CustomerRegistrationService struct {
	UserRepo          *repo.UsersRepository
	WaiverSigningRepo *repo.WaiverSigningRepository
	DB                *sql.DB
}

// NewCustomerRegistrationService initializes a new CustomerRegistrationService instance.
func NewCustomerRegistrationService(container *di.Container) *CustomerRegistrationService {

	identityDb := container.Queries.IdentityDb
	outboxDb := container.Queries.OutboxDb

	return &CustomerRegistrationService{
		UserRepo:          repo.NewUserRepository(identityDb, outboxDb),
		WaiverSigningRepo: repo.NewWaiverSigningRepository(container.Queries.IdentityDb),
		DB:                container.DB,
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
) (identity.UserReadInfo, *errLib.CommonError) {

	var response identity.UserReadInfo

	for _, waiver := range customer.Waivers {
		if !waiver.IsWaiverSigned {
			return response, errLib.New("Waiver is not signed", http.StatusBadRequest)
		}
	}

	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		return response, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	userInfo, err := s.UserRepo.CreateAthleteTx(ctx, tx, customer)

	if err != nil {
		tx.Rollback()
		return response, err
	}

	for _, waiver := range customer.Waivers {
		if err = s.WaiverSigningRepo.CreateWaiverSigningRecordTx(ctx, tx, userInfo.ID, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {
			tx.Rollback()
			return response, err
		}
	}

	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		tx.Rollback()
		return response, errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	return userInfo, nil
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
	customer identity.ParentRegistrationRequestInfo,
) (identity.UserReadInfo, *errLib.CommonError) {

	var response identity.UserReadInfo

	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		return response, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()

		}
	}()

	userInfo, err := s.UserRepo.CreateParentTx(ctx, tx, customer)

	if err != nil {
		tx.Rollback()
		return response, err
	}

	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		tx.Rollback()
		return response, errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	return userInfo, nil
}

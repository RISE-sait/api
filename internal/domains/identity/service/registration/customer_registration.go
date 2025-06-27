package registration

import (
	"api/internal/di"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/persistence/repository/user"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"api/utils/email"
)

// CustomerRegistrationService handles customer registration and related operations.
type CustomerRegistrationService struct {
	UserRepo          *user.UsersRepository
	WaiverSigningRepo *repo.WaiverSigningRepository
	DB                *sql.DB
}

// NewCustomerRegistrationService initializes a new CustomerRegistrationService instance.
func NewCustomerRegistrationService(container *di.Container) *CustomerRegistrationService {

	return &CustomerRegistrationService{
		UserRepo:          user.NewUserRepository(container),
		WaiverSigningRepo: repo.NewWaiverSigningRepository(container),
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

	requiredWaivers, err := s.WaiverSigningRepo.GetRequiredWaivers(ctx)

	if err != nil {
		return identity.UserReadInfo{}, err
	}

	if err = validateWaivers(customer.Waivers, requiredWaivers); err != nil {
		return identity.UserReadInfo{}, err
	}

	waiverUrls, areWaiversSigned := prepareWaiverData(requiredWaivers)

	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		return identity.UserReadInfo{}, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	createdUserInfo, err := s.UserRepo.CreateAthleteTx(ctx, tx, customer)

	if err != nil {
		tx.Rollback()
		return identity.UserReadInfo{}, err
	}

	var userIds []uuid.UUID

	for range waiverUrls {
		userIds = append(userIds, createdUserInfo.ID)
	}

	if err = s.WaiverSigningRepo.CreateWaiversSigningRecordTx(ctx, tx, userIds, waiverUrls, areWaiversSigned); err != nil {
		tx.Rollback()
		return identity.UserReadInfo{}, err
	}

	// Commit the transaction
	if txErr = tx.Commit(); txErr != nil {
		tx.Rollback()
		return identity.UserReadInfo{}, errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}
	if createdUserInfo.Email != nil {
		email.SendSignUpConfirmationEmail(*createdUserInfo.Email, createdUserInfo.FirstName)
	}

	return createdUserInfo, nil
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
	if userInfo.Email != nil {
		email.SendSignUpConfirmationEmail(*userInfo.Email, userInfo.FirstName)
	}

	return userInfo, nil
}

func validateWaivers(customerWaivers []identity.CustomerWaiverSigning, requiredWaivers []identity.Waiver) *errLib.CommonError {
	for _, waiver := range requiredWaivers {
		found := false
		for _, customerWaiver := range customerWaivers {
			if customerWaiver.WaiverUrl == waiver.URL {
				if !customerWaiver.IsWaiverSigned {
					return errLib.New(fmt.Sprintf("Waiver %v, url: %v, is not signed", waiver.Name, waiver.URL), http.StatusBadRequest)
				}
				found = true
				break
			}
		}
		if !found {
			return errLib.New(fmt.Sprintf("Waiver %v, url: %v, is not provided", waiver.Name, waiver.URL), http.StatusBadRequest)
		}
	}
	return nil
}

// prepareWaiverData prepares the data for waiver signing records
func prepareWaiverData(requiredWaivers []identity.Waiver) ([]string, []bool) {
	var (
		waiverUrls       = make([]string, len(requiredWaivers))
		areWaiversSigned = make([]bool, len(requiredWaivers))
	)

	for i, waiver := range requiredWaivers {
		waiverUrls[i] = waiver.URL
		areWaiversSigned[i] = true
	}

	return waiverUrls, areWaiversSigned
}

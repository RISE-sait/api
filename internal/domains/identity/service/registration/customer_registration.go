package registration

import (
	"api/internal/di"
	"api/internal/domains/identity/persistence/repository/user"
	userInfoTempRepo "api/internal/domains/identity/persistence/repository/user_info"
	waiverSigningRepo "api/internal/domains/identity/persistence/repository/waiver_signing"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	jwtLib "api/internal/libs/jwt"
	"context"
	"database/sql"
	"net/http"
)

// CustomerRegistrationService handles customer registration and related operations.
type CustomerRegistrationService struct {
	UserRepo                user.RepositoryInterface
	UserInfoTempRepo        userInfoTempRepo.InfoTempRepositoryInterface
	WaiverSigningRepository waiverSigningRepo.RepositoryInterface
	DB                      *sql.DB
}

// NewCustomerRegistrationService initializes a new CustomerRegistrationService instance.
func NewCustomerRegistrationService(container *di.Container) *CustomerRegistrationService {
	return &CustomerRegistrationService{
		UserRepo:                user.NewUserRepository(container),
		WaiverSigningRepository: waiverSigningRepo.NewWaiverSigningRepository(container),
		UserInfoTempRepo:        userInfoTempRepo.NewInfoTempRepository(container),
		DB:                      container.DB,
	}
}

// RegisterCustomer registers a new customer, ensuring all waivers are signed, creating user and temp info records,
// and signing a JWT token upon successful registration.
//
// Parameters:
// - ctx: Context for request lifecycle management.
// - customer: *values.RegularCustomerRegistrationInfo Customer registration data including name, email, and waiver signings.
//
// Returns:
// - *entity.UserInfo: User information object upon successful registration.
// - *string: Signed JWT token for authentication.
// - *errLib.CommonError: Error if registration fails.
func (s *CustomerRegistrationService) RegisterCustomer(
	ctx context.Context,
	customer *values.RegularCustomerRegistrationInfo,
) (*string, *errLib.CommonError) {

	for _, waiver := range customer.Waivers {
		if !waiver.IsWaiverSigned {
			return nil, errLib.New("Waiver is not signed", http.StatusBadRequest)
		}
	}

	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		return nil, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	userId, err := s.UserRepo.CreateUserTx(ctx, tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = s.UserInfoTempRepo.CreateTempUserInfoTx(ctx, tx, *userId, customer.FirstName, customer.LastName, &customer.Email, nil, customer.Age)

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, waiver := range customer.Waivers {
		if err := s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, *userId, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return nil, errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	customClaims := jwtLib.CustomClaims{
		UserID:    *userId,
		HubspotID: nil,
	}

	signedToken, err := jwtLib.SignJWT(customClaims)

	if err != nil {
		return nil, errLib.New("Registered successfully but failed to sign JWT token. Please try logging in.", http.StatusInternalServerError)
	}

	return &signedToken, nil
}

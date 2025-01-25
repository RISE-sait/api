package registration

import (
	"api/internal/domains/identity/entities"
	"api/internal/domains/identity/registration/infra/persistence/repository"
	errLib "api/internal/libs/errors"
	"api/internal/services/hubspot"
	"context"
	"database/sql"
	"net/http"
	"strings"
)

type AccountRegistrationService struct {
	UsersRepository        *repository.UserRepository
	UserPasswordRepository *repository.UserPasswordRepository
	HubSpotService         *hubspot.HubSpotService
	StaffRepository        *repository.StaffRepository
	DB                     *sql.DB
}

func NewAccountRegistrationService(
	UsersRepository *repository.UserRepository,
	UserPasswordRepository *repository.UserPasswordRepository,
	StaffRepository *repository.StaffRepository,
	db *sql.DB,
	HubSpotService *hubspot.HubSpotService,
) *AccountRegistrationService {
	return &AccountRegistrationService{
		UsersRepository:        UsersRepository,
		UserPasswordRepository: UserPasswordRepository,
		StaffRepository:        StaffRepository,
		DB:                     db,
		HubSpotService:         HubSpotService,
	}
}

func (s *AccountRegistrationService) CreateTraditionalAccount(
	ctx context.Context,
	email, password string, role string, isActive bool,
) (*entities.UserInfo, *errLib.CommonError) {

	// Begin transaction
	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	// Ensure rollback if something goes wrong
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	email = strings.TrimSpace(email)

	if email == "" {
		return nil, errLib.New("Email cannot be empty or whitespace", http.StatusBadRequest)
	}

	if !strings.Contains(email, "@") {
		return nil, errLib.New("Invalid email", http.StatusBadRequest)
	}

	// Insert into users table
	if err := s.UsersRepository.CreateUserTx(ctx, tx, email); err != nil {
		tx.Rollback()
		return nil, err
	}

	password = strings.TrimSpace(password)

	if password == "" {
		return nil, errLib.New("Password cannot be empty or whitespace", http.StatusBadRequest)
	}

	if len(password) < 8 {
		return nil, errLib.New("Password must be at least 8 characters", http.StatusBadRequest)
	}

	// Insert into optional info (email/password).
	if err := s.UserPasswordRepository.CreatePasswordTx(ctx, tx, email, password); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Insert into staff table if role is not 0

	if role != "0" {
		if err := s.StaffRepository.CreateStaffTx(ctx, tx, role, isActive); err != nil {
			tx.Rollback()
			return nil, err
		}
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

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	// Generate JWT for the new user
	userInfo := entities.UserInfo{
		Name:  strings.Split(email, "@")[0],
		Email: email,
		StaffInfo: &entities.StaffInfo{
			Role:     role,
			IsActive: isActive,
		},
	}

	return &userInfo, nil
}

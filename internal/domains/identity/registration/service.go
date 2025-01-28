package registration

import (
	"api/cmd/server/di"
	"api/internal/domains/identity/entities"
	"api/internal/domains/identity/infra/persistence/repository"
	waiver_repository "api/internal/domains/identity/registration/infra/persistence"
	"api/internal/domains/identity/registration/values"
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
	StaffRepository        *repository.StaffRepository
	WaiverRepository       *waiver_repository.WaiverRepository
	DB                     *sql.DB
}

func NewAccountRegistrationService(
	container *di.Container,
) *AccountRegistrationService {
	return &AccountRegistrationService{
		UsersRepository:        repository.NewUserRepository(container.Queries.IdentityDb),
		UserPasswordRepository: repository.NewUserCredentialsRepository(container.Queries.IdentityDb),
		StaffRepository:        repository.NewStaffRepository(container.Queries.IdentityDb),
		WaiverRepository:       waiver_repository.NewWaiverRepository(container.Queries.WaiversDb),
		DB:                     container.DB,
		HubSpotService:         container.HubspotService,
	}
}

func (s *AccountRegistrationService) CreateAccount(
	ctx context.Context,
	userPasswordCreate *values.UserPasswordCreate,
	staffCreate *values.StaffCreate,
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

	if err := staffCreate.Validate(); err != nil {
		tx.Rollback()
		return nil, err
	}

	role := staffCreate.Role
	isActive := staffCreate.IsActive

	// Insert into staff table if role is not ""

	if role != "" {

		roleExists := false

		dbStaffRoles, err := s.StaffRepository.GetStaffRolesTx(ctx, tx)
		staffRoles := []string{}

		if err != nil {
			tx.Rollback()
			return nil, err
		}

		for _, staffRole := range dbStaffRoles {
			staffRoles = append(staffRoles, staffRole.RoleName)
			if staffRole.RoleName == role {
				roleExists = true
			}
		}

		if !roleExists {
			tx.Rollback()
			return nil, errLib.New("Role does not exist. Available roles: "+strings.Join(staffRoles, ", "), http.StatusBadRequest)
		}

		if err := s.StaffRepository.CreateStaffTx(ctx, tx, email, role, isActive); err != nil {
			_ = tx.Rollback()
			return nil, err
		}

	} else {
		// User is Customer

		if err := waiverCreate.Validate(); err != nil {
			return nil, err
		}

		_, err := s.WaiverRepository.GetWaiver(ctx, waiverCreate.WaiverUrl)

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
	}

	// // Commit the transaction
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

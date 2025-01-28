package staff

import (
	"api/cmd/server/di"
	"api/internal/domains/identity/customer_registration/values"
	"api/internal/domains/identity/entities"
	"api/internal/domains/identity/infra/persistence/repository"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"net/http"
	"strings"
)

type StaffService struct {
	UsersRepository        *repository.UserRepository
	UserPasswordRepository *repository.UserCredentialsRepository
	StaffRepository        *repository.StaffRepository
	DB                     *sql.DB
}

func NewStaffService(
	container *di.Container,
) *StaffService {
	return &StaffService{
		UsersRepository:        repository.NewUserRepository(container.Queries.IdentityDb),
		UserPasswordRepository: repository.NewUserCredentialsRepository(container.Queries.IdentityDb),
		StaffRepository:        repository.NewStaffRepository(container.Queries.IdentityDb),
		DB:                     container.DB,
	}
}

func (s *StaffService) CreateAccount(
	ctx context.Context,
	userPasswordCreate *values.UserPasswordCreate,
	staffCreate *values.StaffCreate,
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
	_, err := s.UsersRepository.CreateUserTx(ctx, tx, email)

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

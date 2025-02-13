package identity

import (
	"api/internal/di"
	entity "api/internal/domains/identity/entities"
	"api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
)

type StaffRegistrationService struct {
	AccountRegistrationService *AccountCreationService
	UserOptionalInfoService    *UserOptionalInfoService
	StaffRepository            *repository.StaffRepository
	DB                         *sql.DB
}

func NewStaffRegistrationService(
	container *di.Container,
) *StaffRegistrationService {

	accountRegistrationService := NewAccountCreationService(container)

	return &StaffRegistrationService{
		StaffRepository:            repository.NewStaffRepository(container),
		UserOptionalInfoService:    NewUserOptionalInfoService(container),
		DB:                         container.DB,
		AccountRegistrationService: accountRegistrationService,
	}
}

func (s *StaffRegistrationService) RegisterStaff(
	ctx context.Context,
	staffDetails *values.StaffRegistrationInfo,
) (*entity.UserInfo, *errLib.CommonError) {

	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		log.Println("Failed to begin transaction. Error: ", txErr)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	// Ensure rollback if something goes wrong
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	_, _, err := s.AccountRegistrationService.CreateAccount(ctx, tx, staffDetails.Email, false)

	if err != nil {
		tx.Rollback()
		log.Println("Failed to create account. Error: ", err)
		return nil, err
	}

	if _, err := s.UserOptionalInfoService.CreateUserOptionalInfoTx(ctx, tx, staffDetails.UserInfo, nil); err != nil {
		tx.Rollback()
		log.Println("Failed to create user optional info. Error: ", err)
		return nil, err
	}

	role := staffDetails.RoleName
	isActive := staffDetails.IsActive

	roleExists := false

	dbStaffRoles, err := s.StaffRepository.GetStaffRolesTx(ctx, tx)
	staffRoles := []string{}

	if err != nil {
		log.Println("Failed to get staff roles. Error: ", err)
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

	email := staffDetails.Email

	if err := s.StaffRepository.AssignStaffRoleAndStatusTx(ctx, tx, email, role, isActive); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	userInfo := entity.UserInfo{
		FirstName: staffDetails.UserInfo.FirstName,
		LastName:  staffDetails.UserInfo.LastName,
		Email:     email,
		StaffInfo: &entity.StaffInfo{
			Role:     role,
			IsActive: isActive,
		},
	}

	return &userInfo, nil
}

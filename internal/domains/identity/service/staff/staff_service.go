package staff

import (
	"api/internal/di"
	staffRepo "api/internal/domains/identity/persistence/repository/staff"
	"api/internal/domains/identity/persistence/repository/user"
	staffValues "api/internal/domains/user/values/staff"
	"api/internal/services/hubspot"
	"github.com/google/uuid"

	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
)

type RegistrationService struct {
	HubSpotService  *hubspot.Service
	UsersRepository user.IRepository
	StaffRepository staffRepo.RepositoryInterface
	DB              *sql.DB
}

func NewStaffRegistrationService(
	container *di.Container,
) *RegistrationService {

	return &RegistrationService{
		StaffRepository: staffRepo.NewStaffRepository(container),
		DB:              container.DB,
		UsersRepository: user.NewUserRepository(container),
		HubSpotService:  container.HubspotService,
	}
}

func (s *RegistrationService) GetStaffInfo(ctx context.Context, userId uuid.UUID) *staffValues.ReadValues {

	staff, err := s.StaffRepository.GetStaffByUserId(ctx, userId)

	if err != nil {
		return nil
	}

	return &staffValues.ReadValues{
		RoleName: staff.RoleName,
		IsActive: staff.IsActive,
	}
}

func (s *RegistrationService) RegisterStaff(
	ctx context.Context,
	staffDetails *identity.StaffRegistrationRequestInfo,
) *errLib.CommonError {

	_, err := s.HubSpotService.GetUserById(staffDetails.HubSpotID)

	if err != nil {
		return err
	}

	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		log.Println("Failed to begin transaction. Error: ", txErr)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	// Ensure rollback if something goes wrong
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	userId, err := s.UsersRepository.CreateUserTx(ctx, tx, staffDetails.HubSpotID)

	if err != nil {
		tx.Rollback()
		log.Println("Failed to create account. Error: ", err)
		return err
	}

	role := staffDetails.RoleName
	isActive := staffDetails.IsActive

	roleExists := false

	dbStaffRoles, err := s.StaffRepository.GetStaffRolesTx(ctx, tx)
	var staffRoles []string

	if err != nil {
		log.Println("Failed to get repository roles. Error: ", err)
		tx.Rollback()
		return err
	}

	for _, staffRole := range dbStaffRoles {
		staffRoles = append(staffRoles, staffRole)
		if staffRole == role {
			roleExists = true
		}
	}

	if !roleExists {
		tx.Rollback()
		return errLib.New("RoleName does not exist. Available roles: "+strings.Join(staffRoles, ", "), http.StatusBadRequest)
	}

	if err := s.StaffRepository.AssignStaffRoleAndStatusTx(ctx, tx, *userId, role, isActive); err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	return nil
}

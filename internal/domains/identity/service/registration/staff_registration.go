package registration

import (
	"api/internal/di"
	identityRepo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/persistence/repository/user"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/services/hubspot"
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
)

type StaffsRegistrationService struct {
	HubSpotService  *hubspot.Service
	UsersRepository *user.UsersRepository
	StaffRepository *identityRepo.StaffRepository
	DB              *sql.DB
}

func NewStaffRegistrationService(
	container *di.Container,
) *StaffsRegistrationService {

	identityDb := container.Queries.IdentityDb
	outboxDb := container.Queries.OutboxDb

	return &StaffsRegistrationService{
		StaffRepository: identityRepo.NewStaffRepository(identityDb, outboxDb),
		DB:              container.DB,
		UsersRepository: user.NewUserRepository(identityDb, outboxDb),
		HubSpotService:  container.HubspotService,
	}
}

func (s *StaffsRegistrationService) RegisterPendingStaff(
	ctx context.Context,
	staffDetails identity.PendingStaffRegistrationRequestInfo,
) *errLib.CommonError {

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

	role := staffDetails.RoleName

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

	if err = s.StaffRepository.CreatePendingStaff(ctx, tx, staffDetails); err != nil {
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

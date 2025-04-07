package registration

import (
	"api/internal/di"
	identityRepo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/persistence/repository/user"
	identity "api/internal/domains/identity/values"
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
		StaffRepository: identityRepo.NewStaffRepository(identityDb),
		DB:              container.DB,
		UsersRepository: user.NewUserRepository(identityDb, outboxDb),
		HubSpotService:  container.HubspotService,
	}
}

func (s *StaffsRegistrationService) RegisterPendingStaff(
	ctx context.Context,
	staffDetails identity.StaffRegistrationRequestInfo,
) *errLib.CommonError {

	role := staffDetails.RoleName

	roleExists := false

	dbStaffRoles, err := s.StaffRepository.GetStaffRolesTx(ctx)
	var staffRoles []string

	if err != nil {
		log.Println("Failed to get repository roles. Error: ", err)
		return err
	}

	for _, staffRole := range dbStaffRoles {
		staffRoles = append(staffRoles, staffRole)
		if staffRole == role {
			roleExists = true
		}
	}

	if !roleExists {
		return errLib.New("RoleName does not exist. Available roles: "+strings.Join(staffRoles, ", "), http.StatusBadRequest)
	}

	if err = s.StaffRepository.CreatePendingStaff(ctx, staffDetails); err != nil {
		return err
	}

	return nil
}

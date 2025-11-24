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
	"api/utils/email"

	"github.com/google/uuid"
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

	return &StaffsRegistrationService{
		StaffRepository: identityRepo.NewStaffRepository(container),
		DB:              container.DB,
		UsersRepository: user.NewUserRepository(container),
		HubSpotService:  container.HubspotService,
	}
}

func (s *StaffsRegistrationService) RegisterPendingStaff(
	ctx context.Context,
	staffDetails identity.StaffRegistrationRequestInfo,
) *errLib.CommonError {

	role := staffDetails.RoleName

	roleExists := false

	dbStaffRoles, err := s.StaffRepository.GetStaffRoles(ctx)
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
	
	email.SendSignUpConfirmationEmail(staffDetails.Email, staffDetails.FirstName)

	return nil
}

func (s *StaffsRegistrationService) ApproveStaff(
	ctx context.Context,
	id uuid.UUID,
) *errLib.CommonError {

	if err := s.StaffRepository.ApproveStaff(ctx, id); err != nil {
		return err
	}

	return nil
}
func (s *StaffsRegistrationService) GetPendingStaffs(ctx context.Context) ([]identity.PendingStaffInfo, *errLib.CommonError) {
	return s.StaffRepository.GetPendingStaffs(ctx)
}

func (s *StaffsRegistrationService) DeletePendingStaff(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.StaffRepository.DeletePendingStaff(ctx, id)
}

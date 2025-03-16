package staff

import (
	"api/internal/di"
	identityRepo "api/internal/domains/identity/persistence/repository"
	staffValues "api/internal/domains/user/values"
	"github.com/google/uuid"

	"context"
)

type RegistrationService struct {
	StaffRepository *identityRepo.StaffRepository
}

func NewStaffRegistrationService(
	container *di.Container,
) *RegistrationService {

	return &RegistrationService{
		StaffRepository: identityRepo.NewStaffRepository(container.Queries.IdentityDb, container.Queries.OutboxDb),
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

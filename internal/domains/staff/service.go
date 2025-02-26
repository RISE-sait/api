package staff

import (
	"api/internal/di"
	entity "api/internal/domains/staff/entity"
	repository "api/internal/domains/staff/persistence"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Service struct {
	StaffRepository *repository.Repository
	DB              *sql.DB
}

func NewStaffService(
	container *di.Container,
) *Service {
	return &Service{
		StaffRepository: repository.NewStaffRepository(container),
		DB:              container.DB,
	}
}

func (s *Service) GetStaffs(c context.Context, roleIdPtr *uuid.UUID) ([]entity.Staff, *errLib.CommonError) {

	staffs, err := s.StaffRepository.List(c, roleIdPtr)

	if err != nil {
		return nil, err
	}

	return staffs, nil
}

func (s *Service) GetByID(c context.Context, id uuid.UUID) (*entity.Staff, *errLib.CommonError) {

	// Get the staff details
	staff, err := s.StaffRepository.GetByID(c, id)

	if err != nil {
		return nil, err
	}

	return staff, nil
}

func (s *Service) DeleteStaff(c context.Context, id uuid.UUID) *errLib.CommonError {

	err := s.StaffRepository.Delete(c, id)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateStaff(c context.Context, input *entity.Staff) (*entity.Staff, *errLib.CommonError) {

	staff, err := s.StaffRepository.Update(c, input)

	if err != nil {
		return nil, err
	}

	return &staff, nil
}

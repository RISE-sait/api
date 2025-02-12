package staff

import (
	"api/internal/di"
	repository "api/internal/domains/staff/persistence"
	"api/internal/domains/staff/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type StaffService struct {
	StaffRepository *repository.StaffRepository
	DB              *sql.DB
}

func NewStaffService(
	container *di.Container,
) *StaffService {
	return &StaffService{
		StaffRepository: repository.NewStaffRepository(container),
		DB:              container.DB,
	}
}

func (s *StaffService) GetStaffs(c context.Context, roleIdPtr *uuid.UUID) ([]values.StaffAllFields, *errLib.CommonError) {

	staffs, err := s.StaffRepository.List(c, roleIdPtr)

	if err != nil {
		return nil, err
	}

	return staffs, nil
}

func (s *StaffService) GetByID(c context.Context, id uuid.UUID) (*values.StaffAllFields, *errLib.CommonError) {

	// Get the staff details
	staff, err := s.StaffRepository.GetByID(c, id)

	if err != nil {
		return nil, err
	}

	return staff, nil
}

func (s *StaffService) DeleteStaff(c context.Context, id uuid.UUID) *errLib.CommonError {

	err := s.StaffRepository.Delete(c, id)

	if err != nil {
		return err
	}

	return nil
}

func (s *StaffService) UpdateStaff(c context.Context, input *values.StaffAllFields) (*values.StaffAllFields, *errLib.CommonError) {

	staff, err := s.StaffRepository.Update(c, input)

	if err != nil {
		return nil, err
	}

	return &staff, nil
}

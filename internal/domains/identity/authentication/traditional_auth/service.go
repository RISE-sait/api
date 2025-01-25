package traditional_auth

import (
	"api/internal/domains/identity/authentication/infra/repository"
	"api/internal/domains/identity/entities"
	"api/internal/domains/identity/lib"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"net/http"
	"strings"
)

type Service struct {
	UserRepo  *repository.UserRepository
	StaffRepo *repository.StaffRepository
}

func NewService(userRepo *repository.UserRepository, staffRepo *repository.StaffRepository) *Service {
	return &Service{
		UserRepo:  userRepo,
		StaffRepo: staffRepo,
	}
}

func (s *Service) AuthenticateUser(ctx context.Context, input *values.Credentials) (string, *errLib.CommonError) {

	if err := input.Validate(); err != nil {
		return "", err
	}

	email := (*input).Email
	password := (*input).Password

	if !s.UserRepo.IsValidUser(ctx, email, password) {
		return "", errLib.New("Invalid credentials", http.StatusUnauthorized)
	}

	name := s.getNameFromEmail(email)
	staffInfo := s.getStaffInfo(ctx, email)

	userInfo := entities.UserInfo{
		Email:     email,
		Name:      name,
		StaffInfo: &staffInfo,
	}

	return lib.SignJWT(userInfo)
}

func (s *Service) getNameFromEmail(email string) string {
	return strings.Split(email, "@")[0]
}

func (s *Service) getStaffInfo(ctx context.Context, email string) entities.StaffInfo {
	staffInfo := entities.StaffInfo{Role: "Athlete", IsActive: false}

	staff, err := s.StaffRepo.GetStaffByEmail(ctx, email)
	if err == nil {
		staffInfo.Role = string(staff.Role)
		staffInfo.IsActive = staff.IsActive
	}

	return staffInfo
}

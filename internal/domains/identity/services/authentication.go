package identity

import (
	"api/cmd/server/di"
	"api/internal/domains/identity/entities"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"net/http"
	"strings"
)

type AuthenticationService struct {
	UserRepo  *repo.UserRepository
	StaffRepo *repo.StaffRepository
}

func NewAuthenticationService(container *di.Container) *AuthenticationService {

	userRepo := repo.NewUserRepository(container)
	staffRepo := repo.NewStaffRepository(container.Queries.IdentityDb)

	return &AuthenticationService{
		UserRepo:  userRepo,
		StaffRepo: staffRepo,
	}
}

func (s *AuthenticationService) AuthenticateUser(ctx context.Context, credentials values.LoginCredentials) (*entities.UserInfo, *errLib.CommonError) {

	email := credentials.Email
	passwordStr := credentials.Password

	if !s.UserRepo.IsValidUser(ctx, email, passwordStr) {
		return nil, errLib.New("Person with the credentials not found", http.StatusUnauthorized)
	}

	name := s.getNameFromEmail(email)
	staffInfo := s.getStaffInfo(ctx, email)

	return &entities.UserInfo{
		Email:     email,
		Name:      name,
		StaffInfo: &staffInfo,
	}, nil
}

func (s *AuthenticationService) getNameFromEmail(email string) string {
	return strings.Split(email, "@")[0]
}

func (s *AuthenticationService) getStaffInfo(ctx context.Context, email string) entities.StaffInfo {
	staffInfo := entities.StaffInfo{Role: "Athlete", IsActive: false}

	staff, err := s.StaffRepo.GetStaffByEmail(ctx, email)
	if err == nil {
		staffInfo.Role = staff.RoleName
		staffInfo.IsActive = staff.IsActive
	}

	return staffInfo
}

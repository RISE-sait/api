package identity

import (
	"api/internal/di"
	entity "api/internal/domains/identity/entities"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"net/http"
)

type AuthenticationService struct {
	UserOptionalInfoRepository *repo.UserOptionalInfoRepository
	StaffRepo                  *repo.StaffRepository
}

func NewAuthenticationService(container *di.Container) *AuthenticationService {

	userOptionalInfoRepo := repo.NewUserOptionalInfoRepository(container)
	staffRepo := repo.NewStaffRepository(container)

	return &AuthenticationService{
		UserOptionalInfoRepository: userOptionalInfoRepo,
		StaffRepo:                  staffRepo,
	}
}

func (s *AuthenticationService) AuthenticateUser(ctx context.Context, credentials values.LoginCredentials) (*entity.UserInfo, *errLib.CommonError) {

	email := credentials.Email
	passwordStr := credentials.Password

	user := s.UserOptionalInfoRepository.GetUser(ctx, email, passwordStr)

	if user == nil {
		return nil, errLib.New("Person with the credentials not found", http.StatusUnauthorized)
	}

	staffInfo := s.getStaffInfo(ctx, email)

	return &entity.UserInfo{
		Email:     email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		StaffInfo: staffInfo,
	}, nil
}

func (s *AuthenticationService) getStaffInfo(ctx context.Context, email string) *entity.StaffInfo {

	staff, err := s.StaffRepo.GetStaffByEmail(ctx, email)

	if err != nil {
		return nil
	}

	return &entity.StaffInfo{
		Role:     staff.RoleName,
		IsActive: staff.IsActive,
	}
}

package authentication

import (
	"api/internal/di"
	staffRepo "api/internal/domains/identity/persistence/repository/staff"
	"api/internal/domains/identity/persistence/repository/user"
	userInfoTempRepo "api/internal/domains/identity/persistence/repository/user_info"
	"api/internal/domains/identity/service/firebase"
	errLib "api/internal/libs/errors"
	"api/internal/libs/jwt"
	"api/internal/services/hubspot"
	"context"
	"github.com/google/uuid"
	"log"
)

type Service struct {
	FirebaseService  *firebase.Service
	HubSpotService   *hubspot.Service
	UserRepo         user.RepositoryInterface
	StaffRepo        staffRepo.RepositoryInterface
	UserInfoTempRepo userInfoTempRepo.InfoTempRepositoryInterface
}

func NewAuthenticationService(container *di.Container) *Service {

	return &Service{
		FirebaseService:  firebase.NewFirebaseService(container),
		UserRepo:         user.NewUserRepository(container),
		StaffRepo:        staffRepo.NewStaffRepository(container),
		UserInfoTempRepo: userInfoTempRepo.NewInfoTempRepository(container),
		HubSpotService:   container.HubspotService,
	}
}

// AuthenticateUser authenticates a user using their Firebase ID token.
// It retrieves user id and staff role, if applicable, from database, and generates a JWT token.
//
// Parameters:
//   - ctx: The request context.
//   - idToken: The Firebase ID token used for authentication.
//
// Returns:
//   - *entity.UserInfo: The authenticated user's information.
//   - *string: The signed JWT token for authentication.
//   - *errLib.CommonError: An error if authentication fails.
func (s *Service) AuthenticateUser(ctx context.Context, idToken string) (*string, *errLib.CommonError) {

	firebaseUserInfo, err := s.FirebaseService.GetUserInfo(ctx, idToken)

	if err != nil {
		log.Println(err.Message)
		return nil, err
	}

	email := firebaseUserInfo.Email

	var userId uuid.UUID
	var hubspotId *string

	hubspotResponse, err := s.HubSpotService.GetUserByEmail(email)

	// If info is already on HubSpot
	if err == nil {
		hubspotId = &hubspotResponse.HubSpotId

		id, newErr := s.UserRepo.GetUserIdByHubspotId(ctx, *hubspotId)

		if newErr != nil {
			return nil, newErr
		}

		userId = *id
	} else {
		// info is not on HubSpot yet

		userTempInfo, err := s.UserInfoTempRepo.GetTempUserInfoByEmail(ctx, email)

		if err != nil {
			return nil, err
		}

		userId = userTempInfo.ID

	}

	jwtCustomClaims := jwtLib.CustomClaims{
		UserID:    userId,
		HubspotID: hubspotId,
	}

	staffInfoPtr, _ := s.StaffRepo.GetStaffByUserId(ctx, userId)

	if staffInfoPtr != nil {
		jwtCustomClaims.StaffInfo = &jwtLib.StaffInfo{
			Role:     (*staffInfoPtr).RoleName,
			IsActive: (*staffInfoPtr).IsActive,
		}
	}

	jwtToken, err := jwtLib.SignJWT(jwtCustomClaims)

	if err != nil {
		return nil, err
	}

	return &jwtToken, nil
}

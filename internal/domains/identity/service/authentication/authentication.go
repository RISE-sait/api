package authentication

import (
	"api/internal/di"
	userInfoTempRepo "api/internal/domains/identity/persistence/repository/pending_users"
	staffRepo "api/internal/domains/identity/persistence/repository/staff"
	"api/internal/domains/identity/persistence/repository/user"
	"api/internal/domains/identity/service/firebase"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/jwt"
	"api/internal/services/hubspot"
	"context"
	"log"
	"net/http"
)

type Service struct {
	FirebaseService  *firebase.Service
	HubSpotService   *hubspot.Service
	UserRepo         *user.Repository
	StaffRepo        *staffRepo.Repository
	UserInfoTempRepo *userInfoTempRepo.PendingUsersRepo
}

func NewAuthenticationService(container *di.Container) *Service {

	return &Service{
		FirebaseService:  firebase.NewFirebaseService(container),
		UserRepo:         user.NewUserRepository(container),
		StaffRepo:        staffRepo.NewStaffRepository(container),
		UserInfoTempRepo: userInfoTempRepo.NewPendingUserInfoRepository(container),
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
//   - string: The signed JWT token for authentication.
//   - *errLib.CommonError: An error if authentication fails.
func (s *Service) AuthenticateUser(ctx context.Context, idToken string) (string, identity.UserAuthenticationResponseInfo, *errLib.CommonError) {

	var userInfo identity.UserAuthenticationResponseInfo

	email, err := s.FirebaseService.GetUserEmail(ctx, idToken)

	if err != nil {
		log.Println(err.Message)
		return "", userInfo, err
	}

	hubspotResponse, err := s.HubSpotService.GetUserByEmail(email)

	if err != nil {
		return "", userInfo, err
	}

	hubspotId := hubspotResponse.HubSpotId

	userId, newErr := s.UserRepo.GetUserIdByHubspotId(ctx, hubspotId)

	if newErr != nil {
		return "", userInfo, newErr
	}

	jwtCustomClaims := jwtLib.CustomClaims{
		UserID:    userId,
		HubspotID: hubspotId,
	}

	userInfo = identity.UserAuthenticationResponseInfo{
		//Age:       hubspotResponse.Properties.,
		FirstName: hubspotResponse.Properties.FirstName,
		LastName:  hubspotResponse.Properties.LastName,
		Role:      "Athlete",
	}

	staffInfo, err := s.StaffRepo.GetStaffByUserId(ctx, userId)

	if err == nil {
		jwtCustomClaims.StaffInfo = &jwtLib.StaffInfo{
			Role:     staffInfo.RoleName,
			IsActive: staffInfo.IsActive,
		}

		userInfo.Role = staffInfo.RoleName
	}

	jwtToken, err := jwtLib.SignJWT(jwtCustomClaims)

	if err != nil {
		return "", userInfo, err
	}

	return jwtToken, userInfo, nil
}

// AuthenticateChild authenticates a child user by verifying their association with a parent user in HubSpot.
// It retrieves the child's user ID from the database and generates a JWT token.
//
// Parameters:
//   - ctx: The request context.
//   - childHubspotId: The HubSpot ID of the child user.
//   - parentHubspotId: The HubSpot ID of the parent user.
//
// Returns:
//   - string: The signed JWT token for authentication.
//   - *errLib.CommonError: An error if authentication fails.
func (s *Service) AuthenticateChild(ctx context.Context, childHubspotId, parentHubspotId string) (string, identity.UserRegistrationRequestNecessaryInfo, *errLib.CommonError) {

	var userInfo identity.UserRegistrationRequestNecessaryInfo

	parentInfo, err := s.HubSpotService.GetUserById(parentHubspotId)

	if err != nil {
		return "", userInfo, err
	}

	if !isChildAssociated(parentInfo.Associations.Contact.Result, childHubspotId) {
		return "", userInfo, errLib.New("child is not associated with the parent", http.StatusNotFound)
	}

	childId, newErr := s.UserRepo.GetUserIdByHubspotId(ctx, childHubspotId)

	if newErr != nil {
		return "", userInfo, newErr
	}

	childInfo, err := s.HubSpotService.GetUserById(childHubspotId)

	if err != nil {
		return "", userInfo, err
	}

	userInfo = identity.UserRegistrationRequestNecessaryInfo{
		FirstName: childInfo.Properties.FirstName,
		LastName:  childInfo.Properties.LastName,
	}

	jwtCustomClaims := jwtLib.CustomClaims{
		UserID:    childId,
		HubspotID: childHubspotId,
	}

	jwtToken, err := jwtLib.SignJWT(jwtCustomClaims)

	if err != nil {
		return "", userInfo, err
	}

	return jwtToken, userInfo, nil
}

func isChildAssociated(contacts []hubspot.UserAssociationResult, childHubspotId string) bool {
	for _, contact := range contacts {
		if contact.Type == "child_parent" && contact.ID == childHubspotId {
			return true
		}
	}
	return false
}

package authentication

import (
	"api/internal/di"
	identityRepo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/persistence/repository/user"
	"api/internal/domains/identity/service/firebase"
	identity "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	jwtLib "api/internal/libs/jwt"
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Service struct {
	FirebaseService *firebase.Service
	UserRepo        *user.UsersRepository
	StaffRepo       *identityRepo.StaffRepository
}

func NewAuthenticationService(container *di.Container) *Service {

	identityDb := container.Queries.IdentityDb
	outboxDb := container.Queries.OutboxDb

	return &Service{
		FirebaseService: firebase.NewFirebaseService(container),
		UserRepo:        user.NewUserRepository(identityDb, outboxDb),
		StaffRepo:       identityRepo.NewStaffRepository(identityDb),
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
func (s *Service) AuthenticateUser(ctx context.Context, idToken string) (string, identity.UserReadInfo, *errLib.CommonError) {

	var responseUserInfo identity.UserReadInfo

	email, err := s.FirebaseService.GetUserEmail(ctx, idToken)

	if err != nil {
		log.Println(err.Message)
		return "", responseUserInfo, err
	}

	userInfo, err := s.UserRepo.GetUserInfo(ctx, email, uuid.Nil)

	if err != nil {
		return "", responseUserInfo, err
	}

	jwtCustomClaims := jwtLib.CustomClaims{
		UserID: userInfo.ID,
		RoleInfo: &jwtLib.RoleInfo{
			Role: userInfo.Role,
		},
	}

	if userInfo.IsActiveStaff != nil {
		jwtCustomClaims.IsActiveStaff = userInfo.IsActiveStaff
	}

	jwtToken, err := jwtLib.SignJWT(jwtCustomClaims)

	if err != nil {
		return "", responseUserInfo, err
	}

	return jwtToken, userInfo, nil
}

// AuthenticateChild authenticates a child user by verifying their association with a parent user.
// It retrieves the child's user ID from the database and generates a JWT token.
//
// Parameters:
//   - ctx: The request context.
//   - childId: The ID of the child user.
//   - parentEmail: The email of the parent user.
//
// Returns:
//   - string: The signed JWT token for authentication.
//   - *errLib.CommonError: An error if authentication fails.
func (s *Service) AuthenticateChild(ctx context.Context, childId, parentID uuid.UUID) (string, identity.UserReadInfo, *errLib.CommonError) {

	var userInfo identity.UserReadInfo

	if isConnected, err := s.UserRepo.GetIsActualParentChild(ctx, childId, parentID); err != nil {
		return "", userInfo, err
	} else if !isConnected {
		return "", userInfo, errLib.New("child is not associated with the parent", http.StatusNotFound)
	}

	userInfo, err := s.UserRepo.GetUserInfo(ctx, "", childId)

	if err != nil {
		return "", userInfo, err
	}

	jwtCustomClaims := jwtLib.CustomClaims{
		UserID: childId,
	}

	jwtToken, err := jwtLib.SignJWT(jwtCustomClaims)

	if err != nil {
		return "", userInfo, err
	}

	return jwtToken, userInfo, nil
}

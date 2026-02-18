package authentication

import (
	"api/internal/di"
	identityRepo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/persistence/repository/user"
	"api/internal/domains/identity/service/email_verification"
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
	FirebaseService     *firebase.Service
	UserRepo            *user.UsersRepository
	StaffRepo           *identityRepo.StaffRepository
	VerificationService *email_verification.EmailVerificationService
}

func NewAuthenticationService(container *di.Container) *Service {

	return &Service{
		FirebaseService:     firebase.NewFirebaseService(container),
		UserRepo:            user.NewUserRepository(container),
		StaffRepo:           identityRepo.NewStaffRepository(container),
		VerificationService: email_verification.NewEmailVerificationService(container),
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

	// Check if email is verified (except for staff who may have been manually created)
	if userInfo.Role == "customer" || userInfo.Role == "athlete" || userInfo.Role == "parent" {
		isVerified, verifyErr := s.VerificationService.IsEmailVerified(ctx, userInfo.ID)
		if verifyErr != nil {
			log.Printf("Failed to check email verification status for user %s: %v", userInfo.ID, verifyErr)
			// Continue with authentication if check fails (graceful degradation)
		} else if !isVerified {
			log.Printf("Login blocked for unverified user %s (email: %s)", userInfo.ID, email)
			return "", responseUserInfo, errLib.New("Please verify your email address before logging in. Check your inbox for the verification link.", http.StatusForbidden)
		}
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
		RoleInfo: &jwtLib.RoleInfo{
			Role: userInfo.Role,
		},
	}

	jwtToken, err := jwtLib.SignJWT(jwtCustomClaims)

	if err != nil {
		return "", userInfo, err
	}

	return jwtToken, userInfo, nil
}

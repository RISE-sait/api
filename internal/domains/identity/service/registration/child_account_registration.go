package registration

import (
	"api/internal/di"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/persistence/repository/user"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"github.com/google/uuid"
	"log"
	"net/http"
	"api/utils/email"
)

type ChildRegistrationService struct {
	UsersRepository         *user.UsersRepository
	WaiverSigningRepository *repo.WaiverSigningRepository
	DB                      *sql.DB
}

func NewChildAccountRegistrationService(
	container *di.Container,
) *ChildRegistrationService {

	return &ChildRegistrationService{
		UsersRepository:         user.NewUserRepository(container),
		WaiverSigningRepository: repo.NewWaiverSigningRepository(container),
		DB:                      container.DB,
	}
}

func (s *ChildRegistrationService) CreateChildAccount(
	ctx context.Context,
	childRegistrationInfo identity.ChildRegistrationRequestInfo,
) (identity.UserReadInfo, *errLib.CommonError) {

	requiredWaivers, err := s.WaiverSigningRepository.GetRequiredWaivers(ctx)

	if err != nil {
		return identity.UserReadInfo{}, err
	}

	if err = validateWaivers(childRegistrationInfo.Waivers, requiredWaivers); err != nil {
		return identity.UserReadInfo{}, err
	}

	waiverUrls, areWaiversSigned := prepareWaiverData(requiredWaivers)

	var response identity.UserReadInfo

	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		return response, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	createdChild, err := s.UsersRepository.CreateChildTx(ctx, tx, childRegistrationInfo)

	if err != nil {
		tx.Rollback()
		return response, err
	}

	var childIds []uuid.UUID

	for range waiverUrls {
		childIds = append(childIds, createdChild.ID)
	}

	if err = s.WaiverSigningRepository.CreateWaiversSigningRecordTx(ctx, tx, childIds, waiverUrls, areWaiversSigned); err != nil {
		tx.Rollback()
		return response, err
	}

	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Failed to commit transaction: %v", txErr)
		return response, errLib.New("Failed to commit transaction but event is logged", http.StatusInternalServerError)
	}
		if createdChild.Email != nil {
		email.SendSignUpConfirmationEmail(*createdChild.Email, createdChild.FirstName)
	}


	return createdChild, nil

}

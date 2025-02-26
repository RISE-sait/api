package registration

import (
	"api/internal/di"
	"api/internal/domains/identity/persistence/repository/user"
	tempUserInfo "api/internal/domains/identity/persistence/repository/user_info"

	repo "api/internal/domains/identity/persistence/repository/waiver_signing"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/internal/services/hubspot"
	"context"
	"database/sql"
	"log"
	"net/http"
)

type ChildRegistrationService struct {
	HubSpotService          *hubspot.Service
	UserRepository          user.RepositoryInterface
	TempUserInfoRepo        tempUserInfo.InfoTempRepositoryInterface
	WaiverSigningRepository repo.RepositoryInterface
	DB                      *sql.DB
}

func NewChildAccountRegistrationService(
	container *di.Container,
) *ChildRegistrationService {
	return &ChildRegistrationService{
		UserRepository:          user.NewUserRepository(container),
		TempUserInfoRepo:        tempUserInfo.NewInfoTempRepository(container),
		WaiverSigningRepository: repo.NewWaiverSigningRepository(container),
		DB:                      container.DB,
		HubSpotService:          container.HubspotService,
	}
}

func (s *ChildRegistrationService) CreateChildAccount(
	ctx context.Context,
	childRegistrationInfo *values.ChildRegistrationInfo,
) *errLib.CommonError {

	for _, waiver := range childRegistrationInfo.Waivers {
		if !waiver.IsWaiverSigned {
			return errLib.New("Waiver is not signed", http.StatusBadRequest)
		}
	}

	tx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
	if txErr != nil {
		return errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	userId, err := s.UserRepository.CreateUserTx(ctx, tx)

	if err != nil {
		tx.Rollback()
		return err
	}

	for _, waiver := range childRegistrationInfo.Waivers {
		if err := s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, *userId, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {
			tx.Rollback()
			return err
		}

	}

	parent, err := s.HubSpotService.GetUserByEmail(childRegistrationInfo.ParentEmail)

	if err != nil {
		tx.Rollback()
		return err
	}

	err = s.TempUserInfoRepo.CreateTempUserInfoTx(ctx, tx, *userId, childRegistrationInfo.FirstName, childRegistrationInfo.LastName, nil, &parent.Properties.Email, childRegistrationInfo.Age)

	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return errLib.New("Failed to commit transaction but event is logged", http.StatusInternalServerError)
	}

	return nil

}

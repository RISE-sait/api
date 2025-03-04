package registration

import (
	"api/internal/di"
	pendingUsers "api/internal/domains/identity/persistence/repository/pending_users"
	waiverSigningRepo "api/internal/domains/identity/persistence/repository/waiver_signing"
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
	PendingUsersRepository  pendingUsers.IPendingUsersRepository
	WaiverSigningRepository waiverSigningRepo.IRepository
	DB                      *sql.DB
}

func NewChildAccountRegistrationService(
	container *di.Container,
) *ChildRegistrationService {
	return &ChildRegistrationService{
		PendingUsersRepository:  pendingUsers.NewPendingUserInfoRepository(container),
		WaiverSigningRepository: waiverSigningRepo.NewPendingUserWaiverSigningRepository(container),
		DB:                      container.DB,
		HubSpotService:          container.HubspotService,
	}
}

func (s *ChildRegistrationService) CreateChildAccount(
	ctx context.Context,
	childRegistrationInfo *identity.ChildRegistrationRequestInfo,
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

	parent, err := s.HubSpotService.GetUserByEmail(childRegistrationInfo.ParentEmail)

	if err != nil {
		tx.Rollback()
		return err
	}

	childId, err := s.PendingUsersRepository.CreatePendingUserInfoTx(ctx, tx, childRegistrationInfo.FirstName, childRegistrationInfo.LastName, nil, &parent.HubSpotId, childRegistrationInfo.Age)

	if err != nil {
		tx.Rollback()
		return err
	}

	for _, waiver := range childRegistrationInfo.Waivers {
		if err := s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, childId, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return errLib.New("Failed to commit transaction but event is logged", http.StatusInternalServerError)
	}

	return nil

}

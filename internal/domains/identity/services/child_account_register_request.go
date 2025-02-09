package identity

import (
	"api/internal/di"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"net/http"
)

type ChildAccountRequestService struct {
	PendingChildAccountRepository *repo.PendingChildAccountRepository
	WaiverSigningRepository       *repo.PendingChildAccountWaiverSigningRepository
	DB                            *sql.DB
}

func NewChildAccountRegistrationRequestService(
	container *di.Container,
) *ChildAccountRequestService {
	return &ChildAccountRequestService{
		PendingChildAccountRepository: repo.NewPendingChildAcountRepository(container),
		WaiverSigningRepository:       repo.NewPendingChildAccountWaiverSigningRepository(container),
		DB:                            container.DB,
	}
}

func (s *ChildAccountRequestService) CreatePendingAccount(
	ctx context.Context,
	tx *sql.Tx,
	childAccountCreate *values.CreatePendingChildAccountValueObject,
) *errLib.CommonError {

	childEmail := childAccountCreate.Email

	for _, waiver := range childAccountCreate.Waivers {
		if !waiver.IsWaiverSigned {
			return errLib.New("Waiver is not signed", http.StatusBadRequest)
		}
	}

	if tx == nil {
		newTx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
		if txErr != nil {
			return errLib.New("Failed to begin transaction", http.StatusInternalServerError)
		}

		tx = newTx
	}

	for _, waiver := range childAccountCreate.Waivers {
		if err := s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, childEmail, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {

			tx.Rollback()
			return err
		}

	}

	return nil

}

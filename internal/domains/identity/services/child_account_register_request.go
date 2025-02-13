package identity

import (
	"api/internal/di"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/utils/email"
	"context"
	"database/sql"
	"log"
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

func (s *ChildAccountRequestService) CreatePendingChildAccount(
	ctx context.Context,
	tx *sql.Tx,
	childAccountCreate *values.CreatePendingChildAccountValueObject,
) *errLib.CommonError {

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

	child, err := s.PendingChildAccountRepository.CreatePendingChildAccountTx(ctx, tx, childAccountCreate)

	if err != nil {
		tx.Rollback()
		return err
	}

	for _, waiver := range childAccountCreate.Waivers {
		if err := s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, child.UserEmail, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {

			tx.Rollback()
			return err
		}

	}

	if err := email.SendConfirmChildEmail(childAccountCreate.ParentEmail, childAccountCreate.Email); err != nil {
		tx.Rollback()
		return errLib.New("Failed to send email", http.StatusInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return errLib.New("Failed to commit transaction but email is sent to parent", http.StatusInternalServerError)
	}

	return nil

}

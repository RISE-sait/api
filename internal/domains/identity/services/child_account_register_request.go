package identity

import (
	"api/cmd/server/di"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"api/utils/email"
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
	childAccountCreate *values.CreatePendingChildAccountValueObject,
) *errLib.CommonError {

	childEmail := childAccountCreate.ChildEmail
	parentEmail := childAccountCreate.ParentEmail
	password := childAccountCreate.Password

	for _, waiver := range childAccountCreate.Waivers {
		if !waiver.IsWaiverSigned {
			return errLib.New("Waiver is not signed", http.StatusBadRequest)
		}
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return errLib.New("Failed to start transaction", http.StatusInternalServerError)
	}

	if err := s.PendingChildAccountRepository.CreatePendingChildAccountTx(ctx, tx, childEmail, parentEmail, password); err != nil {
		tx.Rollback()
		return err
	}

	// for _, waiver := range childAccountCreate.Waivers {
	// 	if err := s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, childEmail, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {

	// 		tx.Rollback()
	// 		return err
	// 	}

	// }

	if err := email.SendConfirmChildEmail(parentEmail, childEmail); err != nil {
		tx.Rollback()
		return errLib.New("Failed to send email", http.StatusInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	return nil

}

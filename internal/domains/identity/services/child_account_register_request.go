package identity

import (
	"api/cmd/server/di"
	dto "api/internal/domains/identity/dto"
	repo "api/internal/domains/identity/persistence/repository"
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
	credentialsCreate *dto.Credentials,
	customerWaiverCreate *dto.CustomerWaiverCreateDto,
	childAccountCreate *dto.CreateChildAccountDto,
) *errLib.CommonError {

	if err := credentialsCreate.Validate(); err != nil {
		return err
	}

	if err := childAccountCreate.Validate(); err != nil {
		return err
	}

	if err := customerWaiverCreate.Validate(); err != nil {
		return err
	}

	if !customerWaiverCreate.IsWaiverSigned {
		return errLib.New("Waiver is not signed", http.StatusBadRequest)
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return errLib.New("Failed to start transaction", http.StatusInternalServerError)
	}

	if err := s.PendingChildAccountRepository.CreatePendingChildAccountTx(ctx, tx, credentialsCreate.Email, childAccountCreate.ParentEmail, credentialsCreate.Password); err != nil {
		tx.Rollback()
		return err
	}

	if err := s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, credentialsCreate.Email, customerWaiverCreate.WaiverUrl, customerWaiverCreate.IsWaiverSigned); err != nil {

		tx.Rollback()
		return err
	}

	if err := email.SendConfirmChildEmail(childAccountCreate.ParentEmail, credentialsCreate.Email); err != nil {
		tx.Rollback()
		return errLib.New("Failed to send email", http.StatusInternalServerError)
	}

	if err := tx.Commit(); err != nil {
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	return nil

}

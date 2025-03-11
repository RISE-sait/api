package registration

import (
	"api/internal/di"
	repo "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"log"
	"net/http"
)

type ChildRegistrationService struct {
	UsersRepository         *repo.UsersRepository
	WaiverSigningRepository *repo.WaiverSigningRepository
	DB                      *sql.DB
}

func NewChildAccountRegistrationService(
	container *di.Container,
) *ChildRegistrationService {

	identityDb := container.Queries.IdentityDb
	outboxDb := container.Queries.OutboxDb

	return &ChildRegistrationService{
		UsersRepository:         repo.NewUserRepository(identityDb, outboxDb),
		WaiverSigningRepository: repo.NewWaiverSigningRepository(container.Queries.IdentityDb),
		DB:                      container.DB,
	}
}

func (s *ChildRegistrationService) CreateChildAccount(
	ctx context.Context,
	childRegistrationInfo identity.ChildRegistrationRequestInfo,
) (identity.UserReadInfo, *errLib.CommonError) {

	var response identity.UserReadInfo

	for _, waiver := range childRegistrationInfo.Waivers {
		if !waiver.IsWaiverSigned {
			return response, errLib.New("Waiver is not signed", http.StatusBadRequest)
		}
	}

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

	for _, waiver := range childRegistrationInfo.Waivers {
		if err = s.WaiverSigningRepository.CreateWaiverSigningRecordTx(ctx, tx, createdChild.ID, waiver.WaiverUrl, waiver.IsWaiverSigned); err != nil {
			tx.Rollback()
			return response, err
		}
	}

	if txErr = tx.Commit(); txErr != nil {
		log.Printf("Failed to commit transaction: %v", txErr)
		return response, errLib.New("Failed to commit transaction but event is logged", http.StatusInternalServerError)
	}

	return createdChild, nil

}

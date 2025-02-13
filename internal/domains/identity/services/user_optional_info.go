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

type UserOptionalInfoService struct {
	UserOptionalInfoRepository *repo.UserOptionalInfoRepository
	DB                         *sql.DB
}

func NewUserOptionalInfoService(
	container *di.Container,
) *UserOptionalInfoService {
	return &UserOptionalInfoService{
		UserOptionalInfoRepository: repo.NewUserOptionalInfoRepository(container),
		DB:                         container.DB,
	}
}

func (s *UserOptionalInfoService) CreateUserOptionalInfoTx(
	ctx context.Context,
	tx *sql.Tx,
	userInfo values.UserInfo,
	pwd *string,
) (*sql.Tx, *errLib.CommonError) {

	// Begin transaction
	if tx == nil {
		newTx, txErr := s.DB.BeginTx(ctx, &sql.TxOptions{})
		if txErr != nil {
			return nil, errLib.New("Failed to begin transaction", http.StatusInternalServerError)
		}

		tx = newTx
	}

	// Ensure rollback if something goes wrong
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.UserOptionalInfoRepository.CreateUserOptionalInfoTx(ctx, tx, userInfo, pwd); err != nil {
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

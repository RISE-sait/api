package repositories

import (
	"api/internal/di"
	generated "api/internal/domains/notification/persistence/sqlc/generated"
	values "api/internal/domains/notification/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type PushTokenRepository struct {
	queries *generated.Queries
	db      *sql.DB
}

func NewPushTokenRepository(container *di.Container) *PushTokenRepository {
	return &PushTokenRepository{
		queries: generated.New(container.DB),
		db:      container.DB,
	}
}

func (r *PushTokenRepository) WithTx(tx *sql.Tx) *PushTokenRepository {
	return &PushTokenRepository{
		queries: r.queries.WithTx(tx),
		db:      r.db,
	}
}

func (r *PushTokenRepository) UpsertPushToken(ctx context.Context, userID uuid.UUID, token, deviceType string) *errLib.CommonError {
	_, err := r.queries.UpsertPushToken(ctx, generated.UpsertPushTokenParams{
		UserID:        userID,
		ExpoPushToken: token,
		DeviceType:    sql.NullString{String: deviceType, Valid: true},
	})

	if err != nil {
		log.Printf("Failed to upsert push token for user %s: %v", userID, err)
		return errLib.New("Failed to register push token", http.StatusInternalServerError)
	}

	return nil
}

func (r *PushTokenRepository) GetPushTokensByUserID(ctx context.Context, userID uuid.UUID) ([]values.PushToken, *errLib.CommonError) {
	dbTokens, err := r.queries.GetPushTokensByUserID(ctx, userID)
	if err != nil {
		log.Printf("Failed to get push tokens for user %s: %v", userID, err)
		return nil, errLib.New("Failed to get push tokens", http.StatusInternalServerError)
	}

	tokens := make([]values.PushToken, len(dbTokens))
	for i, dbToken := range dbTokens {
		tokens[i] = values.PushToken{
			ID:            int(dbToken.ID),
			UserID:        dbToken.UserID,
			ExpoPushToken: dbToken.ExpoPushToken,
			DeviceType:    dbToken.DeviceType.String,
		}
	}

	return tokens, nil
}

func (r *PushTokenRepository) GetPushTokensByTeamID(ctx context.Context, teamID uuid.UUID) ([]values.PushToken, *errLib.CommonError) {
	dbTokens, err := r.queries.GetPushTokensByTeamID(ctx, uuid.NullUUID{UUID: teamID, Valid: true})
	if err != nil {
		log.Printf("Failed to get push tokens for team %s: %v", teamID, err)
		return nil, errLib.New("Failed to get team push tokens", http.StatusInternalServerError)
	}

	tokens := make([]values.PushToken, len(dbTokens))
	for i, dbToken := range dbTokens {
		tokens[i] = values.PushToken{
			ID:            int(dbToken.ID),
			UserID:        dbToken.UserID,
			ExpoPushToken: dbToken.ExpoPushToken,
			DeviceType:    dbToken.DeviceType.String,
		}
	}

	return tokens, nil
}
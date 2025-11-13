package repository

import (
	"database/sql"

	"api/internal/di"
	db "api/internal/domains/subsidy/persistence/sqlc/generated"
)

type SubsidyRepository struct {
	Queries *db.Queries
	Tx      *sql.Tx
	DB      *sql.DB
}

func NewSubsidyRepository(container *di.Container) *SubsidyRepository {
	return &SubsidyRepository{
		Queries: container.Queries.SubsidyDb,
		DB:      container.DB,
	}
}

func (r *SubsidyRepository) GetTx() *sql.Tx {
	return r.Tx
}

func (r *SubsidyRepository) WithTx(tx *sql.Tx) *SubsidyRepository {
	return &SubsidyRepository{
		Queries: r.Queries.WithTx(tx),
		Tx:      tx,
		DB:      r.DB,
	}
}

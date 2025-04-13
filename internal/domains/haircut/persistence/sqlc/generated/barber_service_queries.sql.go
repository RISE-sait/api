// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: barber_service_queries.sql

package db_haircut

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createBarberService = `-- name: CreateBarberService :execrows
INSERT INTO haircut.barber_services (barber_id, service_id)
VALUES ($1, $2)
`

type CreateBarberServiceParams struct {
	BarberID  uuid.UUID `json:"barber_id"`
	ServiceID uuid.UUID `json:"service_id"`
}

func (q *Queries) CreateBarberService(ctx context.Context, arg CreateBarberServiceParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, createBarberService, arg.BarberID, arg.ServiceID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteBarberService = `-- name: DeleteBarberService :execrows
DELETE
FROM haircut.barber_services
WHERE id = $1
`

func (q *Queries) DeleteBarberService(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteBarberService, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getBarberServices = `-- name: GetBarberServices :many
SELECT bs.id, bs.barber_id, bs.service_id, bs.created_at, bs.updated_at, (u.first_name || ' ' || u.last_name)::text as barber_name, hs.name as haircut_name
FROM haircut.barber_services bs
         JOIN users.users u ON u.id = bs.barber_id
         JOIN haircut.haircut_services hs ON hs.id = bs.service_id
`

type GetBarberServicesRow struct {
	ID          uuid.UUID `json:"id"`
	BarberID    uuid.UUID `json:"barber_id"`
	ServiceID   uuid.UUID `json:"service_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	BarberName  string    `json:"barber_name"`
	HaircutName string    `json:"haircut_name"`
}

func (q *Queries) GetBarberServices(ctx context.Context) ([]GetBarberServicesRow, error) {
	rows, err := q.db.QueryContext(ctx, getBarberServices)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetBarberServicesRow
	for rows.Next() {
		var i GetBarberServicesRow
		if err := rows.Scan(
			&i.ID,
			&i.BarberID,
			&i.ServiceID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.BarberName,
			&i.HaircutName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

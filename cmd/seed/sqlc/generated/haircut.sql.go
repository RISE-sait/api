// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: haircut.sql

package db_seed

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

const insertBarberServices = `-- name: InsertBarberServices :exec
WITH prepared_data AS (SELECT unnest($1::text[]) AS barber_email,
                              unnest($2::text[]) AS service_name)
INSERT
INTO haircut.barber_services(barber_id, service_id)
SELECT u.id, h.id
FROM prepared_data p
         JOIN users.users u ON p.barber_email = u.email
         JOIN haircut.haircut_services h ON p.service_name = h.name
`

type InsertBarberServicesParams struct {
	BarberEmailArray []string `json:"barber_email_array"`
	ServiceNameArray []string `json:"service_name_array"`
}

func (q *Queries) InsertBarberServices(ctx context.Context, arg InsertBarberServicesParams) error {
	_, err := q.db.ExecContext(ctx, insertBarberServices, pq.Array(arg.BarberEmailArray), pq.Array(arg.ServiceNameArray))
	return err
}

const insertHaircutEvents = `-- name: InsertHaircutEvents :exec
WITH prepared_data AS (SELECT unnest($1::timestamptz[]) AS begin_date_time,
                              unnest($2::timestamptz[])   AS end_date_time,
                              unnest($3::uuid[])            AS customer_id,
                              unnest($4::text[])           AS barber_email,
                              unnest($5::text[])           AS haircut_name),
     haircut_data AS (SELECT pd.begin_date_time,
                             pd.end_date_time,
                             pd.customer_id,
                             ub.id AS barber_id,
                             h.id  as haircut_id
                      FROM prepared_data pd
                               LEFT JOIN
                           users.users ub ON ub.email = pd.barber_email
                               JOIN haircut.haircut_services h ON h.name = pd.haircut_name)
INSERT
INTO haircut.events (begin_date_time, end_date_time, customer_id, barber_id, service_type_id)
SELECT begin_date_time,
       end_date_time,
       customer_id,
       barber_id,
       haircut_id
FROM haircut_data
WHERE customer_id IS NOT NULL
  AND barber_id IS NOT NULL
ON CONFLICT DO NOTHING
`

type InsertHaircutEventsParams struct {
	BeginDateTimeArray []time.Time `json:"begin_date_time_array"`
	EndDateTimeArray   []time.Time `json:"end_date_time_array"`
	CustomerIDArray    []uuid.UUID `json:"customer_id_array"`
	BarberEmailArray   []string    `json:"barber_email_array"`
	HaircutNameArray   []string    `json:"haircut_name_array"`
}

func (q *Queries) InsertHaircutEvents(ctx context.Context, arg InsertHaircutEventsParams) error {
	_, err := q.db.ExecContext(ctx, insertHaircutEvents,
		pq.Array(arg.BeginDateTimeArray),
		pq.Array(arg.EndDateTimeArray),
		pq.Array(arg.CustomerIDArray),
		pq.Array(arg.BarberEmailArray),
		pq.Array(arg.HaircutNameArray),
	)
	return err
}

const insertHaircutServices = `-- name: InsertHaircutServices :exec
WITH prepared_data AS (SELECT unnest($1::text[])           AS name,
                              unnest($2::text[])    AS description,
                              unnest($3::numeric[])       AS price,
                              unnest($4::int[]) AS duration_in_min)
INSERT
INTO haircut.haircut_services (name, description, price, duration_in_min)
SELECT name,
       NULLIF(description, ''),
       price,
       duration_in_min
FROM prepared_data
`

type InsertHaircutServicesParams struct {
	NameArray          []string          `json:"name_array"`
	DescriptionArray   []string          `json:"description_array"`
	PriceArray         []decimal.Decimal `json:"price_array"`
	DurationInMinArray []int32           `json:"duration_in_min_array"`
}

func (q *Queries) InsertHaircutServices(ctx context.Context, arg InsertHaircutServicesParams) error {
	_, err := q.db.ExecContext(ctx, insertHaircutServices,
		pq.Array(arg.NameArray),
		pq.Array(arg.DescriptionArray),
		pq.Array(arg.PriceArray),
		pq.Array(arg.DurationInMinArray),
	)
	return err
}

// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: event_queries.sql

package db_haircut

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createHaircutEvent = `-- name: CreateHaircutEvent :one
INSERT INTO haircut.events (begin_date_time, end_date_time, barber_id, customer_id, service_type_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, begin_date_time, end_date_time, customer_id, barber_id, service_type_id, created_at, updated_at,
    (SELECT first_name || ' ' || last_name FROM users.users WHERE id = customer_id)::varchar as customer_name,
    (SELECT first_name || ' ' || last_name FROM users.users WHERE id = barber_id)::varchar as barber_name
`

type CreateHaircutEventParams struct {
	BeginDateTime time.Time `json:"begin_date_time"`
	EndDateTime   time.Time `json:"end_date_time"`
	BarberID      uuid.UUID `json:"barber_id"`
	CustomerID    uuid.UUID `json:"customer_id"`
	ServiceTypeID uuid.UUID `json:"service_type_id"`
}

type CreateHaircutEventRow struct {
	ID            uuid.UUID `json:"id"`
	BeginDateTime time.Time `json:"begin_date_time"`
	EndDateTime   time.Time `json:"end_date_time"`
	CustomerID    uuid.UUID `json:"customer_id"`
	BarberID      uuid.UUID `json:"barber_id"`
	ServiceTypeID uuid.UUID `json:"service_type_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CustomerName  string    `json:"customer_name"`
	BarberName    string    `json:"barber_name"`
}

func (q *Queries) CreateHaircutEvent(ctx context.Context, arg CreateHaircutEventParams) (CreateHaircutEventRow, error) {
	row := q.db.QueryRowContext(ctx, createHaircutEvent,
		arg.BeginDateTime,
		arg.EndDateTime,
		arg.BarberID,
		arg.CustomerID,
		arg.ServiceTypeID,
	)
	var i CreateHaircutEventRow
	err := row.Scan(
		&i.ID,
		&i.BeginDateTime,
		&i.EndDateTime,
		&i.CustomerID,
		&i.BarberID,
		&i.ServiceTypeID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CustomerName,
		&i.BarberName,
	)
	return i, err
}

const deleteEvent = `-- name: DeleteEvents :execrows
DELETE
FROM haircut.events
WHERE id = $1
`

func (q *Queries) DeleteEvent(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteEvent, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getEventById = `-- name: GetEventById :one
SELECT e.id, begin_date_time, end_date_time, customer_id, barber_id, service_type_id, e.created_at, e.updated_at, barbers.id, barbers.hubspot_id, barbers.country_alpha2_code, barbers.gender, barbers.first_name, barbers.last_name, barbers.parent_id, barbers.phone, barbers.email, barbers.has_marketing_email_consent, barbers.has_sms_consent, barbers.created_at, barbers.updated_at, barbers.dob, customers.id, customers.hubspot_id, customers.country_alpha2_code, customers.gender, customers.first_name, customers.last_name, customers.parent_id, customers.phone, customers.email, customers.has_marketing_email_consent, customers.has_sms_consent, customers.created_at, customers.updated_at, customers.dob,
       (barbers.first_name || ' ' || barbers.last_name)::text     as barber_name,
       (customers.first_name || ' ' || customers.last_name)::text as customer_name
FROM haircut.events e
         JOIN users.users barbers
              ON barbers.id = barber_id

         JOIN users.users customers
              ON customers.id = customer_id
WHERE e.id = $1
`

type GetEventByIdRow struct {
	ID                         uuid.UUID      `json:"id"`
	BeginDateTime              time.Time      `json:"begin_date_time"`
	EndDateTime                time.Time      `json:"end_date_time"`
	CustomerID                 uuid.UUID      `json:"customer_id"`
	BarberID                   uuid.UUID      `json:"barber_id"`
	ServiceTypeID              uuid.UUID      `json:"service_type_id"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
	ID_2                       uuid.UUID      `json:"id_2"`
	HubspotID                  sql.NullString `json:"hubspot_id"`
	CountryAlpha2Code          string         `json:"country_alpha2_code"`
	Gender                     sql.NullString `json:"gender"`
	FirstName                  string         `json:"first_name"`
	LastName                   string         `json:"last_name"`
	ParentID                   uuid.NullUUID  `json:"parent_id"`
	Phone                      sql.NullString `json:"phone"`
	Email                      sql.NullString `json:"email"`
	HasMarketingEmailConsent   bool           `json:"has_marketing_email_consent"`
	HasSmsConsent              bool           `json:"has_sms_consent"`
	CreatedAt_2                time.Time      `json:"created_at_2"`
	UpdatedAt_2                time.Time      `json:"updated_at_2"`
	Dob                        time.Time      `json:"dob"`
	ID_3                       uuid.UUID      `json:"id_3"`
	HubspotID_2                sql.NullString `json:"hubspot_id_2"`
	CountryAlpha2Code_2        string         `json:"country_alpha2_code_2"`
	Gender_2                   sql.NullString `json:"gender_2"`
	FirstName_2                string         `json:"first_name_2"`
	LastName_2                 string         `json:"last_name_2"`
	ParentID_2                 uuid.NullUUID  `json:"parent_id_2"`
	Phone_2                    sql.NullString `json:"phone_2"`
	Email_2                    sql.NullString `json:"email_2"`
	HasMarketingEmailConsent_2 bool           `json:"has_marketing_email_consent_2"`
	HasSmsConsent_2            bool           `json:"has_sms_consent_2"`
	CreatedAt_3                time.Time      `json:"created_at_3"`
	UpdatedAt_3                time.Time      `json:"updated_at_3"`
	Dob_2                      time.Time      `json:"dob_2"`
	BarberName                 string         `json:"barber_name"`
	CustomerName               string         `json:"customer_name"`
}

func (q *Queries) GetEventById(ctx context.Context, id uuid.UUID) (GetEventByIdRow, error) {
	row := q.db.QueryRowContext(ctx, getEventById, id)
	var i GetEventByIdRow
	err := row.Scan(
		&i.ID,
		&i.BeginDateTime,
		&i.EndDateTime,
		&i.CustomerID,
		&i.BarberID,
		&i.ServiceTypeID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ID_2,
		&i.HubspotID,
		&i.CountryAlpha2Code,
		&i.Gender,
		&i.FirstName,
		&i.LastName,
		&i.ParentID,
		&i.Phone,
		&i.Email,
		&i.HasMarketingEmailConsent,
		&i.HasSmsConsent,
		&i.CreatedAt_2,
		&i.UpdatedAt_2,
		&i.Dob,
		&i.ID_3,
		&i.HubspotID_2,
		&i.CountryAlpha2Code_2,
		&i.Gender_2,
		&i.FirstName_2,
		&i.LastName_2,
		&i.ParentID_2,
		&i.Phone_2,
		&i.Email_2,
		&i.HasMarketingEmailConsent_2,
		&i.HasSmsConsent_2,
		&i.CreatedAt_3,
		&i.UpdatedAt_3,
		&i.Dob_2,
		&i.BarberName,
		&i.CustomerName,
	)
	return i, err
}

const getHaircutEvents = `-- name: GetHaircutEvents :many
SELECT e.id, e.begin_date_time, e.end_date_time, e.customer_id, e.barber_id, e.service_type_id, e.created_at, e.updated_at,
       (barbers.first_name || ' ' || barbers.last_name)::text     as barber_name,
       (customers.first_name || ' ' || customers.last_name)::text as customer_name
FROM haircut.events e
         JOIN users.users barbers
              ON barbers.id = e.barber_id
         JOIN users.users customers
              ON customers.id = e.customer_id
WHERE
    (barber_id = $1 OR $1 IS NULL) -- Filter by barber_id
  AND (customer_id = $2 OR $2 IS NULL)
  AND ($3 >= begin_date_time OR $3 IS NULL) -- within boundary
  AND ($4 <= end_date_time OR $4 IS NULL)
`

type GetHaircutEventsParams struct {
	BarberID   uuid.NullUUID `json:"barber_id"`
	CustomerID uuid.NullUUID `json:"customer_id"`
	Before     sql.NullTime  `json:"before"`
	After      sql.NullTime  `json:"after"`
}

type GetHaircutEventsRow struct {
	ID            uuid.UUID `json:"id"`
	BeginDateTime time.Time `json:"begin_date_time"`
	EndDateTime   time.Time `json:"end_date_time"`
	CustomerID    uuid.UUID `json:"customer_id"`
	BarberID      uuid.UUID `json:"barber_id"`
	ServiceTypeID uuid.UUID `json:"service_type_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	BarberName    string    `json:"barber_name"`
	CustomerName  string    `json:"customer_name"`
}

func (q *Queries) GetHaircutEvents(ctx context.Context, arg GetHaircutEventsParams) ([]GetHaircutEventsRow, error) {
	rows, err := q.db.QueryContext(ctx, getHaircutEvents,
		arg.BarberID,
		arg.CustomerID,
		arg.Before,
		arg.After,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetHaircutEventsRow
	for rows.Next() {
		var i GetHaircutEventsRow
		if err := rows.Scan(
			&i.ID,
			&i.BeginDateTime,
			&i.EndDateTime,
			&i.CustomerID,
			&i.BarberID,
			&i.ServiceTypeID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.BarberName,
			&i.CustomerName,
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

const updateEvent = `-- name: UpdateEvent :one
UPDATE haircut.events
SET
    begin_date_time = $1,
    end_date_time = $2,
    barber_id = $3,
    customer_id = $4,
    updated_at  = current_timestamp
WHERE id = $5
RETURNING id, begin_date_time, end_date_time, customer_id, barber_id, service_type_id, created_at, updated_at
`

type UpdateEventParams struct {
	BeginDateTime time.Time `json:"begin_date_time"`
	EndDateTime   time.Time `json:"end_date_time"`
	BarberID      uuid.UUID `json:"barber_id"`
	CustomerID    uuid.UUID `json:"customer_id"`
	ID            uuid.UUID `json:"id"`
}

func (q *Queries) UpdateEvent(ctx context.Context, arg UpdateEventParams) (HaircutEvent, error) {
	row := q.db.QueryRowContext(ctx, updateEvent,
		arg.BeginDateTime,
		arg.EndDateTime,
		arg.BarberID,
		arg.CustomerID,
		arg.ID,
	)
	var i HaircutEvent
	err := row.Scan(
		&i.ID,
		&i.BeginDateTime,
		&i.EndDateTime,
		&i.CustomerID,
		&i.BarberID,
		&i.ServiceTypeID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

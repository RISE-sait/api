// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: programs.sql

package db_payment

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const getProgram = `-- name: GetProgram :one
SELECT id, name
FROM program.programs
WHERE id = $1
`

type GetProgramRow struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func (q *Queries) GetProgram(ctx context.Context, id uuid.UUID) (GetProgramRow, error) {
	row := q.db.QueryRowContext(ctx, getProgram, id)
	var i GetProgramRow
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getProgramCapacityStatus = `-- name: GetProgramCapacityStatus :one
SELECT
        capacity,
        (SELECT COUNT(*) FROM program.customer_enrollment ce WHERE ce.program_id = $1) AS enrolled_count
    FROM program.programs
    WHERE id = $1
`

type GetProgramCapacityStatusRow struct {
	Capacity      sql.NullInt32 `json:"capacity"`
	EnrolledCount int64         `json:"enrolled_count"`
}

func (q *Queries) GetProgramCapacityStatus(ctx context.Context, programID uuid.UUID) (GetProgramCapacityStatusRow, error) {
	row := q.db.QueryRowContext(ctx, getProgramCapacityStatus, programID)
	var i GetProgramCapacityStatusRow
	err := row.Scan(&i.Capacity, &i.EnrolledCount)
	return i, err
}

const getProgramIdByStripePriceId = `-- name: GetProgramIdByStripePriceId :one
SELECT pm.program_id
FROM program.program_membership pm
WHERE pm.stripe_program_price_id = $1
`

func (q *Queries) GetProgramIdByStripePriceId(ctx context.Context, stripeProgramPriceID string) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getProgramIdByStripePriceId, stripeProgramPriceID)
	var program_id uuid.UUID
	err := row.Scan(&program_id)
	return program_id, err
}

const getProgramRegistrationPriceIdForCustomer = `-- name: GetProgramRegistrationPriceIdForCustomer :one
SELECT pm.stripe_program_price_id
FROM program.program_membership pm
WHERE pm.membership_id = (SELECT mp.membership_id
                          FROM users.customer_membership_plans cmp
                                   LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
                                   LEFT JOIN membership.memberships m ON m.id = mp.membership_id
                          WHERE customer_id = $1
                            AND status = 'active'
                          ORDER BY cmp.start_date DESC
                          LIMIT 1)
  AND pm.program_id = $2
`

type GetProgramRegistrationPriceIdForCustomerParams struct {
	CustomerID uuid.UUID `json:"customer_id"`
	ProgramID  uuid.UUID `json:"program_id"`
}

func (q *Queries) GetProgramRegistrationPriceIdForCustomer(ctx context.Context, arg GetProgramRegistrationPriceIdForCustomerParams) (string, error) {
	row := q.db.QueryRowContext(ctx, getProgramRegistrationPriceIdForCustomer, arg.CustomerID, arg.ProgramID)
	var stripe_program_price_id string
	err := row.Scan(&stripe_program_price_id)
	return stripe_program_price_id, err
}

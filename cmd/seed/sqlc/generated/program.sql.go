// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: program.sql

package db_seed

import (
	"context"

	"github.com/lib/pq"
)

const getProgramByType = `-- name: GetProgramByType :one
SELECT id, name, description, level, type, capacity, created_at, updated_at, pay_per_event
FROM program.programs
WHERE type = $1
LIMIT 1
`

func (q *Queries) GetProgramByType(ctx context.Context, type_ ProgramProgramType) (ProgramProgram, error) {
	row := q.db.QueryRowContext(ctx, getProgramByType, type_)
	var i ProgramProgram
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Level,
		&i.Type,
		&i.Capacity,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.PayPerEvent,
	)
	return i, err
}

const insertBuiltInPrograms = `-- name: InsertBuiltInPrograms :exec
INSERT INTO program.programs (id, name, type, description)
VALUES
    (gen_random_uuid(), 'Course', 'course'::program.program_type, 'Default program for courses'),
    (gen_random_uuid(), 'Other', 'other'::program.program_type, 'Default program for other events'),
    (gen_random_uuid(), 'Tournament', 'tournament'::program.program_type, 'Default program for tournaments'),
    (gen_random_uuid(), 'Tryouts', 'tryouts'::program.program_type, 'Default program for tryouts'),
    (gen_random_uuid(), 'Event', 'event'::program.program_type, 'Default program for events')
ON CONFLICT (name) DO UPDATE
SET type = EXCLUDED.type,
    description = EXCLUDED.description
`

func (q *Queries) InsertBuiltInPrograms(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, insertBuiltInPrograms)
	return err
}

const insertProgramFees = `-- name: InsertProgramFees :exec
WITH prepared_data AS (
    SELECT
        unnest($1::varchar[])            AS program_name,
        unnest($2::varchar[])         AS membership_name,
        unnest($3::varchar[]) AS stripe_program_price_id
)
INSERT INTO program.fees (program_id, membership_id, stripe_price_id)
SELECT
    p.id,
    CASE WHEN m.id IS NULL THEN NULL::uuid ELSE m.id END,
    stripe_program_price_id
FROM prepared_data
JOIN program.programs p ON p.name = program_name
LEFT JOIN membership.memberships m ON m.name = membership_name
WHERE stripe_program_price_id <> ''
`

type InsertProgramFeesParams struct {
	ProgramNameArray          []string `json:"program_name_array"`
	MembershipNameArray       []string `json:"membership_name_array"`
	StripeProgramPriceIDArray []string `json:"stripe_program_price_id_array"`
}

func (q *Queries) InsertProgramFees(ctx context.Context, arg InsertProgramFeesParams) error {
	_, err := q.db.ExecContext(ctx, insertProgramFees, pq.Array(arg.ProgramNameArray), pq.Array(arg.MembershipNameArray), pq.Array(arg.StripeProgramPriceIDArray))
	return err
}

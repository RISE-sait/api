// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: program.sql

package db_seed

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const insertCourses = `-- name: InsertCourses :exec
WITH prepared_data as (SELECT unnest($1::text[])                   as name,
                              unnest($2::text[])            as description,
                              unnest($3::program.program_level[]) as level,
                              unnest($4::boolean[])        AS pay_per_event)
INSERT
INTO program.programs (name, description, type, level, pay_per_event)
SELECT name,
       description,
       'course',
       level,
         pay_per_event
FROM prepared_data
`

type InsertCoursesParams struct {
	NameArray          []string              `json:"name_array"`
	DescriptionArray   []string              `json:"description_array"`
	LevelArray         []ProgramProgramLevel `json:"level_array"`
	IsPayPerEventArray []bool                `json:"is_pay_per_event_array"`
}

func (q *Queries) InsertCourses(ctx context.Context, arg InsertCoursesParams) error {
	_, err := q.db.ExecContext(ctx, insertCourses,
		pq.Array(arg.NameArray),
		pq.Array(arg.DescriptionArray),
		pq.Array(arg.LevelArray),
		pq.Array(arg.IsPayPerEventArray),
	)
	return err
}

const insertGames = `-- name: InsertGames :exec
WITH prepared_data as (SELECT unnest($5::text[])                   as name,
                              unnest($6::text[])            as description,
                              unnest($7::program.program_level[]) as level,
                              unnest($8::boolean[])        AS pay_per_event
                              ),
     game_ids AS (
         INSERT INTO program.programs (name, description, type, level, pay_per_event)
             SELECT name, description, 'game', level, pay_per_event
             FROM prepared_data
             RETURNING id)
INSERT
INTO program.games (id, win_team, lose_team, win_score, lose_score)
VALUES (unnest(ARRAY(SELECT id FROM game_ids)), unnest($1::uuid[]), unnest($2::uuid[]),
        unnest($3::int[]), unnest($4::int[]))
`

type InsertGamesParams struct {
	WinTeamArray       []uuid.UUID           `json:"win_team_array"`
	LoseTeamArray      []uuid.UUID           `json:"lose_team_array"`
	WinScoreArray      []int32               `json:"win_score_array"`
	LoseScoreArray     []int32               `json:"lose_score_array"`
	NameArray          []string              `json:"name_array"`
	DescriptionArray   []string              `json:"description_array"`
	LevelArray         []ProgramProgramLevel `json:"level_array"`
	IsPayPerEventArray []bool                `json:"is_pay_per_event_array"`
}

func (q *Queries) InsertGames(ctx context.Context, arg InsertGamesParams) error {
	_, err := q.db.ExecContext(ctx, insertGames,
		pq.Array(arg.WinTeamArray),
		pq.Array(arg.LoseTeamArray),
		pq.Array(arg.WinScoreArray),
		pq.Array(arg.LoseScoreArray),
		pq.Array(arg.NameArray),
		pq.Array(arg.DescriptionArray),
		pq.Array(arg.LevelArray),
		pq.Array(arg.IsPayPerEventArray),
	)
	return err
}

const insertPractices = `-- name: InsertPractices :many
WITH prepared_data as (SELECT unnest($1::text[])                   as name,
                              unnest($2::text[])            as description,
                              unnest($3::program.program_level[]) as level,
                       unnest($4::boolean[])        AS pay_per_event)
INSERT
INTO program.programs (name, description, type, level, pay_per_event)
SELECT name,
       description,
       'practice',
       level,
       pay_per_event
FROM prepared_data
RETURNING id
`

type InsertPracticesParams struct {
	NameArray          []string              `json:"name_array"`
	DescriptionArray   []string              `json:"description_array"`
	LevelArray         []ProgramProgramLevel `json:"level_array"`
	IsPayPerEventArray []bool                `json:"is_pay_per_event_array"`
}

func (q *Queries) InsertPractices(ctx context.Context, arg InsertPracticesParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertPractices,
		pq.Array(arg.NameArray),
		pq.Array(arg.DescriptionArray),
		pq.Array(arg.LevelArray),
		pq.Array(arg.IsPayPerEventArray),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertProgramFees = `-- name: InsertProgramFees :exec
WITH prepared_data AS (SELECT unnest($1::varchar[])            AS program_name,
                              unnest($2::varchar[])         AS membership_name,
                              unnest($3::varchar[]) AS stripe_program_price_id)
INSERT
INTO program.fees (program_id, membership_id, stripe_price_id)
SELECT p.id,
       CASE
           WHEN m.id IS NULL THEN NULL::uuid
           ELSE m.id
           END,
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

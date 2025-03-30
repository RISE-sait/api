// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: general.sql

package db_seed

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

const insertCoachStats = `-- name: InsertCoachStats :exec
INSERT INTO athletic.coach_stats (coach_id, wins, losses)
VALUES ((SELECT id FROM users.users WHERE email = 'viktor.djurasic+1@abcfitness.com'),
        1,
        1),
       ((SELECT id FROM users.users WHERE email = 'coach@test.com'),
        1,
        2)
`

func (q *Queries) InsertCoachStats(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, insertCoachStats)
	return err
}

const insertCourses = `-- name: InsertCourses :exec
WITH prepared_data as (SELECT unnest($1::text[]) as name,
        unnest($2::text[]) as description,
        unnest($3::program.program_level[]) as level)
INSERT INTO program.programs (name, description, type, level)
SELECT name,
       description,
       'course',
       level
FROM prepared_data
`

type InsertCoursesParams struct {
	NameArray        []string              `json:"name_array"`
	DescriptionArray []string              `json:"description_array"`
	LevelArray       []ProgramProgramLevel `json:"level_array"`
}

func (q *Queries) InsertCourses(ctx context.Context, arg InsertCoursesParams) error {
	_, err := q.db.ExecContext(ctx, insertCourses, pq.Array(arg.NameArray), pq.Array(arg.DescriptionArray), pq.Array(arg.LevelArray))
	return err
}

const insertEnrollmentFees = `-- name: InsertEnrollmentFees :exec
WITH prepared_data AS (SELECT unnest($1::uuid[])       AS program_id,
                              unnest($2::uuid[])    AS membership_id,
                              unnest($3::numeric[]) AS drop_in_price,
                              unnest($4::numeric[]) AS program_price)
INSERT
INTO enrollment_fees (program_id, membership_id, drop_in_price, program_price)
SELECT program_id,
       CASE
           WHEN membership_id = '00000000-0000-0000-0000-000000000000' THEN NULL::uuid
           ELSE membership_id
           END AS membership_id,
       CASE
           WHEN drop_in_price = 9999 THEN NULL::numeric
           ELSE drop_in_price
           END AS payg_price,
       CASE
           WHEN program_price = 9999 THEN NULL::numeric
           ELSE program_price
           END AS program_price
FROM prepared_data
`

type InsertEnrollmentFeesParams struct {
	ProgramIDArray    []uuid.UUID       `json:"program_id_array"`
	MembershipIDArray []uuid.UUID       `json:"membership_id_array"`
	DropInPriceArray  []decimal.Decimal `json:"drop_in_price_array"`
	ProgramPriceArray []decimal.Decimal `json:"program_price_array"`
}

func (q *Queries) InsertEnrollmentFees(ctx context.Context, arg InsertEnrollmentFeesParams) error {
	_, err := q.db.ExecContext(ctx, insertEnrollmentFees,
		pq.Array(arg.ProgramIDArray),
		pq.Array(arg.MembershipIDArray),
		pq.Array(arg.DropInPriceArray),
		pq.Array(arg.ProgramPriceArray),
	)
	return err
}

const insertGames = `-- name: InsertGames :exec
WITH prepared_data as (
        SELECT unnest($5::text[]) as name,
        unnest($6::text[]) as description,
        unnest($7::program.program_level[]) as level),
game_ids AS (
    INSERT INTO program.programs (name, description, type, level)
    SELECT name, description, 'game', level
    FROM prepared_data
    RETURNING id
)
INSERT INTO public.games (id, win_team, lose_team, win_score, lose_score)
VALUES (unnest(ARRAY(SELECT id FROM game_ids)), unnest($1::uuid[]), unnest($2::uuid[]), unnest($3::int[]), unnest($4::int[]))
`

type InsertGamesParams struct {
	WinTeamArray     []uuid.UUID           `json:"win_team_array"`
	LoseTeamArray    []uuid.UUID           `json:"lose_team_array"`
	WinScoreArray    []int32               `json:"win_score_array"`
	LoseScoreArray   []int32               `json:"lose_score_array"`
	NameArray        []string              `json:"name_array"`
	DescriptionArray []string              `json:"description_array"`
	LevelArray       []ProgramProgramLevel `json:"level_array"`
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
	)
	return err
}

const insertLocations = `-- name: InsertLocations :exec
INSERT INTO location.locations (name, address)
VALUES (unnest($1::text[]), unnest($2::text[]))
RETURNING id
`

type InsertLocationsParams struct {
	NameArray    []string `json:"name_array"`
	AddressArray []string `json:"address_array"`
}

func (q *Queries) InsertLocations(ctx context.Context, arg InsertLocationsParams) error {
	_, err := q.db.ExecContext(ctx, insertLocations, pq.Array(arg.NameArray), pq.Array(arg.AddressArray))
	return err
}

const insertPractices = `-- name: InsertPractices :many
WITH prepared_data as (
        SELECT unnest($1::text[]) as name,
        unnest($2::text[]) as description,
        unnest($3::program.program_level[]) as level)
INSERT INTO program.programs (name, description, type, level)
SELECT name,
       description,
       'practice',
       level
FROM prepared_data
RETURNING id
`

type InsertPracticesParams struct {
	NameArray        []string              `json:"name_array"`
	DescriptionArray []string              `json:"description_array"`
	LevelArray       []ProgramProgramLevel `json:"level_array"`
}

func (q *Queries) InsertPractices(ctx context.Context, arg InsertPracticesParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertPractices, pq.Array(arg.NameArray), pq.Array(arg.DescriptionArray), pq.Array(arg.LevelArray))
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

const insertTeams = `-- name: InsertTeams :many
WITH prepared_data AS (SELECT unnest($1::text[])          AS coach,
                              unnest($2::int[])             AS capacity,
                              unnest($3::text[]) AS name)
INSERT
INTO athletic.teams(capacity, coach_id, name)
SELECT capacity, u.id, name
FROM prepared_data
JOIN users.users u ON u.email = coach
RETURNING id
`

type InsertTeamsParams struct {
	CoachEmailArray []string `json:"coach_email_array"`
	CapacityArray   []int32  `json:"capacity_array"`
	NameArray       []string `json:"name_array"`
}

func (q *Queries) InsertTeams(ctx context.Context, arg InsertTeamsParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertTeams, pq.Array(arg.CoachEmailArray), pq.Array(arg.CapacityArray), pq.Array(arg.NameArray))
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

const insertWaivers = `-- name: InsertWaivers :exec
INSERT INTO waiver.waiver(waiver_url, waiver_name)
VALUES ('https://storage.googleapis.com/rise-sports/waivers/code.pdf', 'code_pdf'),
       ('https://storage.googleapis.com/rise-sports/waivers/tetris.pdf', 'tetris_pdf')
`

func (q *Queries) InsertWaivers(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, insertWaivers)
	return err
}

// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: game_queries.sql

package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createGame = `-- name: CreateGame :one
INSERT INTO games (name)
VALUES ($1)
RETURNING id, name
`

func (q *Queries) CreateGame(ctx context.Context, name string) (Game, error) {
	row := q.db.QueryRowContext(ctx, createGame, name)
	var i Game
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const deleteGame = `-- name: DeleteGame :execrows
DELETE FROM games WHERE id = $1
`

func (q *Queries) DeleteGame(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteGame, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getGameById = `-- name: GetGameById :one
SELECT id, name FROM games WHERE id = $1
`

func (q *Queries) GetGameById(ctx context.Context, id uuid.UUID) (Game, error) {
	row := q.db.QueryRowContext(ctx, getGameById, id)
	var i Game
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const getGames = `-- name: GetGames :many
SELECT id, name FROM games
WHERE (name ILIKE '%' || $1 || '%' OR $1 IS NULL)
`

func (q *Queries) GetGames(ctx context.Context, name sql.NullString) ([]Game, error) {
	rows, err := q.db.QueryContext(ctx, getGames, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Game
	for rows.Next() {
		var i Game
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
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

const updateGame = `-- name: UpdateGame :one
UPDATE games
SET name = COALESCE($2, name)
WHERE id = $1
RETURNING id, name
`

type UpdateGameParams struct {
	ID   uuid.UUID      `json:"id"`
	Name sql.NullString `json:"name"`
}

func (q *Queries) UpdateGame(ctx context.Context, arg UpdateGameParams) (Game, error) {
	row := q.db.QueryRowContext(ctx, updateGame, arg.ID, arg.Name)
	var i Game
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package db

import (
	"context"
)

const createGame = `-- name: CreateGame :one
INSERT INTO games (
  game_id, game_info, created_dt_tm, created_user_id
) VALUES (
  ?, ?, ?, ?
)
RETURNING id, game_id, game_info, created_dt_tm, created_user_id
`

type CreateGameParams struct {
	GameID        string
	GameInfo      string
	CreatedDtTm   string
	CreatedUserID string
}

func (q *Queries) CreateGame(ctx context.Context, arg CreateGameParams) (Game, error) {
	row := q.db.QueryRowContext(ctx, createGame,
		arg.GameID,
		arg.GameInfo,
		arg.CreatedDtTm,
		arg.CreatedUserID,
	)
	var i Game
	err := row.Scan(
		&i.ID,
		&i.GameID,
		&i.GameInfo,
		&i.CreatedDtTm,
		&i.CreatedUserID,
	)
	return i, err
}

const createGamestate = `-- name: CreateGamestate :one
INSERT INTO gamestate (
  user_id, game_id, game_state, created_dt_tm
) VALUES (
  ?, ?, ?, ?
)
RETURNING id, user_id, game_id, game_state, created_dt_tm
`

type CreateGamestateParams struct {
	UserID      string
	GameID      int64
	GameState   string
	CreatedDtTm string
}

func (q *Queries) CreateGamestate(ctx context.Context, arg CreateGamestateParams) (Gamestate, error) {
	row := q.db.QueryRowContext(ctx, createGamestate,
		arg.UserID,
		arg.GameID,
		arg.GameState,
		arg.CreatedDtTm,
	)
	var i Gamestate
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.GameID,
		&i.GameState,
		&i.CreatedDtTm,
	)
	return i, err
}

const createRateLimit = `-- name: CreateRateLimit :exec
INSERT INTO ratelimit (user_id, calls_remaining, reset_dt_tm)
VALUES (?, ?, ?)
`

type CreateRateLimitParams struct {
	UserID         string
	CallsRemaining int64
	ResetDtTm      string
}

func (q *Queries) CreateRateLimit(ctx context.Context, arg CreateRateLimitParams) error {
	_, err := q.db.ExecContext(ctx, createRateLimit, arg.UserID, arg.CallsRemaining, arg.ResetDtTm)
	return err
}

const deleteGame = `-- name: DeleteGame :exec
DELETE FROM games
WHERE game_id = ?
`

func (q *Queries) DeleteGame(ctx context.Context, gameID string) error {
	_, err := q.db.ExecContext(ctx, deleteGame, gameID)
	return err
}

const deleteGamestate = `-- name: DeleteGamestate :exec
DELETE FROM gamestate
WHERE game_id = ?  and user_id = ?
`

type DeleteGamestateParams struct {
	GameID int64
	UserID string
}

func (q *Queries) DeleteGamestate(ctx context.Context, arg DeleteGamestateParams) error {
	_, err := q.db.ExecContext(ctx, deleteGamestate, arg.GameID, arg.UserID)
	return err
}

const gameExists = `-- name: GameExists :one
SELECT COUNT(*) FROM games
WHERE game_id=? LIMIT 1
`

func (q *Queries) GameExists(ctx context.Context, gameID string) (int64, error) {
	row := q.db.QueryRowContext(ctx, gameExists, gameID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getGame = `-- name: GetGame :one
SELECT id, game_id, game_info, created_dt_tm, created_user_id FROM games
WHERE game_id=? LIMIT 1
`

func (q *Queries) GetGame(ctx context.Context, gameID string) (Game, error) {
	row := q.db.QueryRowContext(ctx, getGame, gameID)
	var i Game
	err := row.Scan(
		&i.ID,
		&i.GameID,
		&i.GameInfo,
		&i.CreatedDtTm,
		&i.CreatedUserID,
	)
	return i, err
}

const getGamesByUser = `-- name: GetGamesByUser :many
SELECT id, game_id, game_info, created_dt_tm, created_user_id FROM games
where created_user_id=?
ORDER BY created_dt_tm DESC
LIMIT 50
`

func (q *Queries) GetGamesByUser(ctx context.Context, createdUserID string) ([]Game, error) {
	rows, err := q.db.QueryContext(ctx, getGamesByUser, createdUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Game
	for rows.Next() {
		var i Game
		if err := rows.Scan(
			&i.ID,
			&i.GameID,
			&i.GameInfo,
			&i.CreatedDtTm,
			&i.CreatedUserID,
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

const getGamestate = `-- name: GetGamestate :one
SELECT id, user_id, game_id, game_state, created_dt_tm FROM gamestate
WHERE game_id = ? AND user_id = ? LIMIT 1
`

type GetGamestateParams struct {
	GameID int64
	UserID string
}

func (q *Queries) GetGamestate(ctx context.Context, arg GetGamestateParams) (Gamestate, error) {
	row := q.db.QueryRowContext(ctx, getGamestate, arg.GameID, arg.UserID)
	var i Gamestate
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.GameID,
		&i.GameState,
		&i.CreatedDtTm,
	)
	return i, err
}

const getRandomGame = `-- name: GetRandomGame :one
SELECT id, game_id, game_info, created_dt_tm, created_user_id FROM games
ORDER BY RANDOM() LIMIT 1
`

func (q *Queries) GetRandomGame(ctx context.Context) (Game, error) {
	row := q.db.QueryRowContext(ctx, getRandomGame)
	var i Game
	err := row.Scan(
		&i.ID,
		&i.GameID,
		&i.GameInfo,
		&i.CreatedDtTm,
		&i.CreatedUserID,
	)
	return i, err
}

const getRateLimit = `-- name: GetRateLimit :one
SELECT id, user_id, calls_remaining, reset_dt_tm
FROM ratelimit
WHERE user_id = ?
`

func (q *Queries) GetRateLimit(ctx context.Context, userID string) (Ratelimit, error) {
	row := q.db.QueryRowContext(ctx, getRateLimit, userID)
	var i Ratelimit
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CallsRemaining,
		&i.ResetDtTm,
	)
	return i, err
}

const updateGame = `-- name: UpdateGame :one
UPDATE games
SET game_info = ?
WHERE game_id = ?
RETURNING id, game_id, game_info, created_dt_tm, created_user_id
`

type UpdateGameParams struct {
	GameInfo string
	GameID   string
}

func (q *Queries) UpdateGame(ctx context.Context, arg UpdateGameParams) (Game, error) {
	row := q.db.QueryRowContext(ctx, updateGame, arg.GameInfo, arg.GameID)
	var i Game
	err := row.Scan(
		&i.ID,
		&i.GameID,
		&i.GameInfo,
		&i.CreatedDtTm,
		&i.CreatedUserID,
	)
	return i, err
}

const updateGamestate = `-- name: UpdateGamestate :one
UPDATE gamestate
SET game_state = ?
WHERE game_id = ? and user_id = ?
RETURNING id, user_id, game_id, game_state, created_dt_tm
`

type UpdateGamestateParams struct {
	GameState string
	GameID    int64
	UserID    string
}

func (q *Queries) UpdateGamestate(ctx context.Context, arg UpdateGamestateParams) (Gamestate, error) {
	row := q.db.QueryRowContext(ctx, updateGamestate, arg.GameState, arg.GameID, arg.UserID)
	var i Gamestate
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.GameID,
		&i.GameState,
		&i.CreatedDtTm,
	)
	return i, err
}

const updateRateLimit = `-- name: UpdateRateLimit :one
UPDATE ratelimit
SET calls_remaining = ?, reset_dt_tm = ?
WHERE user_id = ?
RETURNING id, user_id, calls_remaining, reset_dt_tm
`

type UpdateRateLimitParams struct {
	CallsRemaining int64
	ResetDtTm      string
	UserID         string
}

func (q *Queries) UpdateRateLimit(ctx context.Context, arg UpdateRateLimitParams) (Ratelimit, error) {
	row := q.db.QueryRowContext(ctx, updateRateLimit, arg.CallsRemaining, arg.ResetDtTm, arg.UserID)
	var i Ratelimit
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CallsRemaining,
		&i.ResetDtTm,
	)
	return i, err
}

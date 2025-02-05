-- name: GetGame :one
SELECT * FROM games
WHERE game_id=? LIMIT 1;

-- name: GameExists :one
SELECT COUNT(*) FROM games
WHERE game_id=? LIMIT 1;

-- name: GetRandomGame :one
SELECT * FROM games
ORDER BY RANDOM() LIMIT 1; 

-- name: CreateGame :one
INSERT INTO games (
  game_id, game_info, created_dt_tm
) VALUES (
  ?, ?, ?
)
RETURNING *;

-- name: UpdateGame :one
UPDATE games
SET game_info = ?
WHERE game_id = ?
RETURNING *;

-- name: DeleteGame :exec
DELETE FROM games
WHERE game_id = ?;

-- name: GetGamestate :one
SELECT * FROM gamestate
WHERE game_id = ? AND user_id = ? LIMIT 1;

-- name: CreateGamestate :one
INSERT INTO gamestate (
  user_id, game_id, game_state, created_dt_tm
) VALUES (
  ?, ?, ?, ?
)
RETURNING *;

-- name: UpdateGamestate :one
UPDATE gamestate
SET game_state = ?
WHERE game_id = ? and user_id = ?
RETURNING *;

-- name: DeleteGamestate :exec
DELETE FROM gamestate
WHERE game_id = ?  and user_id = ?;
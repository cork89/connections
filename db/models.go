// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

type Game struct {
	ID          int64
	GameID      string
	GameInfo    string
	CreatedDtTm string
}

type Gamestate struct {
	ID          int64
	UserID      string
	GameID      int64
	GameState   string
	CreatedDtTm string
}

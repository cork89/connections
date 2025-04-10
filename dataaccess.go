package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"log"
	"time"

	dataaccess "com.github.cork89/connections/db"
	"com.github.cork89/connections/models"
	uuid "github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var ddl string

//go:embed db/gameinfo.json
var abc123 string

var queries *dataaccess.Queries

type GameInfo struct {
	GameId string        `json:"game_id"`
	Words  []models.Word `json:"words"`
}

const (
	YellowId int = 1
	GreenId  int = 2
	BlueId   int = 3
	PurpleId int = 4
)

type DataAccess interface {
	createGame(string, []models.Word, string) (string, error)
	getGame(string) ([]models.Word, int64, error)
	getGamestate(string, int64) (models.GameState, error)
	updateGamestate(gameState models.GameState, session string, id int64) error
	initGamestate(gameState models.GameState, session string, id int64) error
	getRandomGame() (string, error)
}

type RealDataAccess struct{}

var realDataAccess DataAccess = RealDataAccess{}

func initDataaccess() error {
	ctx := context.Background()
	dbFile := "file:connections.db"
	// conn, err := sql.Open("sqlite3", ":memory:")
	conn, err := sql.Open("sqlite3", dbFile)

	if err != nil {
		return err
	}

	// create tables
	if _, err := conn.ExecContext(ctx, ddl); err != nil {
		return err
	}

	queries = dataaccess.New(conn)

	_, err = queries.GetGame(ctx, "abc123")

	if err != nil {
		log.Println("failed to get game, err: ", err)
		_, err = queries.CreateGame(ctx, dataaccess.CreateGameParams{GameID: "abc123", GameInfo: abc123, CreatedDtTm: time.Now().Format(time.RFC3339)})
	}

	if err != nil {
		log.Println("failed to create game, err: ", err)
	}

	return nil
}

func (RealDataAccess) getGame(gameId string) ([]models.Word, int64, error) {
	ctx := context.Background()
	words := make([]models.Word, 0)
	game, err := queries.GetGame(ctx, gameId)

	if err != nil {
		return words, 0, err
	}

	var gameInfo GameInfo

	err = json.Unmarshal([]byte(game.GameInfo), &gameInfo)

	if err != nil {
		return words, 0, err
	}

	return gameInfo.Words, game.ID, nil
}

func (RealDataAccess) getRandomGame() (string, error) {
	ctx := context.Background()
	game, err := queries.GetRandomGame(ctx)

	if err != nil {
		return "", err
	}

	return game.GameID, nil
}

func getGamesByUser(session string) (models.MyGamesData, error) {
	ctx := context.Background()
	games, err := queries.GetGamesByUser(ctx, session)

	var myGamesData models.MyGamesData

	if err != nil {
		return myGamesData, err
	}

	myGamesData = make([]models.MyGameData, 0, len(games))

	for _, v := range games {
		var gameInfo GameInfo

		err = json.Unmarshal([]byte(v.GameInfo), &gameInfo)

		if err != nil {
			return myGamesData, err
		}

		var categories models.Categories

		for _, word := range gameInfo.Words {
			if word.Category.CategoryId == YellowId {
				categories.Yellow = word.Category.CategoryName
			} else if word.Category.CategoryId == GreenId {
				categories.Green = word.Category.CategoryName
			} else if word.Category.CategoryId == BlueId {
				categories.Blue = word.Category.CategoryName
			} else if word.Category.CategoryId == PurpleId {
				categories.Purple = word.Category.CategoryName
			}
		}

		var myGame = models.MyGameData{Categories: categories, CreatedDtTm: v.CreatedDtTm, GameId: v.GameID}
		myGamesData = append(myGamesData, myGame)
	}

	return myGamesData, nil
}

func getRecentGamesByUser(session string) (models.MyGamesData, error) {
	ctx := context.Background()
	gamestates, err := queries.GetRecentGamestatesByUser(ctx, session)

	var myGamesData models.MyGamesData

	if err != nil {
		return myGamesData, err
	}

	myGamesData = make([]models.MyGameData, 0, len(gamestates))

	for _, v := range gamestates {
		var gameState models.GameState

		err = json.Unmarshal([]byte(v.GameState), &gameState)

		if err != nil {
			return myGamesData, err
		}

		if len(gameState.Answers) == 0 {
			continue
		}

		var categories models.Categories

		for _, word := range gameState.Answers {
			if word.Category.CategoryId == YellowId {
				categories.Yellow = word.Category.CategoryName
			} else if word.Category.CategoryId == GreenId {
				categories.Green = word.Category.CategoryName
			} else if word.Category.CategoryId == BlueId {
				categories.Blue = word.Category.CategoryName
			} else if word.Category.CategoryId == PurpleId {
				categories.Purple = word.Category.CategoryName
			}
		}

		if categories.Yellow == "" {
			categories.Yellow = "-"
		}
		if categories.Green == "" {
			categories.Green = "-"
		}
		if categories.Blue == "" {
			categories.Blue = "-"
		}
		if categories.Purple == "" {
			categories.Purple = "-"
		}

		var myGame = models.MyGameData{Categories: categories, CreatedDtTm: v.CreatedDtTm, GameId: v.GameID}

		myGamesData = append(myGamesData, myGame)
	}
	return myGamesData, nil
}

func (RealDataAccess) createGame(gameId string, words []models.Word, session string) (string, error) {
	ctx := context.Background()

	gameExists, err := queries.GameExists(ctx, gameId)

	if err != nil || gameExists > 0 {
		log.Println("gameId failed to retrieve or already exists, creating unique id instead")
		gameId = ""
	}

	if gameId == "" {
		gameIdUuid, err := uuid.NewV7()

		if err != nil {
			log.Println("failed to create uuid, err: ", err)
			return "", err
		}
		gameId = gameIdUuid.String()
	}

	gameInfo := GameInfo{GameId: gameId, Words: words}

	bytes, err := json.Marshal(gameInfo)

	if err != nil {
		log.Println("failed to marshal game info, err: ", err)
		return "", err
	}

	_, err = queries.CreateGame(ctx, dataaccess.CreateGameParams{GameID: gameId, GameInfo: string(bytes), CreatedDtTm: time.Now().Format(time.RFC3339), CreatedUserID: session})

	if err != nil {
		log.Println("failed to create game, err: ", err)
		return "", err
	}
	return gameId, nil
}

func (RealDataAccess) initGamestate(gamestate models.GameState, session string, gameId int64) error {
	ctx := context.Background()

	err := queries.DeleteGamestate(ctx, dataaccess.DeleteGamestateParams{GameID: gameId, UserID: session})

	if err != nil {
		log.Println("failed to delete game state, err: ", err)
		return err
	}

	gamestatebytes, err := json.Marshal(gamestate)

	if err != nil {
		log.Println("failed to marshal gamestate, err: ", err)
		return err
	}

	_, err = queries.CreateGamestate(ctx, dataaccess.CreateGamestateParams{UserID: session, GameID: gameId, GameState: string(gamestatebytes), CreatedDtTm: time.Now().Format(time.RFC3339)})
	return err
}

func (RealDataAccess) updateGamestate(gamestate models.GameState, session string, gameId int64) error {
	ctx := context.Background()

	gamestatebytes, err := json.Marshal(gamestate)

	if err != nil {
		log.Println("failed to marshal gamestate, err: ", err)
		return err
	}

	_, err = queries.UpdateGamestate(ctx, dataaccess.UpdateGamestateParams{UserID: session, GameID: gameId, GameState: string(gamestatebytes)})
	return err
}

func (RealDataAccess) getGamestate(session string, gameId int64) (models.GameState, error) {
	ctx := context.Background()

	gamestate, err := queries.GetGamestate(ctx, dataaccess.GetGamestateParams{UserID: session, GameID: gameId})
	var state models.GameState

	if err != nil {
		log.Println("failed to retrieve gamestate, err: ", err)
		return state, err
	}

	err = json.Unmarshal([]byte(gamestate.GameState), &state)

	if err != nil {
		log.Println("failed to unmarshal game state, err: ", err)
	}

	return state, err
}

func getRatelimit(userId string) (int, time.Time, error) {
	ctx := context.Background()

	ratelimit, err := queries.GetRateLimit(ctx, userId)

	if err != nil {
		err = createRatelimit(userId)
		if err != nil {
			return 5, time.Now().UTC().Add(1 * time.Minute), nil
		} else {
			log.Println("failed to retrieve ratelimit, err: ", err)
			return 0, time.Time{}, err
		}
	}

	resetTime, err := time.Parse(time.RFC3339, ratelimit.ResetDtTm)

	if err != nil {
		log.Println("failed to parse reset time, err: ", err)
		return 0, time.Time{}, err
	}

	return int(ratelimit.CallsRemaining), resetTime, nil
}

func updateRatelimit(userId string, callsRemaining int, resetTime time.Time) error {
	ctx := context.Background()

	var resetTimeString string

	if time.Now().After(resetTime) {
		resetTimeString = time.Now().Add(1 * time.Minute).UTC().Format(time.RFC3339)
		callsRemaining = 5
	} else {
		resetTimeString = resetTime.Format(time.RFC3339)
		callsRemaining -= 1
	}

	_, err := queries.UpdateRateLimit(ctx, dataaccess.UpdateRateLimitParams{CallsRemaining: int64(callsRemaining), ResetDtTm: resetTimeString, UserID: userId})

	if err != nil {
		log.Println("failed to retrieve ratelimit, err: ", err)
	}

	return err
}

func createRatelimit(userId string) error {
	ctx := context.Background()

	resetTime := time.Now().UTC().Add(1 * time.Minute).Format(time.RFC3339)

	err := queries.CreateRateLimit(ctx, dataaccess.CreateRateLimitParams{UserID: userId, CallsRemaining: 5, ResetDtTm: resetTime})

	if err != nil {
		log.Println("failed to create rate limit, err: ", err)
	}
	return err
}

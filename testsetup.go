package main

import (
	"errors"

	"com.github.cork89/connections/models"
)

type MockDataaccess struct{}

func (MockDataaccess) createGame(gameId string, words []models.Word, session string) (string, error) {
	if gameId == "error" {
		return "", errors.New("create game error")
	}
	return "newGameID", nil
}

func (MockDataaccess) getGame(gameId string) ([]models.Word, int64, error) {
	return []models.Word{
		{Word: "word1", Category: models.Category{CategoryId: 1}},
		{Word: "word2", Category: models.Category{CategoryId: 2}},
		{Word: "word3", Category: models.Category{CategoryId: 3}},
		{Word: "word4", Category: models.Category{CategoryId: 4}},
	}, 1, nil
}

func (MockDataaccess) getGamestate(session string, id int64) (models.GameState, error) {
	return models.GameState{
		Words:            []models.Word{{Word: "test"}},
		GuessesRemaining: 4,
	}, nil
}

func (MockDataaccess) updateGamestate(gameState models.GameState, session string, id int64) error {
	return nil
}

func (MockDataaccess) initGamestate(gameState models.GameState, session string, id int64) error {
	return nil
}

func (MockDataaccess) getRandomGame() (string, error) {
	return "testGameId", nil
}

var TestDataAccess DataAccess = MockDataaccess{}

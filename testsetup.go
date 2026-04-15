package main

import (
	"errors"

	"com.github.cork89/connections/models"
)

type MockDataaccess struct{}

var testWords = []models.Word{
	{Id: 1, Word: "amber", Category: models.Category{CategoryId: 1, CategoryName: "Warm Colors"}},
	{Id: 2, Word: "gold", Category: models.Category{CategoryId: 1, CategoryName: "Warm Colors"}},
	{Id: 3, Word: "ochre", Category: models.Category{CategoryId: 1, CategoryName: "Warm Colors"}},
	{Id: 4, Word: "rust", Category: models.Category{CategoryId: 1, CategoryName: "Warm Colors"}},
	{Id: 5, Word: "fern", Category: models.Category{CategoryId: 2, CategoryName: "Plants"}},
	{Id: 6, Word: "moss", Category: models.Category{CategoryId: 2, CategoryName: "Plants"}},
	{Id: 7, Word: "reed", Category: models.Category{CategoryId: 2, CategoryName: "Plants"}},
	{Id: 8, Word: "vine", Category: models.Category{CategoryId: 2, CategoryName: "Plants"}},
	{Id: 9, Word: "azure", Category: models.Category{CategoryId: 3, CategoryName: "Blues"}},
	{Id: 10, Word: "cobalt", Category: models.Category{CategoryId: 3, CategoryName: "Blues"}},
	{Id: 11, Word: "navy", Category: models.Category{CategoryId: 3, CategoryName: "Blues"}},
	{Id: 12, Word: "teal", Category: models.Category{CategoryId: 3, CategoryName: "Blues"}},
	{Id: 13, Word: "iris", Category: models.Category{CategoryId: 4, CategoryName: "Flowers"}},
	{Id: 14, Word: "lily", Category: models.Category{CategoryId: 4, CategoryName: "Flowers"}},
	{Id: 15, Word: "rose", Category: models.Category{CategoryId: 4, CategoryName: "Flowers"}},
	{Id: 16, Word: "tulip", Category: models.Category{CategoryId: 4, CategoryName: "Flowers"}},
}

func (MockDataaccess) createGame(gameId string, words []models.Word, session string) (string, error) {
	if gameId == "error" {
		return "", errors.New("create game error")
	}
	return "newGameID", nil
}

func (MockDataaccess) getGame(gameId string) ([]models.Word, int64, error) {
	return testWords, 1, nil
}

func (MockDataaccess) getGamestate(session string, id int64) (models.GameState, error) {
	return models.GameState{
		Words:            testWords,
		GuessesRemaining: 4,
		Hints: models.Hints{
			Hints: []string{
				testWords[0].Word,
				testWords[4].Word,
				testWords[8].Word,
				testWords[12].Word,
			},
		},
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

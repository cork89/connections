package main

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"testing"

	"com.github.cork89/connections/models"
	"com.github.cork89/connections/templates"
)

func benchmarkRequest(path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, bytes.NewBuffer(body))
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")
	recorder := httptest.NewRecorder()
	checkHandler(recorder, req, newRecordingDataAccess())
	return recorder
}

func BenchmarkCheckHandlerSolve(b *testing.B) {
	body := []byte(`{"selectedIds":[1,2,3,4]}`)

	b.ReportAllocs()
	for b.Loop() {
		recorder := benchmarkRequest("/game/testGameId/check/", body)
		if recorder.Code != 200 {
			b.Fatalf("expected status 200, got %d", recorder.Code)
		}
	}
}

func BenchmarkGameBoardRender(b *testing.B) {
	words := cloneTestWords()
	gameState := models.NewGameState(words, 4, models.BuildHints(words, false))
	gameState.SetAnswers(words, words[0].Category)
	gameState.SetSelectedIDs([]int{5, 6, 7, 9})
	gameState = gameState.Hydrate(words)
	response := models.SelectedResponse{
		Result:    models.Three,
		GameState: gameState,
		Status:    models.Playing,
	}

	b.ReportAllocs()
	for b.Loop() {
		if err := templates.GameBoard(response, gameState.Hints).Render(context.Background(), io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetGameResponseExistingState(b *testing.B) {
	dataAccess := newRecordingDataAccess()
	dataAccess.getGamestateFn = func(session string, id int64) (models.GameState, error) {
		words := cloneTestWords()
		gameState := models.NewGameState(words, 2, models.BuildHints(words, true))
		gameState.SetAnswers(words, words[0].Category)
		gameState.SetAnswers(words, words[8].Category)
		gameState.SetSelectedIDs([]int{5, 6, 7, 9})
		return gameState, nil
	}

	b.ReportAllocs()
	for b.Loop() {
		req := httptest.NewRequest("GET", "/game/testGameId/", nil)
		req = withGameContext(req)
		req.SetPathValue("gameId", "testGameId")
		recorder := httptest.NewRecorder()
		response, err := getGameResponse(recorder, req, dataAccess)
		if err != nil {
			b.Fatal(err)
		}
		if response == nil {
			b.Fatal("expected response")
		}
	}
}

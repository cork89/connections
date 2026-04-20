package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"com.github.cork89/connections/models"
)

func withGameContext(req *http.Request) *http.Request {
	i18n := models.I18N{}
	i18n.English()

	ctx := context.WithValue(req.Context(), SessionCtx, "testSession")
	ctx = context.WithValue(ctx, models.I18Nctx, i18n)
	ctx = context.WithValue(ctx, models.Settingsctx, models.BitPackedSettings{Lang: models.English})

	return req.WithContext(ctx)
}

func TestExtractGameId(t *testing.T) {
	testCases := []struct {
		name        string
		gameId      string
		expected    string
		expectedErr error
	}{
		{
			name:        "Valid Game ID",
			gameId:      "valid-game-ID123",
			expected:    "valid-game-ID123",
			expectedErr: nil,
		},
		{
			name:        "Invalid Game ID - Too Short",
			gameId:      "ab",
			expected:    "",
			expectedErr: errors.New("invalid game id"),
		},
		{
			name:        "Invalid Game ID - Too Long",
			gameId:      strings.Repeat("a", 37),
			expected:    "",
			expectedErr: errors.New("invalid game id"),
		},
		{
			name:        "Invalid Game ID - Invalid Characters",
			gameId:      "invalid@game#id",
			expected:    "",
			expectedErr: errors.New("invalid game id"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/game/"+tc.gameId+"/", nil)
			req.SetPathValue("gameId", tc.gameId)
			gameId, err := ExtractGameId(req)

			if tc.expectedErr != nil {
				if err == nil || err.Error() != tc.expectedErr.Error() {
					t.Errorf("Expected error: %v, got: %v", tc.expectedErr,
						err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if gameId != tc.expected {
					t.Errorf("Expected gameId: %s, got: %s", tc.expected,
						gameId)
				}
			}
		})
	}
}

func TestRetrieveGameState(t *testing.T) {
	req := httptest.NewRequest("GET", "/game/testGameId/", nil)
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	data, err := retrieveGameState(recorder, req, TestDataAccess)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if data.Session != "testSession" {
		t.Errorf("Expected session: testSession, got: %s", data.Session)
	}

	if data.Id != 1 {
		t.Errorf("Expected id: 1, got: %d", data.Id)
	}

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code: %d, got: %d", http.StatusOK,
			recorder.Code)
	}
}

func TestCheckHandler(t *testing.T) {
	requestBody := []byte(`{"Selected": [{"id":1,"word":"first"}, {"id":2,"word":"second"}, {"id":3,"word":"third"}, {"id":4,"word":"fourth"}]}`)
	req := httptest.NewRequest("POST", "/game/testGameId/check/", bytes.NewBuffer(requestBody))
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	checkHandler(recorder, req, TestDataAccess)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code: %d, got: %d", http.StatusOK,
			recorder.Code)
	}
}

type recordingDataAccess struct {
	getGameFn         func(string) ([]models.Word, int64, error)
	getGamestateFn    func(string, int64) (models.GameState, error)
	updateGamestateFn func(models.GameState, string, int64) error
	initGamestateFn   func(models.GameState, string, int64) error
	getRandomGameFn   func() (string, error)
	updates           chan models.GameState
	inits             chan models.GameState
}

func newRecordingDataAccess() *recordingDataAccess {
	return &recordingDataAccess{
		updates: make(chan models.GameState, 4),
		inits:   make(chan models.GameState, 4),
	}
}

func (r *recordingDataAccess) createGame(gameId string, words []models.Word, session string) (string, error) {
	return "unused", nil
}

func (r *recordingDataAccess) getGame(gameId string) ([]models.Word, int64, error) {
	if r.getGameFn != nil {
		return r.getGameFn(gameId)
	}
	return cloneTestWords(), 1, nil
}

func (r *recordingDataAccess) getGamestate(session string, id int64) (models.GameState, error) {
	if r.getGamestateFn != nil {
		return r.getGamestateFn(session, id)
	}
	return models.GameState{
		Words:            cloneTestWords(),
		GuessesRemaining: 4,
		Hints:            models.Hints{Hints: []string{"amber", "fern", "azure", "iris"}},
	}, nil
}

func (r *recordingDataAccess) updateGamestate(gameState models.GameState, session string, id int64) error {
	if r.updateGamestateFn != nil {
		return r.updateGamestateFn(gameState, session, id)
	}
	r.updates <- gameState
	return nil
}

func (r *recordingDataAccess) initGamestate(gameState models.GameState, session string, id int64) error {
	if r.initGamestateFn != nil {
		return r.initGamestateFn(gameState, session, id)
	}
	r.inits <- gameState
	return nil
}

func (r *recordingDataAccess) getRandomGame() (string, error) {
	if r.getRandomGameFn != nil {
		return r.getRandomGameFn()
	}
	return "testGameId", nil
}

func waitForGameState(t *testing.T, ch <-chan models.GameState) models.GameState {
	t.Helper()

	select {
	case state := <-ch:
		return state
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for game state write")
		return models.GameState{}
	}
}

func countSelected(words []models.Word) int {
	selected := 0
	for _, word := range words {
		if word.Selected {
			selected++
		}
	}
	return selected
}

func TestCheckHandlerSolvesCategoryAndPersistsState(t *testing.T) {
	dataAccess := newRecordingDataAccess()
	requestBody := []byte(`{"Selected": [{"id":1}, {"id":2}, {"id":3}, {"id":4}]}`)
	req := httptest.NewRequest("POST", "/game/testGameId/check/", bytes.NewBuffer(requestBody))
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	checkHandler(recorder, req, dataAccess)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "Warm Colors") {
		t.Fatalf("expected solved category to be rendered, body=%q", body)
	}

	if strings.Contains(body, "One Away!") && strings.Contains(body, "oneaway hidden") {
	}

	state := waitForGameState(t, dataAccess.updates)

	if len(state.Answers) != 1 {
		t.Fatalf("expected 1 persisted answer, got %d", len(state.Answers))
	}

	if state.Answers[0].Category.CategoryName != "Warm Colors" {
		t.Fatalf("expected persisted answer category %q, got %q", "Warm Colors", state.Answers[0].Category.CategoryName)
	}

	if state.GuessesRemaining != 4 {
		t.Fatalf("expected guesses remaining to stay at 4, got %d", state.GuessesRemaining)
	}

	if countSelected(state.Words) != 0 {
		t.Fatalf("expected solved check to clear selections")
	}
}

func TestCheckHandlerThreeOfAKindShowsOneAwayAndConsumesGuess(t *testing.T) {
	dataAccess := newRecordingDataAccess()
	requestBody := []byte(`{"Selected": [{"id":1}, {"id":2}, {"id":3}, {"id":5}]}`)
	req := httptest.NewRequest("POST", "/game/testGameId/check/", bytes.NewBuffer(requestBody))
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	checkHandler(recorder, req, dataAccess)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "One Away!") {
		t.Fatalf("expected one away message in body, got %q", body)
	}

	if strings.Contains(body, "oneaway hidden") {
		t.Fatalf("expected one away message to be visible, body=%q", body)
	}

	state := waitForGameState(t, dataAccess.updates)

	if state.GuessesRemaining != 3 {
		t.Fatalf("expected guesses remaining to drop to 3, got %d", state.GuessesRemaining)
	}

	if len(state.Answers) != 0 {
		t.Fatalf("expected no solved answers, got %d", len(state.Answers))
	}

	if countSelected(state.Words) != 4 {
		t.Fatalf("expected selected words to remain marked after near miss")
	}
}

func TestCheckHandlerInvalidSelectionDoesNotPersist(t *testing.T) {
	dataAccess := newRecordingDataAccess()
	requestBody := []byte(`{"Selected": [{"id":1}, {"id":2}, {"id":3}]}`)
	req := httptest.NewRequest("POST", "/game/testGameId/check/", bytes.NewBuffer(requestBody))
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	checkHandler(recorder, req, dataAccess)

	if recorder.Body.String() != NoChange {
		t.Fatalf("expected body %q, got %q", NoChange, recorder.Body.String())
	}

	select {
	case <-dataAccess.updates:
		t.Fatal("expected no gamestate write")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestCheckHandlerAcceptsSelectedIDsOnly(t *testing.T) {
	dataAccess := newRecordingDataAccess()
	requestBody := []byte(`{"selectedIds":[1,2,3,4]}`)
	req := httptest.NewRequest("POST", "/game/testGameId/check/", bytes.NewBuffer(requestBody))
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	checkHandler(recorder, req, dataAccess)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	state := waitForGameState(t, dataAccess.updates)
	if len(state.Answers) != 1 {
		t.Fatalf("expected solved category to persist, got %d answers", len(state.Answers))
	}
}

func TestShuffleHandler(t *testing.T) {
	requestBody := []byte(`{"Selected": [{"id":1,"word":"first"}, {"id":2,"word":"second"}, {"id":3,"word":"third"}, {"id":4,"word":"fourth"}]}`)
	req := httptest.NewRequest("POST", "/game/testGameId/shuffle/", bytes.NewBuffer(requestBody))
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	shuffleHandler(recorder, req, TestDataAccess)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code: %d, got: %d", http.StatusOK,
			recorder.Code)
	}
}

func TestShuffleHandlerPersistsSelectionAndHints(t *testing.T) {
	dataAccess := newRecordingDataAccess()
	requestBody := []byte(`{"Selected": [{"id":1}, {"id":2}, {"id":3}, {"id":4}], "HintsRevealed": true}`)
	req := httptest.NewRequest("POST", "/game/testGameId/shuffle/", bytes.NewBuffer(requestBody))
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	shuffleHandler(recorder, req, dataAccess)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	state := waitForGameState(t, dataAccess.updates)

	if !state.Hints.Revealed {
		t.Fatalf("expected hints to be revealed after shuffle")
	}

	if countSelected(state.Words) != 4 {
		t.Fatalf("expected 4 selected words to persist after shuffle")
	}
}

func TestDeselectHandler(t *testing.T) {
	req := httptest.NewRequest("POST", "/game/testGameId/deselectAll/", nil)
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	deselectHandler(recorder, req, TestDataAccess)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("Expected status code: %d, got: %d", http.StatusNoContent,
			recorder.Code)
	}
}

func TestDeselectHandlerClearsPersistedSelections(t *testing.T) {
	dataAccess := newRecordingDataAccess()
	dataAccess.getGamestateFn = func(session string, id int64) (models.GameState, error) {
		words := cloneTestWords()
		words[0].Selected = true
		words[1].Selected = true
		return models.GameState{
			Words:            words,
			GuessesRemaining: 4,
			Hints:            models.Hints{Hints: []string{"amber", "fern", "azure", "iris"}},
		}, nil
	}

	req := httptest.NewRequest("POST", "/game/testGameId/deselectAll/", nil)
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	deselectHandler(recorder, req, dataAccess)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status code %d, got %d", http.StatusNoContent, recorder.Code)
	}

	state := waitForGameState(t, dataAccess.updates)
	if countSelected(state.Words) != 0 {
		t.Fatalf("expected all selections to be cleared")
	}
}

func TestResetHandler(t *testing.T) {
	req := httptest.NewRequest("POST", "/game/testGameId/reset/", nil)
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	resetHandler(recorder, req, TestDataAccess)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code: %d, got: %d", http.StatusOK,
			recorder.Code)
	}
}

func TestResetHandlerInitializesFreshState(t *testing.T) {
	dataAccess := newRecordingDataAccess()
	req := httptest.NewRequest("POST", "/game/testGameId/reset/", nil)
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	resetHandler(recorder, req, dataAccess)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	state := waitForGameState(t, dataAccess.inits)

	if state.GuessesRemaining != 4 {
		t.Fatalf("expected guesses remaining to reset to 4, got %d", state.GuessesRemaining)
	}

	if len(state.Hints.Hints) != 4 {
		t.Fatalf("expected 4 hints, got %d", len(state.Hints.Hints))
	}

	if countSelected(state.Words) != 0 {
		t.Fatalf("expected reset state to have no selected words")
	}
}

func TestGameHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/game/testGameId/", nil)
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	gameHandler(recorder, req, TestDataAccess)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code: %d, got: %d", http.StatusOK,
			recorder.Code)
	}
}

func TestGetGameResponseInitializesMissingGamestate(t *testing.T) {
	dataAccess := newRecordingDataAccess()
	dataAccess.getGamestateFn = func(session string, id int64) (models.GameState, error) {
		return models.GameState{}, errors.New("missing gamestate")
	}

	req := httptest.NewRequest("GET", "/game/testGameId/", nil)
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()
	response, err := getGameResponse(recorder, req, dataAccess)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response == nil {
		t.Fatal("expected game response")
	}

	if response.Status != models.Playing {
		t.Fatalf("expected playing status, got %q", response.Status)
	}

	if len(response.GameState.Hints.Hints) != 4 {
		t.Fatalf("expected initialized hints, got %d", len(response.GameState.Hints.Hints))
	}

	state := waitForGameState(t, dataAccess.inits)
	if state.GuessesRemaining != 4 {
		t.Fatalf("expected initialized guesses remaining to be 4, got %d", state.GuessesRemaining)
	}
}

func TestGetGameResponseDerivesHintsWithoutPersisting(t *testing.T) {
	dataAccess := newRecordingDataAccess()
	dataAccess.getGamestateFn = func(session string, id int64) (models.GameState, error) {
		return models.GameState{
			Words:            cloneTestWords(),
			GuessesRemaining: 2,
		}, nil
	}

	req := httptest.NewRequest("GET", "/game/testGameId/?debug=1", nil)
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()
	response, err := getGameResponse(recorder, req, dataAccess)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response == nil {
		t.Fatal("expected game response")
	}

	if !response.Debug {
		t.Fatalf("expected debug mode to be enabled")
	}

	if len(response.GameState.Hints.Hints) != 4 {
		t.Fatalf("expected hints to be derived, got %d", len(response.GameState.Hints.Hints))
	}

	select {
	case <-dataAccess.updates:
		t.Fatal("expected no gamestate write when hints are derived at hydrate time")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestGameHandlerRendersGameOverState(t *testing.T) {
	dataAccess := newRecordingDataAccess()
	dataAccess.getGamestateFn = func(session string, id int64) (models.GameState, error) {
		state := models.GameState{
			Words:            cloneTestWords(),
			GuessesRemaining: 0,
			Hints:            models.Hints{Hints: []string{"amber", "fern", "azure", "iris"}},
		}
		state.SetAnswers(state.Words, state.Words[0].Category)
		state.SetAnswers(state.Words, state.Words[4].Category)
		state.SetAnswers(state.Words, state.Words[8].Category)
		return state, nil
	}

	req := httptest.NewRequest("GET", "/game/testGameId/", nil)
	req = withGameContext(req)
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	gameHandler(recorder, req, dataAccess)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, string(models.Lose)) {
		t.Fatalf("expected loser message in body, got %q", body)
	}

	if !strings.Contains(body, `class="gamestatus">loser</div>`) {
		t.Fatalf("expected loser status in body, got %q", body)
	}

	if !strings.Contains(body, "Flowers") {
		t.Fatalf("expected remaining unsolved category to be revealed on game over, got %q", body)
	}
}

func TestRandomHandler(t *testing.T) {
	recorder := httptest.NewRecorder()

	randomHandler(recorder, TestDataAccess)

	if recorder.Code != http.StatusFound {
		t.Errorf("Expected status code: %d, got: %d", http.StatusFound,
			recorder.Code)
	}

	expectedLocation := "/game/testGameId/"
	location := recorder.Header().Get("Location")
	if location != expectedLocation {
		t.Errorf("Expected Location header: %s, got: %s", expectedLocation,
			location)
	}
}

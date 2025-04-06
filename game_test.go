package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	req = req.WithContext(context.WithValue(req.Context(), SessionCtx,
		"testSession"))
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
	// Mock request body
	requestBody := []byte(`{"Selected": [{"id":1,"word":"first"}, {"id":2,"word":"second"}, {"id":3,"word":"third"}, {"id":4,"word":"fourth"}]}`)
	req := httptest.NewRequest("POST", "/game/testGameId/check/", bytes.NewBuffer(requestBody))
	req = req.WithContext(context.WithValue(req.Context(), SessionCtx, "testSession"))
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	checkHandler(recorder, req, TestDataAccess)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code: %d, got: %d", http.StatusOK,
			recorder.Code)
	}
}

func TestShuffleHandler(t *testing.T) {
	requestBody := []byte(`{"Selected": [{"id":1,"word":"first"}, {"id":2,"word":"second"}, {"id":3,"word":"third"}, {"id":4,"word":"fourth"}]}`)
	req := httptest.NewRequest("POST", "/game/testGameId/shuffle/", bytes.NewBuffer(requestBody))
	req = req.WithContext(context.WithValue(req.Context(), SessionCtx, "testSession"))
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	shuffleHandler(recorder, req, TestDataAccess)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code: %d, got: %d", http.StatusOK,
			recorder.Code)
	}
}

func TestDeselectHandler(t *testing.T) {
	req := httptest.NewRequest("POST", "/game/testGameId/deselectAll/", nil)
	req = req.WithContext(context.WithValue(req.Context(), SessionCtx, "testSession"))
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	deselectHandler(recorder, req, TestDataAccess)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("Expected status code: %d, got: %d", http.StatusNoContent,
			recorder.Code)
	}
}

func TestResetHandler(t *testing.T) {
	req := httptest.NewRequest("POST", "/game/testGameId/reset/", nil)
	req = req.WithContext(context.WithValue(req.Context(), SessionCtx, "testSession"))
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	resetHandler(recorder, req, TestDataAccess)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code: %d, got: %d", http.StatusOK,
			recorder.Code)
	}
}

func TestGameHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/game/testGameId/", nil)
	req = req.WithContext(context.WithValue(req.Context(), SessionCtx, "testSession"))
	req.SetPathValue("gameId", "testGameId")

	recorder := httptest.NewRecorder()

	gameHandler(recorder, req, TestDataAccess)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code: %d, got: %d", http.StatusOK,
			recorder.Code)
	}
}

func TestRandomHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/random/", nil)
	recorder := httptest.NewRecorder()

	randomHandler(recorder, req, TestDataAccess)

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

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"com.github.cork89/connections/models"
)

func setup() {
	// main()
	badwords = []string{"badword"}
}

func TestMain(m *testing.M) {
	// Setup code here
	setup()

	// Run the tests
	exitCode := m.Run()
	// Teardown code here
	// teardown()
	os.Exit(exitCode)
}

func TestVerifyCategory_containsBadWords(t *testing.T) {
	tests := []struct {
		name     string
		category VerifyCategory
		want     bool
	}{
		{
			name: "contains bad word",
			category: VerifyCategory{
				Words: []string{"goodword", "badword"},
			},
			want: true,
		},
		{
			name: "does not contain bad word",
			category: VerifyCategory{
				Words: []string{"goodword", "anothergoodword"},
			},
			want: false,
		},
		{
			name: "empty words",
			category: VerifyCategory{
				Words: []string{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.category.containsBadWords(); got != tt.want {
				t.Errorf("VerifyCategory.containsBadWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyCategory_verifyColor(t *testing.T) {
	tests := []struct {
		name     string
		category VerifyCategory
		color    models.Color
		want     string
	}{
		{
			name: "missing category",
			category: VerifyCategory{
				Category: "",
				Words:    []string{"a", "b", "c", "d"},
			},
			color: models.Yellow,
			want:  fmt.Sprintf(string(MissingCategory), string(models.Yellow)),
		},
		{
			name: "not enough words",
			category: VerifyCategory{
				Category: "test",
				Words:    []string{"a", "b", "c"},
			},
			color: models.Yellow,
			want:  fmt.Sprintf(string(NotEnoughWords), string(models.Yellow)),
		},
		{
			name: "contains bad words",
			category: VerifyCategory{
				Category: "test",
				Words:    []string{"a", "b", "c", "badword"},
			},
			color: models.Yellow,
			want:  fmt.Sprintf(string(WordsNotAllowed), string(models.Yellow)),
		},
		{
			name: "valid category",
			category: VerifyCategory{
				Category: "test",
				Words:    []string{"a", "b", "c", "d"},
			},
			color: models.Yellow,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.category.verifyColor(tt.color)
			if !strings.Contains(got, tt.want) {
				t.Errorf("VerifyCategory.verifyColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerify_checkDuplicates(t *testing.T) {
	tests := []struct {
		name           string
		verify         Verify
		wantDuplicates []string
		wantOK         bool
	}{
		{
			name: "no duplicates",
			verify: Verify{
				Yellow: VerifyCategory{Words: []string{"a", "b", "c", "d"}},
				Green:  VerifyCategory{Words: []string{"e", "f", "g", "h"}},
				Blue:   VerifyCategory{Words: []string{"i", "j", "k", "l"}},
				Purple: VerifyCategory{Words: []string{"m", "n", "o", "p"}},
			},
			wantDuplicates: nil,
			wantOK:         false,
		},
		{
			name: "has duplicates",
			verify: Verify{
				Yellow: VerifyCategory{Words: []string{"a", "b", "c", "d"}},
				Green:  VerifyCategory{Words: []string{"e", "f", "g", "a"}},
				Blue:   VerifyCategory{Words: []string{"i", "j", "k", "l"}},
				Purple: VerifyCategory{Words: []string{"m", "n", "o", "p"}},
			},
			wantDuplicates: []string{"A"},
			wantOK:         true,
		},
		{
			name: "all duplicates",
			verify: Verify{
				Yellow: VerifyCategory{Words: []string{"a", "a", "a", "a"}},
				Green:  VerifyCategory{Words: []string{"a", "a", "a", "a"}},
				Blue:   VerifyCategory{Words: []string{"a", "a", "a", "a"}},
				Purple: VerifyCategory{Words: []string{"a", "a", "a", "a"}},
			},
			wantDuplicates: []string{"A"},
			wantOK:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duplicates, ok := tt.verify.checkDuplicates()
			if ok != tt.wantOK {
				t.Errorf("Verify.checkDuplicates() ok = %v, want %v", ok, tt.wantOK)
			}

			if tt.wantOK {
				if len(duplicates) != len(tt.wantDuplicates) {
					t.Errorf("Verify.checkDuplicates() duplicates = %v, want %v", duplicates, tt.wantDuplicates)
				}
			}
		})
	}
}

func TestVerify_verify(t *testing.T) {
	tests := []struct {
		name   string
		verify Verify
		want   string
	}{
		{
			name: "all valid",
			verify: Verify{
				Yellow: VerifyCategory{Category: "yellow", Words: []string{"a", "b", "c", "d"}},
				Green:  VerifyCategory{Category: "green", Words: []string{"e", "f", "g", "h"}},
				Blue:   VerifyCategory{Category: "blue", Words: []string{"i", "j", "k", "l"}},
				Purple: VerifyCategory{Category: "purple", Words: []string{"m", "n", "o", "p"}},
			},
			want: "",
		},
		{
			name: "missing category and duplicate",
			verify: Verify{
				Yellow: VerifyCategory{Category: "", Words: []string{"a", "b", "c", "d"}},
				Green:  VerifyCategory{Category: "green", Words: []string{"e", "f", "g", "a"}},
				Blue:   VerifyCategory{Category: "blue", Words: []string{"i", "j", "k", "l"}},
				Purple: VerifyCategory{Category: "purple", Words: []string{"m", "n", "o", "p"}},
			},
			want: fmt.Sprintf("%s; %s", fmt.Sprintf(string(MissingCategory), string(models.Yellow)), fmt.Sprintf(string(DuplicateWords), "A")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.verify.verify()
			if !strings.Contains(got, tt.want) {
				t.Errorf("Verify.verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyRequest(t *testing.T) {
	tests := []struct {
		name            string
		requestBody     string
		contentLength   string
		expectedStatus  int
		expectedSuccess bool
		expectedReason  string
		err             error
	}{
		{
			name:            "valid request",
			requestBody:     `{"yellow":{"category":"yellow","words":["a","b","c","d"]},"green":{"category":"green","words":["e","f","g","h"]},"blue":{"category":"blue","words":["i","j","k","l"]},"purple":{"category":"purple","words":["m","n","o","p"]}}`,
			contentLength:   "200",
			expectedStatus:  http.StatusOK,
			expectedSuccess: true,
			err:             nil,
		},
		{
			name:            "invalid request - content too long",
			requestBody:     `{"yellow":{"category":"yellow","words":["a","b","c","d"]},"green":{"category":"green","words":["e","f","g","h"]},"blue":{"category":"blue","words":["i","j","k","l"]},"purple":{"category":"purple","words":["m","n","o","p"]}}`,
			contentLength:   "6000",
			expectedStatus:  http.StatusBadRequest,
			expectedSuccess: false,
			err:             errors.New("content too long"),
		},
		{
			name:            "invalid request - missing category",
			requestBody:     `{"yellow":{"category":"","words":["a","b","c","d"]},"green":{"category":"green","words":["e","f","g","h"]},"blue":{"category":"blue","words":["i","j","k","l"]},"purple":{"category":"purple","words":["m","n","o","p"]}}`,
			contentLength:   "200",
			expectedStatus:  http.StatusUnprocessableEntity,
			expectedSuccess: false,
			expectedReason:  fmt.Sprintf(string(MissingCategory), string(models.Yellow)),
			err:             nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/verify", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Length", tt.contentLength)

			w := httptest.NewRecorder()

			response, err := verifyRequest(w, req)

			if tt.err != nil {
				if err == nil || err.Error() != tt.err.Error() {
					t.Errorf("Expected error %v, got %v", tt.err, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			if response.Success != tt.expectedSuccess {
				t.Errorf("Expected success %v, got %v", tt.expectedSuccess, response.Success)
			}

			if !strings.Contains(response.FailureReason, tt.expectedReason) {
				t.Errorf("Expected reason %v, got %v", tt.expectedReason,
					response.FailureReason)
			}
		})
	}
}

func TestVerifyResponse_convertToWords(t *testing.T) {
	tests := []struct {
		name           string
		verifyResponse VerifyResponse
		expectedWords  []models.Word
	}{
		{
			name: "successful conversion",
			verifyResponse: VerifyResponse{
				Success: true,
				Verify: Verify{
					Yellow: VerifyCategory{Category: "yellow",
						Words: []string{"a", "b", "c", "d"}},
					Green: VerifyCategory{Category: "green",
						Words: []string{"e", "f", "g", "h"}},
					Blue: VerifyCategory{Category: "blue",
						Words: []string{"i", "j", "k", "l"}},
					Purple: VerifyCategory{Category: "purple",
						Words: []string{"m", "n", "o", "p"}},
				},
			},
			expectedWords: []models.Word{
				{Id: 0, Word: "a", Category: models.Category{
					CategoryId: 1, CategoryName: "yellow"}},
				{Id: 1, Word: "b", Category: models.Category{
					CategoryId: 1, CategoryName: "yellow"}},
				{Id: 2, Word: "c", Category: models.Category{
					CategoryId: 1, CategoryName: "yellow"}},
				{Id: 3, Word: "d", Category: models.Category{
					CategoryId: 1, CategoryName: "yellow"}},
				{Id: 4, Word: "e", Category: models.Category{
					CategoryId: 2, CategoryName: "green"}},
				{Id: 5, Word: "f", Category: models.Category{
					CategoryId: 2, CategoryName: "green"}},
				{Id: 6, Word: "g", Category: models.Category{
					CategoryId: 2, CategoryName: "green"}},
				{Id: 7, Word: "h", Category: models.Category{
					CategoryId: 2, CategoryName: "green"}},
				{Id: 8, Word: "i", Category: models.Category{
					CategoryId: 3, CategoryName: "blue"}},
				{Id: 9, Word: "j", Category: models.Category{
					CategoryId: 3, CategoryName: "blue"}},
				{Id: 10, Word: "k", Category: models.Category{
					CategoryId: 3, CategoryName: "blue"}},
				{Id: 11, Word: "l", Category: models.Category{
					CategoryId: 3, CategoryName: "blue"}},
				{Id: 12, Word: "m", Category: models.Category{
					CategoryId: 4, CategoryName: "purple"}},
				{Id: 13, Word: "n", Category: models.Category{
					CategoryId: 4, CategoryName: "purple"}},
				{Id: 14, Word: "o", Category: models.Category{
					CategoryId: 4, CategoryName: "purple"}},
				{Id: 15, Word: "p", Category: models.Category{
					CategoryId: 4, CategoryName: "purple"}},
			},
		},
		{
			name: "unsuccessful conversion",
			verifyResponse: VerifyResponse{
				Success: false,
			},
			expectedWords: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words := tt.verifyResponse.convertToWords()
			if tt.verifyResponse.Success {
				for i, word := range words {
					if word != tt.expectedWords[i] {
						t.Errorf("Word at index %d does not match expected.  Got %v, expected %v",
							i, word, tt.expectedWords[i])
					}
				}
			} else {
				if words != nil {
					t.Errorf("Expected nil, got %v", words)
				}
			}
		})
	}
}

func TestFilterEmptyResponses(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no empty responses",
			input:    []string{"response1", "response2", "response3"},
			expected: []string{"response1", "response2", "response3"},
		},
		{
			name:     "empty responses",
			input:    []string{"response1", "", "response3"},
			expected: []string{"response1", "response3"},
		},
		{
			name:     "all empty responses",
			input:    []string{"", "", ""},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterEmptyResponses(tt.input)
			if len(filtered) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(filtered))
			}
			for i, v := range filtered {
				if v != tt.expected[i] {
					t.Errorf("Expected %s at index %d, got %s", tt.expected[i], i, v)
				}
			}
		})
	}
}

func TestVerifyHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "valid request",
			requestBody: `{
				"yellow": {"category": "yellow", "words": ["a", "b", "c", "d"]},
				"green": {"category": "green", "words": ["e", "f", "g", "h"]},
				"blue": {"category": "blue", "words": ["i", "j", "k", "l"]},
				"purple": {"category": "purple", "words": ["m", "n", "o", "p"]}
			}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true}`,
		},
		{
			name: "invalid request - missing category",
			requestBody: `{
				"yellow": {"category": "", "words": ["a", "b", "c", "d"]},
				"green": {"category": "green", "words": ["e", "f", "g", "h"]},
				"blue": {"category": "blue", "words": ["i", "j", "k", "l"]},
				"purple": {"category": "purple", "words": ["m", "n", "o", "p"]}
			}`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"success":false,"failureReason":"Missing following categories: [yellow]"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/verify",
				bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Length", "200")

			w := httptest.NewRecorder()
			verifyHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			body := strings.TrimSuffix(w.Body.String(), "\n")
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

type MockDataaccess struct{}

func (MockDataaccess) createGame(gameId string, words []models.Word, session string) (string, error) {
	if gameId == "error" {
		return "", errors.New("create game error")
	}
	return "newGameID", nil
}

var mockDataaccess DataAccess = MockDataaccess{}

func TestCreatePostHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		sessionCtx     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "valid request",
			requestBody: `{
				"yellow": {"category": "yellow", "words": ["a", "b", "c", "d"]},
				"green": {"category": "green", "words": ["e", "f", "g", "h"]},
				"blue": {"category": "blue", "words": ["i", "j", "k", "l"]},
				"purple": {"category": "purple", "words": ["m", "n", "o", "p"]},
				"gameId": "game123"
			}`,
			sessionCtx:     "session123",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"success":true,"gameId":"newGameID"}`,
		},
		{
			name: "create game returns error",
			requestBody: `{
				"yellow": {"category": "yellow", "words": ["a", "b", "c", "d"]},
				"green": {"category": "green", "words": ["e", "f", "g", "h"]},
				"blue": {"category": "blue", "words": ["i", "j", "k", "l"]},
				"purple": {"category": "purple", "words": ["m", "n", "o", "p"]},
				"gameId": "error"
			}`,
			sessionCtx:     "session123",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   ``,
		},
		{
			name: "invalid request - missing category",
			requestBody: `{
				"yellow": {"category": "", "words": ["a", "b", "c", "d"]},
				"green": {"category": "green", "words": ["e", "f", "g", "h"]},
				"blue": {"category": "blue", "words": ["i", "j", "k", "l"]},
				"purple": {"category": "purple", "words": ["m", "n", "o", "p"]}
			}`,
			sessionCtx:     "session123",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"success":false,"failureReason":"Missing following categories: [yellow]"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/create", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Length", "200")

			ctx := req.Context()
			ctx = context.WithValue(ctx, SessionCtx, tt.sessionCtx)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			createPostHandler(w, req, mockDataaccess)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			body := strings.TrimSuffix(w.Body.String(), "\n")
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

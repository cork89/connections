package models

import (
	"encoding/json"
	"slices"
	"testing"
)

func testGameState() GameState {
	return GameState{
		Words: []Word{
			{Id: 1, Word: "amber", Category: Category{CategoryId: 1, CategoryName: "Warm Colors"}},
			{Id: 2, Word: "gold", Category: Category{CategoryId: 1, CategoryName: "Warm Colors"}},
			{Id: 3, Word: "ochre", Category: Category{CategoryId: 1, CategoryName: "Warm Colors"}},
			{Id: 4, Word: "rust", Category: Category{CategoryId: 1, CategoryName: "Warm Colors"}},
			{Id: 5, Word: "fern", Category: Category{CategoryId: 2, CategoryName: "Plants"}},
			{Id: 6, Word: "moss", Category: Category{CategoryId: 2, CategoryName: "Plants"}},
		},
		GuessesRemaining: 4,
	}
}

func TestGameStateSetAnswersMarksWordsSolvedOnce(t *testing.T) {
	gs := testGameState()
	category := gs.Words[0].Category

	gs.SetAnswers(gs.Words, category)
	gs.SetAnswers(gs.Words, category)
	gs = gs.Hydrate(gs.Words)

	if len(gs.Answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(gs.Answers))
	}

	if gs.Answers[0].Category != category {
		t.Fatalf("expected answer category %+v, got %+v", category, gs.Answers[0].Category)
	}

	if gs.Answers[0].Words != "amber, gold, ochre, rust" {
		t.Fatalf("expected concatenated answer words, got %q", gs.Answers[0].Words)
	}

	for i := 0; i < 4; i++ {
		if !gs.Words[i].Solved {
			t.Fatalf("expected word %d to be solved", gs.Words[i].Id)
		}
	}

	if gs.Words[4].Solved || gs.Words[5].Solved {
		t.Fatalf("expected non-category words to remain unsolved")
	}
}

func TestGameStateSelectionLifecycle(t *testing.T) {
	gs := testGameState()
	selected := []Word{{Id: 1}, {Id: 2}, {Id: 5}}

	gs.SetSelected(selected)
	categories := gs.GetSelectedCategories(gs.Words)
	gs = gs.Hydrate(gs.Words)

	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categories))
	}

	if len(categories[gs.Words[0].Category]) != 2 {
		t.Fatalf("expected 2 selected warm color words, got %d", len(categories[gs.Words[0].Category]))
	}

	if len(categories[gs.Words[4].Category]) != 1 {
		t.Fatalf("expected 1 selected plant word, got %d", len(categories[gs.Words[4].Category]))
	}

	for _, id := range []int{1, 2, 5} {
		found := false
		for _, word := range gs.Words {
			if word.Id == id {
				found = true
				if !word.Selected {
					t.Fatalf("expected word %d to be selected", id)
				}
			}
		}
		if !found {
			t.Fatalf("expected word %d to exist", id)
		}
	}

	gs.DeselectAll()
	gs = gs.Hydrate(gs.Words)

	for _, word := range gs.Words {
		if word.Selected {
			t.Fatalf("expected all words to be deselected")
		}
	}
}

func TestGameStateShufflePreservesWords(t *testing.T) {
	gs := testGameState()
	before := make([]int, 0, len(gs.Words))
	for _, word := range gs.Words {
		before = append(before, word.Id)
	}

	gs.Shuffle(gs.Words)
	gs = gs.Hydrate(gs.Words)

	after := make([]int, 0, len(gs.Words))
	for _, word := range gs.Words {
		after = append(after, word.Id)
	}

	slices.Sort(before)
	slices.Sort(after)

	if !slices.Equal(before, after) {
		t.Fatalf("expected shuffle to preserve word IDs, before=%v after=%v", before, after)
	}
}

func TestGameStateHydrateDerivesAnswersAndHints(t *testing.T) {
	words := testGameState().Words
	gs := GameState{
		GuessesRemaining:  4,
		Hints:             Hints{Revealed: true},
		SolvedCategoryIDs: []int{1},
		SelectedIDs:       []int{5},
		WordOrder:         []int{5, 1, 2, 3, 4, 6},
	}

	gs = gs.Hydrate(words)

	if len(gs.Answers) != 1 {
		t.Fatalf("expected 1 derived answer, got %d", len(gs.Answers))
	}

	if gs.Answers[0].Category.CategoryId != 1 {
		t.Fatalf("expected derived answer category 1, got %d", gs.Answers[0].Category.CategoryId)
	}

	if gs.Answers[0].Words != "amber, gold, ochre, rust" {
		t.Fatalf("expected derived answer words, got %q", gs.Answers[0].Words)
	}

	if len(gs.Hints.Hints) != 4 {
		t.Fatalf("expected 4 derived hints, got %d", len(gs.Hints.Hints))
	}

	if gs.Words[0].Id != 5 {
		t.Fatalf("expected hydrated order to begin with word 5, got %d", gs.Words[0].Id)
	}

	if !gs.Words[0].Selected {
		t.Fatalf("expected selected IDs to mark hydrated word as selected")
	}

	for _, word := range gs.Words {
		if word.Category.CategoryId == 1 && !word.Solved {
			t.Fatalf("expected solved category words to be marked solved")
		}
	}
}

func TestGameStateJSONRoundTripPreservesCompactFields(t *testing.T) {
	original := GameState{
		GuessesRemaining:  2,
		Hints:             Hints{Revealed: true},
		WordOrder:         []int{5, 1, 2, 3, 4, 6},
		SelectedIDs:       []int{5},
		SolvedCategoryIDs: []int{1, 3},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var roundTripped GameState
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if roundTripped.GuessesRemaining != original.GuessesRemaining {
		t.Fatalf("expected guesses remaining %d, got %d", original.GuessesRemaining, roundTripped.GuessesRemaining)
	}

	if !roundTripped.Hints.Revealed {
		t.Fatalf("expected hints revealed to round-trip")
	}

	if !slices.Equal(roundTripped.WordOrder, original.WordOrder) {
		t.Fatalf("expected word order %v, got %v", original.WordOrder, roundTripped.WordOrder)
	}

	if !slices.Equal(roundTripped.SelectedIDs, original.SelectedIDs) {
		t.Fatalf("expected selected ids %v, got %v", original.SelectedIDs, roundTripped.SelectedIDs)
	}

	if !slices.Equal(roundTripped.SolvedCategoryIDs, original.SolvedCategoryIDs) {
		t.Fatalf("expected solved category ids %v, got %v", original.SolvedCategoryIDs, roundTripped.SolvedCategoryIDs)
	}
}

func TestSelectedResponseDetermineStatus(t *testing.T) {
	testCases := []struct {
		name     string
		state    GameState
		expected GameStatus
	}{
		{
			name:     "playing",
			state:    GameState{GuessesRemaining: 3, Answers: []Answer{{}, {}}},
			expected: Playing,
		},
		{
			name:     "winner",
			state:    GameState{GuessesRemaining: 1, Answers: []Answer{{}, {}, {}, {}}},
			expected: Winner,
		},
		{
			name:     "loser",
			state:    GameState{GuessesRemaining: 0, Answers: []Answer{{}}},
			expected: Loser,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := SelectedResponse{GameState: tc.state}
			response.DetermineStatus()

			if response.Status != tc.expected {
				t.Fatalf("expected status %q, got %q", tc.expected, response.Status)
			}
		})
	}
}

func TestGameOverDataDetermineGameOver(t *testing.T) {
	testCases := []struct {
		name        string
		state       GameState
		expected    bool
		expectedMsg GameOverMessage
	}{
		{
			name:        "not over",
			state:       GameState{GuessesRemaining: 2, Answers: []Answer{{}, {}}},
			expected:    false,
			expectedMsg: Unresolved,
		},
		{
			name:        "lose",
			state:       GameState{GuessesRemaining: 0},
			expected:    true,
			expectedMsg: Lose,
		},
		{
			name:        "win",
			state:       GameState{GuessesRemaining: 1, Answers: []Answer{{}, {}, {}, {}}},
			expected:    true,
			expectedMsg: Win,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var data GameOverData
			data.DetermineGameOver(tc.state)

			if data.IsGameOver != tc.expected {
				t.Fatalf("expected game over=%v, got %v", tc.expected, data.IsGameOver)
			}

			if data.Message != tc.expectedMsg {
				t.Fatalf("expected message %q, got %q", tc.expectedMsg, data.Message)
			}
		})
	}
}

package models

import (
	"encoding/json"
	"fmt"
	"testing"
)

func benchmarkWords() []Word {
	return []Word{
		{Id: 1, Word: "amber", Category: Category{CategoryId: 1, CategoryName: "Warm Colors"}},
		{Id: 2, Word: "gold", Category: Category{CategoryId: 1, CategoryName: "Warm Colors"}},
		{Id: 3, Word: "ochre", Category: Category{CategoryId: 1, CategoryName: "Warm Colors"}},
		{Id: 4, Word: "rust", Category: Category{CategoryId: 1, CategoryName: "Warm Colors"}},
		{Id: 5, Word: "fern", Category: Category{CategoryId: 2, CategoryName: "Plants"}},
		{Id: 6, Word: "moss", Category: Category{CategoryId: 2, CategoryName: "Plants"}},
		{Id: 7, Word: "reed", Category: Category{CategoryId: 2, CategoryName: "Plants"}},
		{Id: 8, Word: "vine", Category: Category{CategoryId: 2, CategoryName: "Plants"}},
		{Id: 9, Word: "azure", Category: Category{CategoryId: 3, CategoryName: "Blues"}},
		{Id: 10, Word: "cobalt", Category: Category{CategoryId: 3, CategoryName: "Blues"}},
		{Id: 11, Word: "navy", Category: Category{CategoryId: 3, CategoryName: "Blues"}},
		{Id: 12, Word: "teal", Category: Category{CategoryId: 3, CategoryName: "Blues"}},
		{Id: 13, Word: "iris", Category: Category{CategoryId: 4, CategoryName: "Flowers"}},
		{Id: 14, Word: "lily", Category: Category{CategoryId: 4, CategoryName: "Flowers"}},
		{Id: 15, Word: "rose", Category: Category{CategoryId: 4, CategoryName: "Flowers"}},
		{Id: 16, Word: "tulip", Category: Category{CategoryId: 4, CategoryName: "Flowers"}},
	}
}

func benchmarkHints(words []Word) Hints {
	return Hints{
		Hints: []string{
			words[0].Word,
			words[4].Word,
			words[8].Word,
			words[12].Word,
		},
	}
}

func BenchmarkGameStateCheckMiss(b *testing.B) {
	words := benchmarkWords()
	selectedIDs := []int{1, 2, 3, 5}

	b.ReportAllocs()
	for b.Loop() {
		gameState := NewGameState(words, 4, benchmarkHints(words))
		gameState.SetSelectedIDs(selectedIDs)
		_ = gameState.GetSelectedCategories(words)
		gameState.Hydrate(words)
	}
}

func BenchmarkGameStateCheckSolve(b *testing.B) {
	words := benchmarkWords()
	selectedIDs := []int{1, 2, 3, 4}
	category := words[0].Category

	b.ReportAllocs()
	for b.Loop() {
		gameState := NewGameState(words, 4, benchmarkHints(words))
		gameState.SetSelectedIDs(selectedIDs)
		gameState.SetAnswers(words, category).DeselectAll().Shuffle(words)
		gameState.Hydrate(words)
	}
}

func benchmarkCompactGameState(words []Word) GameState {
	gameState := NewGameState(words, 2, BuildHints(words, true))
	gameState.WordOrder = []int{5, 1, 2, 3, 4, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	gameState.SelectedIDs = []int{5, 6, 7, 9}
	gameState.SolvedCategoryIDs = []int{1, 3}
	return gameState.Hydrate(words)
}

func benchmarkLegacyGameState(words []Word) GameState {
	gameState := NewGameState(words, 2, BuildHints(words, true))
	gameState.SetSelectedIDs([]int{5, 6, 7, 9})
	gameState.SetAnswers(words, words[0].Category)
	gameState.SetAnswers(words, words[8].Category)
	gameState = gameState.Hydrate(words)
	return GameState{
		Answers:          gameState.Answers,
		Words:            gameState.Words,
		GuessesRemaining: gameState.GuessesRemaining,
		Hints:            gameState.Hints,
	}
}

func BenchmarkGameStateJSONMarshal(b *testing.B) {
	words := benchmarkWords()
	gameState := benchmarkCompactGameState(words)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := json.Marshal(gameState); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGameStateJSONUnmarshalCompact(b *testing.B) {
	words := benchmarkWords()
	data, err := json.Marshal(benchmarkCompactGameState(words))
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		var gameState GameState
		if err := json.Unmarshal(data, &gameState); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGameStateJSONUnmarshalLegacy(b *testing.B) {
	words := benchmarkWords()
	data, err := json.Marshal(benchmarkLegacyGameState(words))
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		var gameState GameState
		if err := json.Unmarshal(data, &gameState); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGameStateHydrateProgressiveState(b *testing.B) {
	words := benchmarkWords()

	for solvedCount := 0; solvedCount <= 4; solvedCount++ {
		b.Run(fmt.Sprintf("solved_%d", solvedCount), func(b *testing.B) {
			gameState := NewGameState(words, 4, BuildHints(words, false))
			for i := 0; i < solvedCount; i++ {
				gameState.SetAnswers(words, words[i*4].Category)
			}

			b.ReportAllocs()
			for b.Loop() {
				gameState.Hydrate(words)
			}
		})
	}
}

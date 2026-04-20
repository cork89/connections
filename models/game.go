package models

import (
	_ "embed"
	"encoding/json"
	"math/rand"
	"slices"
	"strings"
)

//go:embed static/winner.svg
var winnersvg string

//go:embed static/loser.svg
var losersvg string

type GameStatus string

const (
	Playing GameStatus = "playing"
	Winner  GameStatus = "winner"
	Loser   GameStatus = "loser"
)

type Category struct {
	CategoryId   int    `json:"category_id,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
}

type Word struct {
	Id       int      `json:"id"`
	Word     string   `json:"word"`
	Category Category `json:"category,omitempty"`
	Solved   bool     `json:"solved,omitempty"`
	Selected bool     `json:"selected,omitempty"`
}

type Answer struct {
	Category
	Words string
}

type Hints struct {
	Revealed bool
	Hints    []string `json:"-"`
}

type GameState struct {
	Answers           []Answer `json:"-"`
	GuessesRemaining  int
	Hints             Hints
	WordOrder         []int  `json:"word_order,omitempty"`
	SelectedIDs       []int  `json:"selected_ids,omitempty"`
	SolvedCategoryIDs []int  `json:"solved_category_ids,omitempty"`
	Words             []Word `json:"-"`
}

func NewGameState(words []Word, guessesRemaining int, hints Hints) GameState {
	gameState := GameState{
		GuessesRemaining: guessesRemaining,
		Hints:            hints,
	}
	gameState.ensureWordOrder(words)
	return gameState.Hydrate(words)
}

func (gs *GameState) AnswerIsNew(category Category) bool {
	for _, solvedCategoryID := range gs.SolvedCategoryIDs {
		if solvedCategoryID == category.CategoryId {
			return false
		}
	}
	return true
}

func (gs *GameState) setSolvedCategory(categoryID int) *GameState {
	if !slices.Contains(gs.SolvedCategoryIDs, categoryID) {
		gs.SolvedCategoryIDs = append(gs.SolvedCategoryIDs, categoryID)
	}
	return gs
}

func (gs *GameState) SetAnswers(words []Word, category Category) *GameState {
	if !gs.AnswerIsNew(category) {
		return gs
	}

	for _, word := range words {
		if word.Category.CategoryId == category.CategoryId {
			gs.setSolvedCategory(word.Category.CategoryId)
			break
		}
	}
	return gs
}

func (gs *GameState) SetSelectedIDs(selectedIDs []int) *GameState {
	gs.SelectedIDs = gs.SelectedIDs[:0]
	for _, selectedID := range selectedIDs {
		if !slices.Contains(gs.SelectedIDs, selectedID) {
			gs.SelectedIDs = append(gs.SelectedIDs, selectedID)
		}
	}
	return gs
}

func (gs *GameState) SetSelected(selected []Word) *GameState {
	selectedIDs := make([]int, 0, len(selected))
	for _, selectedWord := range selected {
		selectedIDs = append(selectedIDs, selectedWord.Id)
	}
	return gs.SetSelectedIDs(selectedIDs)
}

func (gs *GameState) GetSelectedCategories(words []Word) map[Category][]Word {
	var categories = make(map[Category][]Word, 4)

	for _, word := range words {
		if !slices.Contains(gs.SelectedIDs, word.Id) {
			continue
		}

		_, ok := categories[word.Category]
		if !ok {
			categories[word.Category] = []Word{word}
		} else {
			categories[word.Category] = append(categories[word.Category], word)
		}
	}
	return categories
}

func (gs *GameState) GetAllCategories(words []Word) map[Category][]Word {
	var categories = make(map[Category][]Word, 4)

	for _, word := range words {
		categories[word.Category] = append(categories[word.Category], word)
	}

	return categories
}

func (gs *GameState) DeselectAll() *GameState {
	gs.SelectedIDs = gs.SelectedIDs[:0]
	return gs
}

func (gs *GameState) Shuffle(words []Word) *GameState {
	gs.ensureWordOrder(words)
	for i := len(gs.WordOrder) - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		gs.WordOrder[i], gs.WordOrder[j] = gs.WordOrder[j], gs.WordOrder[i]
	}
	return gs
}

func (gs GameState) Hydrate(words []Word) GameState {
	gs.ensureWordOrder(words)
	if len(gs.Hints.Hints) != 4 {
		gs.Hints = BuildHints(words, gs.Hints.Revealed)
	}
	gs.Answers = deriveAnswers(words, gs.SolvedCategoryIDs)

	hydratedWords := make([]Word, 0, len(words))
	for _, wordID := range gs.WordOrder {
		for _, candidate := range words {
			if candidate.Id != wordID {
				continue
			}

			candidate.Selected = slices.Contains(gs.SelectedIDs, candidate.Id)
			candidate.Solved = slices.Contains(gs.SolvedCategoryIDs, candidate.Category.CategoryId)
			hydratedWords = append(hydratedWords, candidate)
			break
		}
	}

	gs.Words = hydratedWords
	return gs
}

func (gs *GameState) ensureWordOrder(words []Word) {
	if len(gs.WordOrder) == len(words) {
		return
	}

	gs.WordOrder = make([]int, 0, len(words))
	for _, word := range words {
		gs.WordOrder = append(gs.WordOrder, word.Id)
	}
}

func BuildHints(words []Word, revealed bool) Hints {
	hints := make([]string, 4)
	for _, word := range words {
		categoryID := word.Category.CategoryId
		if categoryID < 1 || categoryID > 4 {
			continue
		}

		if hints[categoryID-1] == "" {
			hints[categoryID-1] = word.Word
		}
	}

	return Hints{Revealed: revealed, Hints: hints}
}

func deriveAnswers(words []Word, solvedCategoryIDs []int) []Answer {
	if len(solvedCategoryIDs) == 0 {
		return nil
	}

	answers := make([]Answer, 0, len(solvedCategoryIDs))
	for _, solvedCategoryID := range solvedCategoryIDs {
		answerWords := make([]string, 0, 4)
		var category Category
		for _, word := range words {
			if word.Category.CategoryId != solvedCategoryID {
				continue
			}

			category = word.Category
			answerWords = append(answerWords, word.Word)
		}

		if len(answerWords) == 0 {
			continue
		}

		answers = append(answers, Answer{
			Category: category,
			Words:    strings.Join(answerWords, ", "),
		})
	}

	return answers
}

func (gs *GameState) UnmarshalJSON(data []byte) error {
	type gameStateAlias struct {
		GuessesRemaining  int
		Hints             Hints
		WordOrder         []int `json:"word_order,omitempty"`
		SelectedIDs       []int `json:"selected_ids,omitempty"`
		SolvedCategoryIDs []int `json:"solved_category_ids,omitempty"`
		Words             []Word
		Answers           []Answer
	}

	var raw gameStateAlias
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	gs.GuessesRemaining = raw.GuessesRemaining
	gs.Hints = Hints{Revealed: raw.Hints.Revealed}
	gs.WordOrder = raw.WordOrder
	gs.SelectedIDs = raw.SelectedIDs
	gs.SolvedCategoryIDs = raw.SolvedCategoryIDs
	gs.Words = nil

	if len(raw.WordOrder) > 0 || len(raw.SelectedIDs) > 0 || len(raw.SolvedCategoryIDs) > 0 {
		return nil
	}

	if len(gs.WordOrder) == 0 && len(raw.Words) > 0 {
		gs.ensureWordOrder(raw.Words)
	}

	if len(gs.SelectedIDs) == 0 {
		for _, word := range raw.Words {
			if word.Selected {
				gs.SelectedIDs = append(gs.SelectedIDs, word.Id)
			}
		}
	}

	if len(gs.SolvedCategoryIDs) == 0 {
		for _, answer := range raw.Answers {
			gs.setSolvedCategory(answer.Category.CategoryId)
		}
		for _, word := range raw.Words {
			if word.Solved {
				gs.setSolvedCategory(word.Category.CategoryId)
			}
		}
	}

	return nil
}

type SelectedRequest struct {
	Selected      []Word `json:"selected"`
	SelectedIDs   []int  `json:"selectedIds"`
	HintsRevealed bool   `json:"hintsRevealed"`
}

func (sr SelectedRequest) IDs() []int {
	if len(sr.SelectedIDs) > 0 {
		return sr.SelectedIDs
	}

	selectedIDs := make([]int, 0, len(sr.Selected))
	for _, selectedWord := range sr.Selected {
		selectedIDs = append(selectedIDs, selectedWord.Id)
	}
	return selectedIDs
}

type SelectedResponse struct {
	// OneAway   bool `json:"oneaway"`
	// Success   bool `json:"success"`
	// Category  int  `json:"category"`
	Result       GuessResult
	GameState    GameState
	Debug        bool
	Status       GameStatus
	GameOverData GameOverData
}

func (sr *SelectedResponse) DetermineStatus() {
	if sr.GameState.GuessesRemaining == 0 {
		sr.Status = Loser
	} else if len(sr.GameState.SolvedCategoryIDs) == 4 || len(sr.GameState.Answers) == 4 {
		sr.Status = Winner
	} else {
		sr.Status = Playing
	}
}

type GameOverMessage string

const (
	Unresolved GameOverMessage = ""
	Win        GameOverMessage = "Winner Winner Chicken Dinner!"
	Lose       GameOverMessage = "Better Luck Next Time, Kid."
)

type GameOverData struct {
	IsGameOver bool
	Message    GameOverMessage
	Guy        string
}

func (g *GameOverData) DetermineGameOver(gamestate GameState) {
	if gamestate.GuessesRemaining == 0 {
		g.Message = Lose
		g.Guy = losersvg
		g.IsGameOver = true
	} else if len(gamestate.SolvedCategoryIDs) == 4 || len(gamestate.Answers) == 4 {
		g.Message = Win
		g.Guy = winnersvg
		g.IsGameOver = true
	}
}

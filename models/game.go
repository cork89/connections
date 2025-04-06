package models

import (
	_ "embed"
	"fmt"
	"math/rand"
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

type GameState struct {
	Answers          []Answer
	Words            []Word
	GuessesRemaining int
}

func (gs *GameState) AnswerIsNew(category Category) bool {
	if gs.Answers != nil {
		for _, answer := range gs.Answers {
			if answer.Category == category {
				return false
			}
		}
		return true
	}
	return true
}

func (gs *GameState) SetSolvedWord(idx int) *GameState {
	word := &gs.Words[idx]
	word.Solved = true
	return gs
}

func (gs *GameState) SetAnswers(category Category) *GameState {
	if !gs.AnswerIsNew(category) {
		return gs
	}

	var insert bool = true
	for i, selected := range gs.Words {
		if selected.Category == category {
			gs.SetSolvedWord(i)

			if gs.Answers == nil {
				gs.Answers = make([]Answer, 0)
			}

			if insert {
				gs.Answers = append(gs.Answers, Answer{Category: category, Words: selected.Word})
				insert = false
			} else {
				lastAnswer := gs.Answers[len(gs.Answers)-1]
				gs.Answers[len(gs.Answers)-1].Words = fmt.Sprintf("%s, %s", lastAnswer.Words, selected.Word)
			}
		}
	}
	return gs
}

func (gs *GameState) SetSelected(selected []Word) *GameState {
	for _, selectedWord := range selected {
		for i, word := range gs.Words {
			if word.Id == selectedWord.Id {
				gs.Words[i].Selected = true
			}
		}
	}
	return gs
}

func (gs *GameState) GetSelectedCategories(selected []Word) map[Category][]Word {
	var categories = make(map[Category][]Word, 4)

	for _, selectedWord := range selected {
		for i, word := range gs.Words {
			if word.Id == selectedWord.Id {
				_, ok := categories[word.Category]
				gs.Words[i].Selected = true
				if !ok {
					gameWords := []Word{word}
					categories[word.Category] = gameWords
				} else {
					categories[word.Category] = append(categories[word.Category], word)
				}
			}
		}
	}
	return categories
}

func (gs *GameState) DeselectAll() *GameState {
	for i := range gs.Words {
		gs.Words[i].Selected = false
	}
	return gs
}

func (gs *GameState) Shuffle() *GameState {
	for i := len(gs.Words) - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		gs.Words[i], gs.Words[j] = gs.Words[j], gs.Words[i]
	}
	return gs
}

type SelectedRequest struct {
	Selected []Word `json:"selected"`
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
	} else if len(sr.GameState.Answers) == 4 {
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
	} else if len(gamestate.Answers) == 4 {
		g.Message = Win
		g.Guy = winnersvg
		g.IsGameOver = true
	}
}

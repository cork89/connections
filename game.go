package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"regexp"
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

func (gs *GameState) answerIsNew(category Category) bool {
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

func (gs *GameState) setSolvedWord(idx int) *GameState {
	word := &gs.Words[idx]
	word.Solved = true
	return gs
}

func (gs *GameState) setAnswers(category Category) *GameState {
	if !gs.answerIsNew(category) {
		return gs
	}

	var insert bool = true
	for i, selected := range gs.Words {
		if selected.Category == category {
			gs.setSolvedWord(i)

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

func (gs *GameState) setSelected(selected []Word) *GameState {
	for _, selectedWord := range selected {
		for i, word := range gs.Words {
			if word.Id == selectedWord.Id {
				gs.Words[i].Selected = true
			}
		}
	}
	return gs
}

func (gs *GameState) getSelectedCategories(selected []Word) map[Category][]Word {
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

func (gs *GameState) deselectAll() *GameState {
	for i := range gs.Words {
		gs.Words[i].Selected = false
	}
	return gs
}

func (gs *GameState) shuffle() *GameState {
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

func (sr *SelectedResponse) determineStatus() {
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
	Lose       GameOverMessage = "Better Luck Next Time Kid."
)

type GameOverData struct {
	IsGameOver bool
	Message    GameOverMessage
	Guy        template.HTML
}

func (g *GameOverData) determineGameOver(gamestate GameState) {
	if gamestate.GuessesRemaining == 0 {
		g.Message = Lose
		g.Guy = template.HTML(losersvg)
		g.IsGameOver = true
	} else if len(gamestate.Answers) == 4 {
		g.Message = Win
		g.Guy = template.HTML(winnersvg)
		g.IsGameOver = true
	}
}

type RequestData struct {
	Game      []Word
	GameState GameState
	Session   string
	Id        int64
}

var check SelectedRequest

var validGameId = regexp.MustCompile("^[-_a-zA-Z0-9]{3,36}$")

func ExtractGameId(r *http.Request) (string, error) {
	m := validGameId.FindStringSubmatch(r.PathValue("gameId"))
	if len(m) != 1 {
		return "", errors.New("invalid game id")
	}
	return m[0], nil
}

func retrieveGameState(w http.ResponseWriter, r *http.Request) (RequestData, error) {
	gameId, err := ExtractGameId(r)

	var data RequestData

	if err != nil {
		log.Printf("failed to extract game id, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return data, err
	}

	session := r.Context().Value(SessionCtx).(string)

	words, id, err := getGame(gameId)

	if err != nil {
		log.Println("failed to retrieve game, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return data, err
	}

	gameState, err := getGamestate(session, id)

	if err != nil {
		log.Println("failed to retrieve game, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return data, err
	}

	return RequestData{Game: words, GameState: gameState, Id: id, Session: session}, nil
}

// Handler for /game/{gameId}/check/
// checks the state of the game board:
// if 4 words in a category are selected: "solve" that category and return board with category solved
// if 3 words in a category are selected, decrement the remaining guesses and display a "One Away" message
// if < 3 words in a category are selected, decrement the remaining guesses
func checkHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("check2Handler")

	// gameId, err := ExtractGameId(r)

	// if err != nil {
	// 	log.Printf("failed to extract game id, err=%v\n", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	requestData, err := retrieveGameState(w, r)
	if err != nil {
		return
	}

	words, gameState, session, id := requestData.Game, requestData.GameState, requestData.Session, requestData.Id

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&check); err != nil {
		log.Printf("failed to unmarshal request body, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(check.Selected) != 4 {
		w.Write([]byte(NoChange))
		return
	}

	// session := r.Context().Value(SessionCtx).(string)

	// words, id, err := getGame(gameId)

	// if err != nil {
	// 	log.Println("failed to retrieve game, err: ", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// gameState, err := getGamestate(session, id)

	// if err != nil {
	// 	log.Println("failed to retrieve game state, err: ", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	categories := gameState.deselectAll().getSelectedCategories(check.Selected)

	var checkResponse = SelectedResponse{Result: Other}
	var success bool = false
	for k, v := range categories {
		if len(v) == 3 {
			checkResponse.Result = Three
		} else if len(v) == 4 {
			gameState.setAnswers(k).deselectAll()
			checkResponse.Result = Four
			success = true
		}
	}

	if !success && gameState.GuessesRemaining > 0 {
		gameState.GuessesRemaining--
	} else if success {
		gameState.shuffle()
	}

	checkResponse.GameState = gameState
	checkResponse.determineStatus()

	gameOverData := GameOverData{IsGameOver: false}
	gameOverData.determineGameOver(gameState)
	checkResponse.GameOverData = gameOverData

	if gameOverData.IsGameOver {
		categories = checkResponse.GameState.getSelectedCategories(words)
		for k := range categories {
			checkResponse.GameState.setAnswers(k)
		}
		checkResponse.GameState.deselectAll()
		// checkResponse.GameState = gameState
	}

	go updateGamestate(gameState, session, id)

	err = tmpl["board"].ExecuteTemplate(w, "board", checkResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for /game/{gameId}/shuffle/
// shuffles the game board, saves to db, and returns shuffled board
func shuffleHandler(w http.ResponseWriter, r *http.Request) {
	requestData, err := retrieveGameState(w, r)
	if err != nil {
		return
	}

	gameState, session, id := requestData.GameState, requestData.Session, requestData.Id

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&check); err != nil {
		log.Printf("failed to unmarshal request body, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	gameState.setSelected(check.Selected).shuffle()

	go updateGamestate(gameState, session, id)

	checkResponse := SelectedResponse{GameState: gameState, Result: Other}

	err = tmpl["board"].ExecuteTemplate(w, "board", checkResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for /game/{gameId}/deselectAll/
// deselect all selected words on the game board, saves to db, and returns board
func deselectHandler(w http.ResponseWriter, r *http.Request) {
	// gameId, err := ExtractGameId(r)

	// if err != nil {
	// 	log.Printf("failed to extract game id, err=%v\n", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// session := r.Context().Value(SessionCtx).(string)

	// _, id, err := getGame(gameId)

	// if err != nil {
	// 	log.Println("failed to retrieve game, err: ", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }

	// gameState, err := getGamestate(session, id)

	// if err != nil {
	// 	fmt.Println("failed to retrieve game state, err: ", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	requestData, err := retrieveGameState(w, r)
	if err != nil {
		return
	}

	gameState, session, id := requestData.GameState, requestData.Session, requestData.Id

	gameState.deselectAll()
	err = updateGamestate(gameState, session, id)

	if err != nil {
		log.Println("failed to update game state, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Handler for /game/{gameId}/reset/
// resets gameboard to initial state, saves to db, and returns board
func resetHandler(w http.ResponseWriter, r *http.Request) {
	gameId, err := ExtractGameId(r)

	if err != nil {
		log.Printf("failed to extract game id, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session := r.Context().Value(SessionCtx).(string)

	words, id, err := getGame(gameId)

	if err != nil {
		log.Println("failed to retrieve game, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	gameState := GameState{Words: words, GuessesRemaining: 4}
	gameState.shuffle()
	err = initGamestate(gameState, session, id)

	if err != nil {
		log.Println("failed to reset game state, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	checkResponse := SelectedResponse{GameState: gameState, Result: Other}

	err = tmpl["board"].ExecuteTemplate(w, "board", checkResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for /game/{gameId}/
// retrieves the gameboard for a specific user session
func gameHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("gameHandler")
	gameId, err := ExtractGameId(r)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	words, id, err := getGame(gameId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		err = tmpl["404"].ExecuteTemplate(w, "base.html", nil)

		if err != nil {
			log.Println("failed to load 404 tmpl, err: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}

	session := r.Context().Value(SessionCtx).(string)

	gameState, err := getGamestate(session, id)

	if err != nil {
		gameState = GameState{Words: words, GuessesRemaining: 4}
		gameState.shuffle()
		go initGamestate(gameState, session, id)
	}

	gameResponse := SelectedResponse{GameState: gameState}
	gameResponse.determineStatus()

	gameOverData := GameOverData{IsGameOver: false}
	gameOverData.determineGameOver(gameState)
	gameResponse.GameOverData = gameOverData

	debugParam := r.FormValue("debug")
	if debugParam == "1" {
		gameResponse.Debug = true
	}

	err = tmpl["game"].ExecuteTemplate(w, "base.html", gameResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for /random/
// retrieve a random gameId and redirect to /game/{gameId}/
func randomHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("randomHandler")
	gameId, err := getRandomGame()

	if err != nil {
		log.Println("failed to get random game, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("/game/%s/", gameId)

	log.Printf("randomHandler: Redirecting to: %s", redirectURL)

	// http.Redirect(w, r, redirectURL, http.StatusFound)

	w.Header().Set("Location", redirectURL)
	w.WriteHeader(http.StatusFound)
	log.Println("redirect issued")
}

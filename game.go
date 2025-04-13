package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"com.github.cork89/connections/models"
	"com.github.cork89/connections/templates"
)

type RequestData struct {
	Game      []models.Word
	GameState models.GameState
	Session   string
	Id        int64
}

var validGameId = regexp.MustCompile("^[-_a-zA-Z0-9]{3,36}$")

func ExtractGameId(r *http.Request) (string, error) {
	m := validGameId.FindStringSubmatch(r.PathValue("gameId"))
	if len(m) != 1 {
		return "", errors.New("invalid game id")
	}
	return m[0], nil
}

func retrieveGameState(w http.ResponseWriter, r *http.Request, dataaccess DataAccess) (RequestData, error) {
	gameId, err := ExtractGameId(r)

	var data RequestData

	if err != nil {
		log.Printf("failed to extract game id, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return data, err
	}

	session := r.Context().Value(SessionCtx).(string)

	words, id, err := dataaccess.getGame(gameId)

	if err != nil {
		log.Println("failed to retrieve game, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return data, err
	}

	gameState, err := dataaccess.getGamestate(session, id)

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
func checkHandler(w http.ResponseWriter, r *http.Request, dataaccess DataAccess) {
	requestData, err := retrieveGameState(w, r, dataaccess)
	if err != nil {
		return
	}

	words, gameState, session, id := requestData.Game, requestData.GameState, requestData.Session, requestData.Id

	defer r.Body.Close()
	var check models.SelectedRequest

	if err := json.NewDecoder(r.Body).Decode(&check); err != nil {
		log.Printf("failed to unmarshal request body, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(check.Selected) != 4 {
		w.Write([]byte(NoChange))
		return
	}

	categories := gameState.DeselectAll().GetSelectedCategories(check.Selected)

	var checkResponse = models.SelectedResponse{Result: models.Other}
	var success bool = false
	for k, v := range categories {
		if len(v) == 3 {
			checkResponse.Result = models.Three
		} else if len(v) == 4 {
			gameState.SetAnswers(k).DeselectAll()
			checkResponse.Result = models.Four
			success = true
		}
	}

	if !success && gameState.GuessesRemaining > 0 {
		gameState.GuessesRemaining--
	} else if success {
		gameState.Shuffle()
	}

	checkResponse.GameState = gameState
	checkResponse.DetermineStatus()

	gameOverData := models.GameOverData{IsGameOver: false}
	gameOverData.DetermineGameOver(gameState)
	checkResponse.GameOverData = gameOverData

	if gameOverData.IsGameOver {
		categories = checkResponse.GameState.GetSelectedCategories(words)
		for k := range categories {
			checkResponse.GameState.SetAnswers(k)
		}
		checkResponse.GameState.DeselectAll()
		// checkResponse.GameState = gameState
	}

	go dataaccess.updateGamestate(gameState, session, id)

	gameBoard := templates.GameBoard(checkResponse)
	err = gameBoard.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for /game/{gameId}/shuffle/
// shuffles the game board, saves to db, and returns shuffled board
func shuffleHandler(w http.ResponseWriter, r *http.Request, dataacess DataAccess) {
	requestData, err := retrieveGameState(w, r, dataacess)
	if err != nil {
		return
	}

	gameState, session, id := requestData.GameState, requestData.Session, requestData.Id

	defer r.Body.Close()
	var check models.SelectedRequest

	if err := json.NewDecoder(r.Body).Decode(&check); err != nil {
		log.Printf("failed to unmarshal request body, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	gameState.SetSelected(check.Selected).Shuffle()

	go dataacess.updateGamestate(gameState, session, id)

	checkResponse := models.SelectedResponse{GameState: gameState, Result: models.Other}

	gameBoard := templates.GameBoard(checkResponse)
	err = gameBoard.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for /game/{gameId}/deselectAll/
// deselect all selected words on the game board, saves to db, and returns board
func deselectHandler(w http.ResponseWriter, r *http.Request, dataacess DataAccess) {
	requestData, err := retrieveGameState(w, r, dataacess)
	if err != nil {
		return
	}

	gameState, session, id := requestData.GameState, requestData.Session, requestData.Id

	gameState.DeselectAll()
	err = dataacess.updateGamestate(gameState, session, id)

	if err != nil {
		log.Println("failed to update game state, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Handler for /game/{gameId}/reset/
// resets gameboard to initial state, saves to db, and returns board
func resetHandler(w http.ResponseWriter, r *http.Request, dataaccess DataAccess) {
	gameId, err := ExtractGameId(r)

	if err != nil {
		log.Printf("failed to extract game id, err=%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	session := r.Context().Value(SessionCtx).(string)

	words, id, err := dataaccess.getGame(gameId)

	if err != nil {
		log.Println("failed to retrieve game, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	gameState := models.GameState{Words: words, GuessesRemaining: 4}
	gameState.Shuffle()
	err = dataaccess.initGamestate(gameState, session, id)

	if err != nil {
		log.Println("failed to reset game state, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	checkResponse := models.SelectedResponse{GameState: gameState, Result: models.Other}

	gameBoard := templates.GameBoard(checkResponse)
	err = gameBoard.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Retrieve the game board for a user session and game id
func getGameResponse(w http.ResponseWriter, r *http.Request, dataaccess DataAccess) (*models.SelectedResponse, error) {
	gameId, err := ExtractGameId(r)
	i18n := r.Context().Value(models.I18Nctx).(models.I18N)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	words, id, err := dataaccess.getGame(gameId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)

		component := templates.Base(templates.EmptyHead(), templates.Body404(), i18n)

		err = component.Render(context.Background(), w)

		if err != nil {
			log.Println("failed to load 404 tmpl, err: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, err
		}

		return nil, err
	}

	session := r.Context().Value(SessionCtx).(string)

	gameState, err := dataaccess.getGamestate(session, id)

	if err != nil {
		gameState = models.GameState{Words: words, GuessesRemaining: 4}
		gameState.Shuffle()
		go dataaccess.initGamestate(gameState, session, id)
	}

	gameResponse := models.SelectedResponse{GameState: gameState}
	gameResponse.DetermineStatus()

	gameOverData := models.GameOverData{IsGameOver: false}
	gameOverData.DetermineGameOver(gameState)
	gameResponse.GameOverData = gameOverData

	debugParam := r.FormValue("debug")
	if debugParam == "1" {
		gameResponse.Debug = true
	}
	return &gameResponse, nil
}

// Handler for /game/{gameId}/
// retrieves the gameboard for a specific user session
func gameHandler(w http.ResponseWriter, r *http.Request, dataaccess DataAccess) {
	gameResponse, err := getGameResponse(w, r, dataaccess)

	if err != nil {
		return
	}
	i18n := r.Context().Value(models.I18Nctx).(models.I18N)

	gameHead := templates.GameHead()
	gameBoard := templates.GameBoard(*gameResponse)
	gameBody := templates.GameBody(gameBoard, gameResponse.Debug)
	component := templates.Base(gameHead, gameBody, i18n)

	err = component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handler for /random/
// retrieve a random gameId and redirect to /game/{gameId}/
func randomHandler(w http.ResponseWriter, dataaccess DataAccess) {
	gameId, err := dataaccess.getRandomGame()

	if err != nil {
		log.Println("failed to get random game, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("/game/%s/", gameId)

	log.Printf("randomHandler: Redirecting to: %s", redirectURL)

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Location", redirectURL)
	w.WriteHeader(http.StatusFound)
	log.Println("redirect issued")
}

// Handler for /randomHtmx/
// retrieve a random gameId and render the game for /game/{gameId}/
func randomHtmxHandler(w http.ResponseWriter, r *http.Request, dataaccess DataAccess) {
	gameId, err := dataaccess.getRandomGame()

	if err != nil {
		log.Println("failed to get random game, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.SetPathValue("gameId", gameId)

	gameResponse, err := getGameResponse(w, r, dataaccess)

	if err != nil {
		return
	}

	w.Header().Set("HX-Push-Url", fmt.Sprintf("/game/%s/", gameId))
	w.Header().Set("Content-Type", "text/html")

	gameHead := templates.GameHead()
	gameBoard := templates.GameBoard(*gameResponse)
	gameBody := templates.GameBody(gameBoard, gameResponse.Debug)
	component := templates.BaseHtmx(gameHead, gameBody)

	err = component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// redirectURL := fmt.Sprintf("/game/%s/", gameId)

	// log.Printf("randomHandler: Redirecting to: %s", redirectURL)

	// w.Header().Set("Location", redirectURL)
	// w.WriteHeader(http.StatusFound)
	// log.Println("redirect issued")
}

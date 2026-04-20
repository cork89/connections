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
		if len(check.IDs()) != 4 {
			w.Write([]byte(NoChange))
			return
		}
	}

	selectedIDs := check.IDs()
	if len(selectedIDs) != 4 {
		w.Write([]byte(NoChange))
		return
	}

	gameState.DeselectAll().SetSelectedIDs(selectedIDs)
	categories := gameState.GetSelectedCategories(words)
	gameState.Hints.Revealed = check.HintsRevealed

	var checkResponse = models.SelectedResponse{Result: models.Other}
	var success bool = false
	for k, v := range categories {
		if len(v) == 3 {
			checkResponse.Result = models.Three
		} else if len(v) == 4 {
			gameState.SetAnswers(words, k).DeselectAll()
			checkResponse.Result = models.Four
			success = true
		}
	}

	if !success && gameState.GuessesRemaining > 0 {
		gameState.GuessesRemaining--
	} else if success {
		gameState.Shuffle(words)
	}

	gameState = gameState.Hydrate(words)
	checkResponse.GameState = gameState
	checkResponse.DetermineStatus()

	gameOverData := models.GameOverData{IsGameOver: false}
	gameOverData.DetermineGameOver(gameState)
	checkResponse.GameOverData = gameOverData

	if gameOverData.IsGameOver {
		categories = checkResponse.GameState.GetAllCategories(words)
		for k := range categories {
			checkResponse.GameState.SetAnswers(words, k)
		}
		checkResponse.GameState.DeselectAll()
		checkResponse.GameState = checkResponse.GameState.Hydrate(words)
		// checkResponse.GameState = gameState
	}

	go dataaccess.updateGamestate(gameState, session, id)

	gameBoard := templates.GameBoard(checkResponse, checkResponse.GameState.Hints)
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

	gameState.SetSelectedIDs(check.IDs()).Shuffle(requestData.Game)
	gameState.Hints.Revealed = check.HintsRevealed
	gameState = gameState.Hydrate(requestData.Game)

	go dataacess.updateGamestate(gameState, session, id)

	checkResponse := models.SelectedResponse{GameState: gameState, Result: models.Other}

	gameBoard := templates.GameBoard(checkResponse, gameState.Hints)
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
	gameState = gameState.Hydrate(requestData.Game)
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
	hints := models.BuildHints(words, false)
	gameState := models.NewGameState(words, 4, hints)
	gameState.Shuffle(words)
	gameState = gameState.Hydrate(words)
	err = dataaccess.initGamestate(gameState, session, id)

	if err != nil {
		log.Println("failed to reset game state, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	checkResponse := models.SelectedResponse{GameState: gameState, Result: models.Other}

	gameBoard := templates.GameBoard(checkResponse, hints)
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
	settings := r.Context().Value(models.Settingsctx).(models.BitPackedSettings)

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
		hints := models.BuildHints(words, false)
		if settings.UnhideHints {
			hints.Revealed = true
		}

		gameState = models.NewGameState(words, 4, hints)
		gameState.Shuffle(words)
		gameState = gameState.Hydrate(words)
		go dataaccess.initGamestate(gameState, session, id)
	}

	gameState = gameState.Hydrate(words)
	gameResponse := models.SelectedResponse{GameState: gameState}
	gameResponse.DetermineStatus()

	gameOverData := models.GameOverData{IsGameOver: false}
	gameOverData.DetermineGameOver(gameState)
	gameResponse.GameOverData = gameOverData

	if gameOverData.IsGameOver {
		categories := gameResponse.GameState.GetAllCategories(words)
		for k := range categories {
			gameResponse.GameState.SetAnswers(words, k)
		}
		gameResponse.GameState.DeselectAll()
		gameResponse.GameState = gameResponse.GameState.Hydrate(words)
	}

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
	gameBoard := templates.GameBoard(*gameResponse, gameResponse.GameState.Hints)
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
	gameBoard := templates.GameBoard(*gameResponse, gameResponse.GameState.Hints)
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

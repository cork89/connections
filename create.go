package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type CreateData struct {
	Debug bool
}

type VerifyResponse struct {
	Success       bool   `json:"success"`
	FailureReason string `json:"failureReason,omitempty"`
	Verify        Verify `json:"-"`
	GameId        string `json:"gameId,omitempty"`
}

func (v *VerifyResponse) convertToWords() []Word {
	if !v.Success {
		return nil
	}
	words := make([]Word, 0)

	colors := []VerifyCategory{v.Verify.Yellow, v.Verify.Green, v.Verify.Blue, v.Verify.Purple}

	id := 0
	for i, color := range colors {
		category := Category{CategoryId: i + 1, CategoryName: color.Category}
		for _, colorWord := range color.Words {
			word := Word{Id: id, Word: colorWord, Category: category}
			words = append(words, word)
			id++
		}
	}
	return words
}

type VerifyCategory struct {
	Category string   `json:"category"`
	Words    []string `json:"words"`
}

func (v *VerifyCategory) containsBadWords() bool {
	for _, word := range v.Words {
		if slices.Contains(badwords, word) {
			return true
		}
	}
	return false
}

func (v *VerifyCategory) verifyColor(color Color) string {
	if v.Category == "" {
		return fmt.Sprintf(string(MissingCategory), color)
	} else if len(v.Words) != 4 {
		return fmt.Sprintf(string(NotEnoughWords), color)
	} else if v.containsBadWords() {
		return fmt.Sprintf(string(WordsNotAllowed), color)
	}
	return ""
}

type Verify struct {
	Yellow VerifyCategory `json:"yellow"`
	Green  VerifyCategory `json:"green"`
	Blue   VerifyCategory `json:"blue"`
	Purple VerifyCategory `json:"purple"`
	GameId string         `json:"gameId,omitempty"`
}

func (v *Verify) checkDuplicates() ([]string, bool) {
	allWords := slices.Concat(v.Yellow.Words, v.Green.Words, v.Blue.Words, v.Purple.Words)

	wordsMap := make(map[string]bool, 0)
	duplicates := make([]string, 0)

	for _, word := range allWords {
		uppercaseWord := strings.ToUpper(word)
		_, ok := wordsMap[uppercaseWord]
		if !ok {
			wordsMap[uppercaseWord] = false
		} else {
			wordsMap[uppercaseWord] = true
		}
	}

	for k, v := range wordsMap {
		if v {
			duplicates = append(duplicates, k)
		}
	}

	if len(duplicates) > 0 {
		return duplicates, true
	}
	return nil, false
}

func filterEmptyResponses(responses []string) []string {
	var filteredResponses []string
	for _, response := range responses {
		if response != "" {
			filteredResponses = append(filteredResponses, response)
		}
	}
	return filteredResponses
}

func (v *Verify) verify() string {
	yellowResponse := v.Yellow.verifyColor(Yellow)
	greenResponse := v.Green.verifyColor(Green)
	blueResponse := v.Blue.verifyColor(Blue)
	purpleResponse := v.Purple.verifyColor(Purple)
	colorResponse := []string{yellowResponse, greenResponse, blueResponse, purpleResponse}

	duplicates, ok := v.checkDuplicates()
	if ok {
		duplicateResponse := fmt.Sprintf(string(DuplicateWords), strings.Join(duplicates, ", "))
		colorResponse = append(colorResponse, duplicateResponse)
	}

	filteredResponse := filterEmptyResponses(colorResponse)

	return strings.Join(filteredResponse, "; ")
}

func verifyRequest(w http.ResponseWriter, r *http.Request) (VerifyResponse, error) {
	var verifyResponse = VerifyResponse{Success: false}

	defer r.Body.Close()

	contentLen, err := strconv.Atoi(r.Header.Get("Content-Length"))

	if err != nil || contentLen > 5000 {
		log.Println("invalid request")
		w.WriteHeader(http.StatusBadRequest)
		return verifyResponse, errors.New("content too long")
	}

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		log.Println("failed to read body, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return verifyResponse, err
	}

	var verify Verify

	err = json.Unmarshal(bytes, &verify)

	if err != nil {
		log.Println("failed to unmarshal verify request, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return verifyResponse, err
	}

	failureReason := verify.verify()
	if failureReason != "" {
		verifyResponse.FailureReason = failureReason
	} else {
		verifyResponse.Success = true
		verifyResponse.Verify = verify
		verifyResponse.GameId = verify.GameId
	}

	return verifyResponse, nil
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	verifyResponse, err := verifyRequest(w, r)

	if err != nil {
		return
	}

	if !verifyResponse.Success {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	bytes, err := json.Marshal(verifyResponse)

	if err != nil {
		log.Println("failed to marshal verify response, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	verifyResponse, err := verifyRequest(w, r)

	if err != nil {
		return
	}

	if !verifyResponse.Success {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		words := verifyResponse.convertToWords()

		session := r.Context().Value(SessionCtx).(string)

		gameId, err := createGame(verifyResponse.GameId, words, session)

		if err != nil {
			log.Println("failed to create game, err: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		verifyResponse.GameId = gameId
		w.WriteHeader(http.StatusOK)
	}

	bytes, err := json.Marshal(verifyResponse)

	if err != nil {
		log.Println("failed to marshal verify response, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	var createData CreateData

	debugParam := r.FormValue("debug")
	if debugParam == "1" {
		createData.Debug = true
	}

	err := tmpl["create"].ExecuteTemplate(w, "base.html", createData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

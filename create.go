package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"com.github.cork89/connections/models"
	"com.github.cork89/connections/templates"
	"github.com/revrost/go-openrouter"
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

func (v *VerifyResponse) convertToWords() []models.Word {
	if !v.Success {
		return nil
	}
	words := make([]models.Word, 0)

	colors := []VerifyCategory{v.Verify.Yellow, v.Verify.Green, v.Verify.Blue, v.Verify.Purple}

	id := 0
	for i, color := range colors {
		category := models.Category{CategoryId: i + 1, CategoryName: color.Category}
		for _, colorWord := range color.Words {
			word := models.Word{Id: id, Word: colorWord, Category: category}
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

func (v *VerifyCategory) verifyColor(color models.Color) string {
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
	yellowResponse := v.Yellow.verifyColor(models.Yellow)
	greenResponse := v.Green.verifyColor(models.Green)
	blueResponse := v.Blue.verifyColor(models.Blue)
	purpleResponse := v.Purple.verifyColor(models.Purple)
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
		w.WriteHeader(http.StatusUnprocessableEntity)
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

	bytes, err := json.Marshal(verifyResponse)

	if err != nil {
		log.Println("failed to marshal verify response, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func createPostHandler(w http.ResponseWriter, r *http.Request, dataaccess DataAccess) {
	verifyResponse, err := verifyRequest(w, r)

	if err != nil {
		return
	}

	if !verifyResponse.Success {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		words := verifyResponse.convertToWords()

		session := r.Context().Value(SessionCtx).(string)

		gameId, err := dataaccess.createGame(verifyResponse.GameId, words, session)

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

func createHtmxHandler(w http.ResponseWriter, r *http.Request) {
	var createData CreateData

	debugParam := r.FormValue("debug")
	if debugParam == "1" {
		createData.Debug = true
	}
	bitPackedSettings := r.Context().Value(models.Settingsctx).(models.BitPackedSettings)

	w.Header().Set("Content-Type", "text/html")

	head := templates.CreateHead()
	body := templates.CreateBody(createData.Debug, bitPackedSettings)
	component := templates.BaseHtmx(head, body)
	err := component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	var createData CreateData
	i18n := r.Context().Value(models.I18Nctx).(models.I18N)

	debugParam := r.FormValue("debug")
	if debugParam == "1" {
		createData.Debug = true
	}
	bitPackedSettings := r.Context().Value(models.Settingsctx).(models.BitPackedSettings)

	createHead := templates.CreateHead()
	createBody := templates.CreateBody(createData.Debug, bitPackedSettings)
	component := templates.Base(createHead, createBody, i18n)

	err := component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

const instructions = "Instructions\n```\nYou are a direct assistant. Answer concisely and do not provide any internal reasoning or step-by-step thinking.\nRespond with a list of words\nThe words should be comma separated\nThe words should be related to the topic provided\nDo not respond with anything but the words\nRespond with only ascii characters\n```"

func suggestionsHandler(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		log.Println("failed to read body, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := model.CreateChatCompletion(context.Background(), openrouter.ChatCompletionRequest{
		Model:       "openai/gpt-oss-120b",
		Temperature: 0.9,
		MaxTokens:   1000,
		Messages: []openrouter.ChatCompletionMessage{
			openrouter.SystemMessage(instructions),
			openrouter.UserMessage(string(bytes)),
		},
	})

	if err != nil {
		log.Printf("failed to generate content: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(resp.Choices) < 1 || len(resp.Choices[0].Message.Content.Text) < 1 {
		log.Println("no content parts in response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	suggestion := resp.Choices[0].Message.Content.Text
	suggestions := strings.Split(suggestion, ",")

	for i := range suggestions {
		suggestions[i] = strings.TrimSpace(suggestions[i])
	}

	jsonData, err := json.Marshal(suggestions)

	if err != nil {
		log.Println("failed to marshal json, err: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonData)
}

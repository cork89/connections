package main

import (
	"context"
	"fmt"
	"net/http"

	"com.github.cork89/connections/models"
	"com.github.cork89/connections/templates"
)

func mygamesHtmxHandler(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(SessionCtx).(string)

	myGamesData, err := getGamesByUser(session)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")

	myGamesData.CreateShortLinks()

	myGamesHead := templates.MyGamesHead()
	myGamesBody := templates.MyGamesBody(myGamesData, models.MyGamesData{}, models.MyGamesDisplay{})
	component := templates.BaseHtmx(myGamesHead, myGamesBody)

	err = component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func mygamesHandler(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(SessionCtx).(string)

	queryVals := r.URL.Query()

	var display models.MyGamesDisplay
	display.DetermineDisplays(queryVals.Get("table"))
	fmt.Println(queryVals.Get("table"))
	fmt.Println(display)

	myGamesData, err := getGamesByUser(session)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	myRecentGames, err := getRecentGamesByUser(session)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	myGamesData.CreateShortLinks()
	myRecentGames.CreateShortLinks()

	myGamesHead := templates.MyGamesHead()
	myGamesBody := templates.MyGamesBody(myGamesData, myRecentGames, display)
	component := templates.Base(myGamesHead, myGamesBody)

	err = component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

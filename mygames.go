package main

import (
	"context"
	"net/http"

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
	myGamesBody := templates.MyGamesBody(myGamesData)
	component := templates.BaseHtmx(myGamesHead, myGamesBody)

	err = component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func mygamesHandler(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(SessionCtx).(string)

	myGamesData, err := getGamesByUser(session)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	myGamesData.CreateShortLinks()

	myGamesHead := templates.MyGamesHead()
	myGamesBody := templates.MyGamesBody(myGamesData)
	component := templates.Base(myGamesHead, myGamesBody)

	err = component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

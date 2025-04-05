package main

import (
	"context"
	"net/http"

	"com.github.cork89/connections/templates"
)

// type Categories struct {
// 	Yellow string
// 	Green  string
// 	Blue   string
// 	Purple string
// }

// type MyGameData struct {
// 	Categories  Categories
// 	CreatedDtTm string
// 	GameId      string
// 	ShortLink   string
// }

// type MyGamesData []MyGameData

func mygamesHandler(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(SessionCtx).(string)

	myGamesData, err := getGamesByUser(session)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	myGamesData.CreateShortLinks()

	// err = tmpl["mygames"].ExecuteTemplate(w, "base.html", myGamesData)

	myGamesHead := templates.MyGamesHead()
	myGamesBody := templates.MyGamesBody(myGamesData)
	component := templates.Base(myGamesHead, myGamesBody)

	err = component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

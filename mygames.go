package main

import (
	"fmt"
	"net/http"
	"strings"
)

type Categories struct {
	Yellow string
	Green  string
	Blue   string
	Purple string
}

type MyGameData struct {
	Categories  Categories
	CreatedDtTm string
	GameId      string
	ShortLink   string
}

type MyGamesData []MyGameData

func (mgd *MyGamesData) createShortLinks() {
	for i := 0; i < len(*mgd); i++ {
		gameId := (*mgd)[i].GameId

		if len(gameId) > 15 {
			(*mgd)[i].ShortLink = gameId[:strings.Index(gameId, "-")]
		} else {
			(*mgd)[i].ShortLink = gameId
		}

		createdDtTm := (*mgd)[i].CreatedDtTm
		(*mgd)[i].CreatedDtTm = createdDtTm[:strings.Index(createdDtTm, "T")]
	}
}

func mygamesHandler(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(SessionCtx).(string)

	myGamesData, err := getGamesByUser(session)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	myGamesData.createShortLinks()
	fmt.Println(myGamesData)

	err = tmpl["mygames"].ExecuteTemplate(w, "base.html", myGamesData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

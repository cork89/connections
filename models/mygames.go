package models

import "strings"

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

func (mgd *MyGamesData) CreateShortLinks() {
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

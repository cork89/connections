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

type Void string

const (
	void  Void = "void"
	empty Void = ""
)

type MyGamesDisplay struct {
	Created Void
	Recent  Void
	Checked bool
}

func (d *MyGamesDisplay) DetermineDisplays(queryVal string) {
	if queryVal == "recent" {
		d.Recent = empty
		d.Created = void
		d.Checked = true
	} else {
		d.Recent = void
		d.Created = empty
		d.Checked = false
	}
}

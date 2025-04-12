package models

type Language string

const (
	English        Language = "en"
	Spanish        Language = "es"
	French         Language = "fr"
	SUGG           int      = 0x1
	EN             int      = 0x10
	ES             int      = 0x100
	FR             int      = 0x1000
	SettingsCookie string   = "ConnectionsSettings"
)

var bitsByLang = map[Language]int{
	English: EN,
	Spanish: ES,
	French:  FR,
}

var langByBits = map[int]Language{
	EN: English,
	ES: Spanish,
	FR: French,
}

type BitPackedSettings struct {
	Lang        Language
	Suggestions bool
}

func (s BitPackedSettings) ToBitPacked() int {
	finalSettings := 0
	finalSettings |= bitsByLang[s.Lang]
	if s.Suggestions {
		finalSettings |= SUGG
	}
	return finalSettings
}

func (s *BitPackedSettings) FromBitPacked(bitPacked int) {
	english, ok := langByBits[bitPacked&EN]
	if ok {
		s.Lang = english
	}
	french, ok := langByBits[bitPacked&FR]
	if ok {
		s.Lang = french
	}
	spanish, ok := langByBits[bitPacked&ES]
	if ok {
		s.Lang = spanish
	}
	if bitPacked&SUGG == SUGG {
		s.Suggestions = true
	}
}

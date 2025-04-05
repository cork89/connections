package models

type Color string
type GuessResult string

const (
	Yellow    Color       = "yellow"
	Green     Color       = "green"
	Blue      Color       = "blue"
	Purple    Color       = "purple"
	Undefined Color       = ""
	Three     GuessResult = "three"
	Four      GuessResult = "four"
	Other     GuessResult = "other"
)

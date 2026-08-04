package game

type screenMode uint8

const (
	screenMainMenu screenMode = iota
	screenLevelSelect
	screenPuzzle
	screenReveal
	screenSettings
	screenTips
	screenEditor
	screenCommunity
)

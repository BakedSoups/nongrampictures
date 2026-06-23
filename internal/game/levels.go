package game

type levelInfo struct {
	ID    string
	Label string
	Path  string
}

var gameLevels = []levelInfo{
	{ID: "l1", Label: "L1", Path: "assets/puzzles/l1/puzzle.json"},
	{ID: "l2", Label: "L2", Path: "assets/puzzles/l2/puzzle.json"},
}

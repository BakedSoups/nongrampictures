package game

import (
	"fmt"
	"image"
	"math"
	"time"

	"github.com/alex/nongrampictures/internal/assets"
	"github.com/alex/nongrampictures/internal/nonogram"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 540
	ScreenHeight = 780

	levelSelectPages    = 2
	levelSelectPageSize = 16
)

type Game struct {
	puzzle *nonogram.Puzzle
	board  *nonogram.Board

	rowClues [][]int
	colClues [][]int

	skeleton *ebiten.Image
	reveal   *ebiten.Image
	icons    *assets.UIIcons

	skeletonPixels [][]assets.PixelCell
	revealPixels   [][]assets.PixelCell

	tool        nonogram.Tool
	mode        screenMode
	layout      boardLayout
	undoStack   [][][]nonogram.CellState
	startTime   time.Time
	timePenalty time.Duration
	completedIn time.Duration
	bestTimes   map[string]time.Duration
	levelThumbs map[string][][]assets.PixelCell

	audioEnabled      bool
	autoCorrect       bool
	penaltyFlashUntil time.Time
	correctFlashUntil time.Time
	correctFlashX     int
	correctFlashY     int

	pointerDown bool
	dragging    bool
	lastCellX   int
	lastCellY   int
	strokeState nonogram.CellState

	revealStart time.Time
	sparkles    []sparkle

	menuNotice      string
	menuNoticeUntil time.Time
	levelPage       int
}

type sparkle struct {
	x     float64
	y     float64
	delay float64
	size  float32
}

func New(puzzlePath string) (*Game, error) {
	loaded, err := assets.LoadPuzzleAssets(puzzlePath)
	if err != nil {
		return nil, err
	}
	icons, err := assets.LoadUIIcons()
	if err != nil {
		return nil, err
	}

	g := &Game{
		puzzle:         loaded.Puzzle,
		board:          nonogram.NewBoard(loaded.Puzzle.Width, loaded.Puzzle.Height),
		rowClues:       nonogram.RowClues(loaded.Puzzle.Solution),
		colClues:       nonogram.ColumnClues(loaded.Puzzle.Solution),
		skeleton:       loaded.Skeleton,
		reveal:         loaded.Reveal,
		skeletonPixels: loaded.SkeletonPixels,
		revealPixels:   loaded.RevealPixels,
		icons:          icons,
		lastCellX:      -1,
		lastCellY:      -1,
		correctFlashX:  -1,
		correctFlashY:  -1,
		startTime:      time.Now(),
		revealStart:    time.Now(),
		audioEnabled:   true,
		autoCorrect:    true,
		mode:           screenMainMenu,
		bestTimes:      loadBestTimes(),
		levelThumbs:    loadLevelThumbs(),
	}
	g.sparkles = makeSparkles()
	return g, nil
}

func loadBestTimes() map[string]time.Duration {
	best := make(map[string]time.Duration, len(gameLevels))
	for _, level := range gameLevels {
		if d := loadSavedBest(level.ID); d > 0 {
			best[level.ID] = d
		}
	}
	return best
}

func loadLevelThumbs() map[string][][]assets.PixelCell {
	thumbs := make(map[string][][]assets.PixelCell, len(gameLevels))
	for _, level := range gameLevels {
		loaded, err := assets.LoadPuzzleAssets(level.Path)
		if err == nil {
			thumbs[level.ID] = loaded.RevealPixels
		}
	}
	return thumbs
}

func (g *Game) Update() error {
	g.layout = calculateLayout(g.puzzle.Width, g.puzzle.Height)
	g.updateInput()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.draw(screen)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) pushUndo() {
	g.undoStack = append(g.undoStack, g.board.Snapshot())
	if len(g.undoStack) > 40 {
		g.undoStack = g.undoStack[1:]
	}
}

func (g *Game) undo() {
	if len(g.undoStack) == 0 {
		return
	}
	last := g.undoStack[len(g.undoStack)-1]
	g.undoStack = g.undoStack[:len(g.undoStack)-1]
	g.board.Restore(last)
}

func (g *Game) reset() {
	g.pushUndo()
	g.board.Reset()
	g.mode = screenPuzzle
}

func (g *Game) godModeFill() {
	g.pushUndo()
	for y := 0; y < g.board.Height; y++ {
		for x := 0; x < g.board.Width; x++ {
			if g.puzzle.Solution[y][x] {
				g.board.Cells[y][x] = nonogram.CellFilled
			} else {
				g.board.Cells[y][x] = nonogram.CellMarked
			}
		}
	}
	g.completePuzzle()
}

func (g *Game) loadPuzzle(path string) error {
	loaded, err := assets.LoadPuzzleAssets(path)
	if err != nil {
		return err
	}
	g.puzzle = loaded.Puzzle
	g.board = nonogram.NewBoard(loaded.Puzzle.Width, loaded.Puzzle.Height)
	g.rowClues = nonogram.RowClues(loaded.Puzzle.Solution)
	g.colClues = nonogram.ColumnClues(loaded.Puzzle.Solution)
	g.skeleton = loaded.Skeleton
	g.reveal = loaded.Reveal
	g.skeletonPixels = loaded.SkeletonPixels
	g.revealPixels = loaded.RevealPixels
	g.undoStack = nil
	g.startTime = time.Now()
	g.timePenalty = 0
	g.completedIn = 0
	g.pointerDown = false
	g.dragging = false
	g.lastCellX = -1
	g.lastCellY = -1
	g.correctFlashX = -1
	g.correctFlashY = -1
	g.strokeState = nonogram.CellEmpty
	g.mode = screenPuzzle
	return nil
}

func (g *Game) loadLevel(index int) error {
	if index < 0 || index >= len(gameLevels) {
		return fmt.Errorf("level %d is not available", index+1)
	}
	return g.loadPuzzle(gameLevels[index].Path)
}

func (g *Game) prevLevelPage() {
	if g.levelPage > 0 {
		g.levelPage--
	}
}

func (g *Game) nextLevelPage() {
	if g.levelPage < levelSelectPages-1 {
		g.levelPage++
	}
}

func (g *Game) retry() {
	g.board.Reset()
	g.undoStack = nil
	g.mode = screenPuzzle
	g.startTime = time.Now()
	g.timePenalty = 0
	g.completedIn = 0
	g.revealStart = time.Now()
}

func (g *Game) chooseBoard(size int) {
	rows := puzzleRows(size)
	p := &nonogram.Puzzle{
		ID:          fmt.Sprintf("test_%02d", size),
		Title:       fmt.Sprintf("%dx%d Board", size, size),
		Width:       size,
		Height:      size,
		SolutionRaw: rows,
		SkeletonArt: g.puzzle.SkeletonArt,
		RevealArt:   g.puzzle.RevealArt,
	}
	_ = p.ParseSolution()

	g.puzzle = p
	g.board = nonogram.NewBoard(size, size)
	g.rowClues = nonogram.RowClues(p.Solution)
	g.colClues = nonogram.ColumnClues(p.Solution)
	g.undoStack = nil
	g.mode = screenPuzzle
	g.startTime = time.Now()
	g.timePenalty = 0
	g.completedIn = 0
	g.pointerDown = false
	g.dragging = false
	g.lastCellX = -1
	g.lastCellY = -1
	g.correctFlashX = -1
	g.correctFlashY = -1
	g.strokeState = nonogram.CellEmpty
}

func (g *Game) elapsed() time.Duration {
	return time.Since(g.startTime) + g.timePenalty
}

func (g *Game) completePuzzle() {
	g.completedIn = g.elapsed()
	if g.puzzle != nil && g.puzzle.ID != "" {
		if previous := g.bestTimes[g.puzzle.ID]; previous == 0 || g.completedIn < previous {
			g.bestTimes[g.puzzle.ID] = g.completedIn
			saveBest(g.puzzle.ID, g.completedIn)
		}
	}
	g.mode = screenReveal
	g.revealStart = time.Now()
	playWebSFX("complete")
}

func (g *Game) showMenuNotice(text string) {
	g.menuNotice = text
	g.menuNoticeUntil = time.Now().Add(900 * time.Millisecond)
}

func makeSparkles() []sparkle {
	return []sparkle{
		{x: 95, y: 255, delay: 0.05, size: 2.5},
		{x: 418, y: 242, delay: 0.16, size: 2},
		{x: 122, y: 515, delay: 0.27, size: 2.4},
		{x: 420, y: 545, delay: 0.38, size: 1.8},
		{x: 270, y: 205, delay: 0.51, size: 2.2},
		{x: 310, y: 585, delay: 0.64, size: 2},
	}
}

func puzzleRows(size int) []string {
	switch size {
	case 5:
		return []string{
			"01110",
			"11111",
			"01110",
			"00100",
			"00100",
		}
	case 15:
		return []string{
			"000001111000000",
			"000011111100000",
			"000111111110000",
			"001111111111000",
			"000001111000000",
			"000001111000000",
			"000001111000000",
			"000001111000000",
			"000001111000000",
			"000000110000000",
			"000000110000000",
			"000000110000000",
			"000000110000000",
			"000001111000000",
			"000011111100000",
		}
	default:
		return []string{
			"0011110000",
			"0111111000",
			"1111111100",
			"0011110000",
			"0011110000",
			"0011110000",
			"0011110000",
			"0001100000",
			"0001100000",
			"0001100000",
		}
	}
}

func imageBounds(img *ebiten.Image) image.Rectangle {
	if img == nil {
		return image.Rect(0, 0, 1, 1)
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if w <= 0 || h <= 0 {
		return image.Rect(0, 0, 1, 1)
	}
	return image.Rect(0, 0, w, h)
}

func pulse(t float64) float64 {
	return 1 + math.Sin(t*math.Pi*2)*0.025
}

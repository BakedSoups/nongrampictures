package game

import (
	"image"
	"time"

	"github.com/alex/nongrampictures/internal/nonogram"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) updateInput() {
	if g.mode == screenMainMenu {
		g.updateMainMenuInput()
		return
	}
	if g.mode == screenLevelSelect {
		g.updateLevelSelectInput()
		return
	}
	if g.mode == screenSettings {
		g.updateSettingsInput()
		return
	}
	if g.mode == screenReveal {
		g.updateRevealInput()
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.tool = nonogram.ToolFill
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyX) || inpututil.IsKeyJustPressed(ebiten.KeyM) {
		g.tool = nonogram.ToolMark
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mode = screenMainMenu
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.godModeFill()
		return
	}

	x, y, down, justPressed, justReleased := pointerState()
	if justReleased {
		g.pointerDown = false
		g.dragging = false
		g.lastCellX = -1
		g.lastCellY = -1
		g.strokeState = nonogram.CellEmpty
	}
	if !down {
		return
	}

	if justPressed {
		switch {
		case g.layout.fillTrigger.Contains(x, y):
			g.tool = nonogram.ToolFill
			return
		case g.layout.markTrigger.Contains(x, y):
			g.tool = nonogram.ToolMark
			return
		case g.layout.godModeButton.Contains(x, y):
			g.godModeFill()
			return
		case g.layout.menuButton.Contains(x, y):
			g.mode = screenMainMenu
			return
		case g.layout.settingsButton.Contains(x, y):
			g.mode = screenSettings
			return
		}
	}

	cellX, cellY, ok := g.layout.CellAt(x, y, g.board.Width, g.board.Height)
	if !ok {
		return
	}
	if justPressed {
		g.pushUndo()
		g.pointerDown = true
		g.strokeState = nonogram.TargetState(g.tool)
		if g.board.Cells[cellY][cellX] == g.strokeState {
			g.strokeState = nonogram.CellEmpty
		}
	}
	if !g.pointerDown && !g.dragging {
		return
	}
	if cellX == g.lastCellX && cellY == g.lastCellY {
		return
	}

	next, corrected := g.correctedStrokeState(cellX, cellY, g.strokeState)
	if g.board.SetCell(cellX, cellY, next) {
		if corrected {
			g.timePenalty += 10 * time.Second
			g.penaltyFlashUntil = time.Now().Add(900 * time.Millisecond)
			g.correctFlashUntil = time.Now().Add(850 * time.Millisecond)
			g.correctFlashX = cellX
			g.correctFlashY = cellY
			playWebSFX("correct")
		} else if next == nonogram.CellFilled {
			playWebSFX("pencil")
		} else if next == nonogram.CellMarked || next == nonogram.CellEmpty {
			playWebSFX("eraser")
		}
		if nonogram.IsSolved(g.board, g.puzzle.Solution) {
			g.completePuzzle()
		}
	}
	g.dragging = true
	g.lastCellX = cellX
	g.lastCellY = cellY
}

func (g *Game) correctedStrokeState(cellX, cellY int, attempted nonogram.CellState) (nonogram.CellState, bool) {
	if !g.autoCorrect || attempted == nonogram.CellEmpty {
		return attempted, false
	}
	if attempted == nonogram.CellFilled && !g.puzzle.Solution[cellY][cellX] {
		return nonogram.CellMarked, true
	}
	if attempted == nonogram.CellMarked && g.puzzle.Solution[cellY][cellX] {
		return nonogram.CellFilled, true
	}
	return attempted, false
}

func (g *Game) updateRevealInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyR) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.retry()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyL) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mode = screenLevelSelect
		return
	}

	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	if g.layout.retryButton.Contains(x, y) {
		g.retry()
		return
	}
	if g.layout.revealLevelsButton.Contains(x, y) {
		g.mode = screenLevelSelect
	}
}

func (g *Game) updateMainMenuInput() {
	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	switch {
	case g.layout.levelSelectButton.Contains(x, y):
		g.mode = screenLevelSelect
	case g.layout.mainSettingsButton.Contains(x, y):
		g.mode = screenSettings
	}
}

func (g *Game) updateLevelSelectInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mode = screenMainMenu
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.prevLevelPage()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.nextLevelPage()
		return
	}

	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	if g.layout.levelBackButton.Contains(x, y) {
		g.mode = screenMainMenu
		return
	}
	if g.layout.levelPrevButton.Contains(x, y) {
		g.prevLevelPage()
		return
	}
	if g.layout.levelNextButton.Contains(x, y) {
		g.nextLevelPage()
		return
	}
	pageStart := g.levelPage * levelSelectPageSize
	for slot := 0; slot < levelSelectPageSize; slot++ {
		if levelTileRect(slot).Contains(x, y) {
			levelIndex := pageStart + slot
			if levelIndex < len(gameLevels) {
				_ = g.loadLevel(levelIndex)
			} else {
				g.showMenuNotice("LW")
			}
			return
		}
	}
}

func (g *Game) updateSettingsInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mode = screenMainMenu
		return
	}

	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	switch {
	case g.layout.soundButton.Contains(x, y):
		g.audioEnabled = !g.audioEnabled
		setWebMusicMuted(!g.audioEnabled)
	case g.layout.autoCorrectButton.Contains(x, y):
		g.autoCorrect = !g.autoCorrect
	case g.layout.settingsCloseButton.Contains(x, y):
		g.mode = screenMainMenu
	}
}

func pointerState() (int, int, bool, bool, bool) {
	x, y := ebiten.CursorPosition()
	down := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	justPressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	justReleased := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)

	touches := ebiten.AppendTouchIDs(nil)
	if len(touches) > 0 {
		tx, ty := ebiten.TouchPosition(touches[0])
		x, y = tx, ty
		down = true
		justPressed = inpututil.IsTouchJustReleased(touches[0]) == false && inpututil.TouchPressDuration(touches[0]) == 1
		justReleased = false
	}
	return x, y, down, justPressed, justReleased
}

type rect struct {
	x float64
	y float64
	w float64
	h float64
}

func (r rect) Contains(px, py int) bool {
	return float64(px) >= r.x && float64(px) <= r.x+r.w && float64(py) >= r.y && float64(py) <= r.y+r.h
}

func (r rect) ImageRect() image.Rectangle {
	return image.Rect(int(r.x), int(r.y), int(r.x+r.w), int(r.y+r.h))
}

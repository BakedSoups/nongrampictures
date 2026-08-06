package game

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/BakedSoups/community_nongrams/internal/assets"
	"github.com/BakedSoups/community_nongrams/internal/nonogram"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

var (
	colInk        = color.RGBA{54, 52, 49, 255}
	colMuted      = color.RGBA{116, 109, 101, 255}
	colBackdrop   = color.RGBA{202, 195, 181, 255}
	colPanel      = color.RGBA{232, 224, 207, 255}
	colPanelDark  = color.RGBA{176, 166, 151, 255}
	colCell       = color.RGBA{244, 239, 224, 255}
	colCellAlt    = color.RGBA{238, 231, 214, 255}
	colFill       = color.RGBA{63, 76, 83, 255}
	colGrid       = color.RGBA{119, 112, 102, 255}
	colGridHeavy  = color.RGBA{69, 65, 60, 255}
	colAccent     = color.RGBA{151, 83, 71, 255}
	colAccentSoft = color.RGBA{220, 145, 126, 255}
	colBlue       = color.RGBA{86, 115, 134, 255}
	colGreen      = color.RGBA{100, 132, 97, 255}
	colWhite      = color.RGBA{255, 252, 240, 255}
)

var face font.Face = basicfont.Face7x13

var renderedButtonRects []rect

type boardLayout struct {
	boardX             float64
	boardY             float64
	boardSize          float64
	cellSize           float64
	clueLeft           float64
	clueTop            float64
	fillTrigger        rect
	markTrigger        rect
	godModeButton      rect
	menuButton         rect
	settingsButton     rect
	retryButton        rect
	revealLevelsButton rect

	board5Button        rect
	board10Button       rect
	board15Button       rect
	menuCloseButton     rect
	soundButton         rect
	autoCorrectButton   rect
	settingsCloseButton rect
	levelSelectButton   rect
	mainSettingsButton  rect
	level1Button        rect
	levelTestButton     rect
	levelBackButton     rect
	levelPrevButton     rect
	levelNextButton     rect
}

func calculateLayout(width, height int) boardLayout {
	cell := math.Floor(math.Min(38, 390/math.Max(float64(width), float64(height))))
	boardW := cell * float64(width)
	boardX := math.Floor((ScreenWidth - boardW + 78) / 2)
	if boardX < 104 {
		boardX = 104
	}
	boardY := 330.0

	return boardLayout{
		boardX:              boardX,
		boardY:              boardY,
		boardSize:           boardW,
		cellSize:            cell,
		clueLeft:            boardX - 86,
		clueTop:             boardY - 132,
		fillTrigger:         rect{x: 366, y: 112, w: 54, h: 58},
		markTrigger:         rect{x: 432, y: 112, w: 54, h: 58},
		godModeButton:       rect{x: 366, y: 184, w: 120, h: 38},
		settingsButton:      rect{x: 366, y: 50, w: 58, h: 46},
		menuButton:          rect{x: 432, y: 50, w: 58, h: 46},
		retryButton:         rect{x: 80, y: 675, w: 180, h: 46},
		revealLevelsButton:  rect{x: 280, y: 675, w: 180, h: 46},
		board5Button:        rect{x: 145, y: 294, w: 250, h: 48},
		board10Button:       rect{x: 145, y: 356, w: 250, h: 48},
		board15Button:       rect{x: 145, y: 418, w: 250, h: 48},
		menuCloseButton:     rect{x: 202, y: 496, w: 136, h: 42},
		soundButton:         rect{x: 145, y: 310, w: 250, h: 48},
		autoCorrectButton:   rect{x: 145, y: 374, w: 250, h: 48},
		settingsCloseButton: rect{x: 202, y: 484, w: 136, h: 42},
		levelSelectButton:   rect{x: 128, y: 324, w: 284, h: 46},
		mainSettingsButton:  rect{x: 128, y: 398, w: 284, h: 46},
		level1Button:        rect{x: 135, y: 312, w: 270, h: 50},
		levelTestButton:     rect{x: 135, y: 380, w: 270, h: 50},
		levelBackButton:     rect{x: 202, y: 708, w: 136, h: 42},
		levelPrevButton:     rect{x: 108, y: 642, w: 92, h: 42},
		levelNextButton:     rect{x: 340, y: 642, w: 92, h: 42},
	}
}

func (l boardLayout) CellAt(px, py, width, height int) (int, int, bool) {
	x := int((float64(px) - l.boardX) / l.cellSize)
	y := int((float64(py) - l.boardY) / l.cellSize)
	return x, y, x >= 0 && y >= 0 && x < width && y < height
}

func (g *Game) draw(screen *ebiten.Image) {
	renderedButtonRects = renderedButtonRects[:0]
	screen.Fill(colPanel)

	if g.mode == screenMainMenu {
		g.drawMainMenu(screen)
		return
	}
	if g.mode == screenLevelSelect {
		g.drawLevelSelect(screen)
		return
	}
	if g.mode == screenSettings {
		g.drawSettings(screen)
		return
	}
	if g.mode == screenTips {
		g.drawTips(screen)
		return
	}
	if g.mode == screenEditor {
		g.drawEditor(screen)
		return
	}
	if g.mode == screenCommunity {
		g.drawCommunity(screen)
		return
	}
	if g.mode == screenReveal {
		g.drawReveal(screen)
		return
	}

	g.drawPuzzle(screen)
}

func (g *Game) drawPuzzle(screen *ebiten.Image) {
	drawRounded(screen, rect{x: g.layout.clueLeft - 2, y: g.layout.clueTop - 2, w: g.layout.boardX - g.layout.clueLeft + g.layout.boardSize + 8, h: g.layout.boardY - g.layout.clueTop + g.layout.boardSize + 8}, 8, colWhite)

	g.drawClues(screen)
	g.drawBoard(screen)
	g.drawStatusPanel(screen)
	g.drawToolTrigger(screen)
	g.drawTopButtons(screen)
}

func (g *Game) drawClues(screen *ebiten.Image) {
	hoverX, hoverY := g.hoverCell()
	for y := 0; y < g.board.Height; y++ {
		row := rect{x: g.layout.clueLeft, y: g.layout.boardY + float64(y)*g.layout.cellSize, w: g.layout.boardX - g.layout.clueLeft, h: g.layout.cellSize}
		c := color.RGBA{245, 245, 244, 255}
		if y%2 == 0 {
			c = color.RGBA{234, 234, 232, 255}
		}
		if y == hoverY {
			c = color.RGBA{90, 199, 229, 255}
		}
		vector.DrawFilledRect(screen, float32(row.x), float32(row.y), float32(row.w), float32(row.h), c, false)
	}
	for x := 0; x < g.board.Width; x++ {
		col := rect{x: g.layout.boardX + float64(x)*g.layout.cellSize, y: g.layout.clueTop, w: g.layout.cellSize, h: g.layout.boardY - g.layout.clueTop}
		c := color.RGBA{246, 246, 245, 255}
		if x%2 == 0 {
			c = color.RGBA{236, 236, 235, 255}
		}
		if x == hoverX {
			c = color.RGBA{90, 199, 229, 255}
		}
		vector.DrawFilledRect(screen, float32(col.x), float32(col.y), float32(col.w), float32(col.h), c, false)
	}

	for y, clues := range g.rowClues {
		label := clueLabel(clues)
		tx := int(g.layout.boardX-10) - text.BoundString(face, label).Dx()
		ty := int(g.layout.boardY + float64(y)*g.layout.cellSize + g.layout.cellSize/2 + 5)
		clueColor := color.Color(colInk)
		if g.rowSatisfied(y) {
			clueColor = colMuted
		}
		drawText(screen, label, tx, ty, clueColor)
	}
	for x, clues := range g.colClues {
		parts := make([]string, len(clues))
		for i, n := range clues {
			parts[i] = fmt.Sprint(n)
		}
		clueColor := color.Color(colInk)
		if g.columnSatisfied(x) {
			clueColor = colMuted
		}
		cx := int(g.layout.boardX + float64(x)*g.layout.cellSize + g.layout.cellSize/2)
		step := columnClueStep(len(parts))
		bottomY := int(g.layout.boardY - 10)
		startY := bottomY - (len(parts)-1)*step
		for i, part := range parts {
			drawText(screen, part, cx-text.BoundString(face, part).Dx()/2, startY+i*step, clueColor)
		}
	}
}

func (g *Game) rowSatisfied(y int) bool {
	if g.board == nil || y < 0 || y >= g.board.Height || y >= len(g.rowClues) {
		return false
	}
	filled := make([]bool, g.board.Width)
	for x := range filled {
		filled[x] = g.board.Cells[y][x] == nonogram.CellFilled
	}
	return equalClues(lineCluesFromFilled(filled), g.rowClues[y])
}

func (g *Game) columnSatisfied(x int) bool {
	if g.board == nil || x < 0 || x >= g.board.Width || x >= len(g.colClues) {
		return false
	}
	filled := make([]bool, g.board.Height)
	for y := 0; y < g.board.Height; y++ {
		filled[y] = g.board.Cells[y][x] == nonogram.CellFilled
	}
	return equalClues(lineCluesFromFilled(filled), g.colClues[x])
}

func lineCluesFromFilled(filled []bool) []int {
	clues := []int{}
	run := 0
	for _, cell := range filled {
		if cell {
			run++
			continue
		}
		if run > 0 {
			clues = append(clues, run)
			run = 0
		}
	}
	if run > 0 {
		clues = append(clues, run)
	}
	if len(clues) == 0 {
		return []int{0}
	}
	return clues
}

func equalClues(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func columnClueStep(count int) int {
	if count >= 5 {
		return 13
	}
	return 16
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	l := g.layout
	hoverX, hoverY := g.hoverCell()
	drawRounded(screen, rect{x: l.boardX - 7, y: l.boardY - 7, w: l.cellSize*float64(g.board.Width) + 14, h: l.cellSize*float64(g.board.Height) + 14}, 8, color.RGBA{95, 92, 86, 255})
	for y := 0; y < g.board.Height; y++ {
		for x := 0; x < g.board.Width; x++ {
			cellRect := rect{
				x: l.boardX + float64(x)*l.cellSize + 1,
				y: l.boardY + float64(y)*l.cellSize + 1,
				w: l.cellSize - 2,
				h: l.cellSize - 2,
			}
			c := colCell
			if (x+y)%2 == 1 {
				c = color.RGBA{248, 248, 246, 255}
			}
			if x == hoverX || y == hoverY {
				c = color.RGBA{232, 242, 240, 255}
			}
			if x == g.correctFlashX && y == g.correctFlashY && time.Now().Before(g.correctFlashUntil) {
				c = color.RGBA{255, 224, 214, 255}
			}
			vector.DrawFilledRect(screen, float32(cellRect.x), float32(cellRect.y), float32(cellRect.w), float32(cellRect.h), c, false)
			switch g.board.Cells[y][x] {
			case nonogram.CellFilled:
				drawRounded(screen, inset(cellRect, 4), 4, colFill)
			case nonogram.CellMarked:
				markInset := math.Max(2, math.Min(5, l.cellSize*0.18))
				drawX(screen, inset(cellRect, markInset), color.RGBA{245, 139, 17, 255})
			}
		}
	}

	if time.Now().Before(g.correctFlashUntil) && g.board.InBounds(g.correctFlashX, g.correctFlashY) {
		t := time.Until(g.correctFlashUntil).Seconds() / 0.85
		alpha := uint8(95 + 160*t)
		r := rect{
			x: l.boardX + float64(g.correctFlashX)*l.cellSize + 2,
			y: l.boardY + float64(g.correctFlashY)*l.cellSize + 2,
			w: l.cellSize - 4,
			h: l.cellSize - 4,
		}
		drawRectOutline(screen, r, 4, color.RGBA{226, 52, 36, alpha})
	}

	for x := 0; x <= g.board.Width; x++ {
		lineCol := colGrid
		thick := float32(1)
		if x%5 == 0 {
			lineCol = colGridHeavy
			thick = 2
		}
		xx := float32(l.boardX + float64(x)*l.cellSize)
		vector.StrokeLine(screen, xx, float32(l.boardY), xx, float32(l.boardY+l.cellSize*float64(g.board.Height)), thick, lineCol, false)
	}
	for y := 0; y <= g.board.Height; y++ {
		lineCol := colGrid
		thick := float32(1)
		if y%5 == 0 {
			lineCol = colGridHeavy
			thick = 2
		}
		yy := float32(l.boardY + float64(y)*l.cellSize)
		vector.StrokeLine(screen, float32(l.boardX), yy, float32(l.boardX+l.cellSize*float64(g.board.Width)), yy, thick, lineCol, false)
	}
}

func (g *Game) drawStatusPanel(screen *ebiten.Image) {
	panel := rect{x: 66, y: 58, w: 246, h: 88}
	drawRounded(screen, rect{x: panel.x + 4, y: panel.y + 4, w: panel.w, h: panel.h}, 8, color.RGBA{110, 104, 95, 150})
	drawRounded(screen, panel, 8, color.RGBA{35, 36, 36, 255})
	drawRounded(screen, rect{x: panel.x + 12, y: panel.y + 10, w: 96, h: 26}, 8, color.RGBA{8, 8, 8, 255})
	drawCenteredText(screen, "PUZZLE", rect{x: panel.x + 12, y: panel.y + 10, w: 96, h: 26}, colWhite)
	drawText(screen, formatTimer(g.elapsed()), int(panel.x+132), int(panel.y+31), colWhite)
	if time.Now().Before(g.penaltyFlashUntil) {
		drawText(screen, "+10s", int(panel.x+132), int(panel.y+65), color.RGBA{255, 104, 78, 255})
	}
}

func (g *Game) drawToolTrigger(screen *ebiten.Image) {
	drawTrigger(screen, g.layout.fillTrigger, g.tool == nonogram.ToolFill, colBlue, g.icons.Pencil)
	drawTrigger(screen, g.layout.markTrigger, g.tool == nonogram.ToolMark, colAccent, g.icons.X)
	drawButton(screen, g.layout.godModeButton, "GOD")
}

func drawTrigger(screen *ebiten.Image, r rect, active bool, c color.RGBA, icon *ebiten.Image) {
	registerButtonRect(r)
	base := color.RGBA{193, 184, 167, 255}
	if active {
		base = c
	}
	drawRounded(screen, rect{x: r.x + 4, y: r.y + 5, w: r.w, h: r.h}, 10, color.RGBA{132, 124, 112, 125})
	drawRounded(screen, r, 10, base)
	drawRounded(screen, inset(r, 8), 7, color.RGBA{238, 230, 211, 255})
	if active {
		drawRounded(screen, inset(r, 14), 6, c)
	}
	drawIconImage(screen, icon, inset(r, 13), 1)
}

func (g *Game) drawTopButtons(screen *ebiten.Image) {
	drawIconButton(screen, g.layout.settingsButton)
	drawIconImage(screen, g.icons.Gear, inset(g.layout.settingsButton, 10), 1)
	drawIconButton(screen, g.layout.menuButton)
	drawIconImage(screen, g.icons.Home, inset(g.layout.menuButton, 10), 1)
}

func (g *Game) drawReveal(screen *ebiten.Image) {
	elapsed := time.Since(g.revealStart).Seconds()
	drawScaledText(screen, strings.ToUpper(g.puzzle.Title), 54, 72, 1.85, colInk)
	drawScaledText(screen, "COMPLETE", 54, 116, 1.35, colAccent)
	drawText(screen, "time "+formatTimer(g.completedIn), 382, 78, colInk)

	artRect := rect{x: 118, y: 205, w: 330, h: 330}
	drawRounded(screen, rect{x: artRect.x - 12, y: artRect.y - 12, w: artRect.w + 24, h: artRect.h + 24}, 8, colGridHeavy)
	drawRounded(screen, rect{x: artRect.x - 6, y: artRect.y - 6, w: artRect.w + 12, h: artRect.h + 12}, 6, colWhite)
	colorSpread := clamp((elapsed-0.75)/2.8, 0, 1)
	if colorSpread > 0 {
		drawPixelMatrixSpread(screen, g.revealPixels, g.skeletonPixels, artRect, colorSpread)
	}
	blackAlpha := 1 - clamp((elapsed-1.05)/1.7, 0, 1)
	if blackAlpha > 0 {
		drawPixelMatrix(screen, g.skeletonPixels, artRect, blackAlpha)
	}

	for _, s := range g.sparkles {
		t := elapsed - s.delay
		if t < 0 || t > 1.2 {
			continue
		}
		alpha := uint8(255 * (1 - math.Min(1, t/1.2)))
		vector.DrawFilledCircle(screen, float32(s.x), float32(s.y-18*t), s.size, color.RGBA{255, 247, 192, alpha}, false)
	}

	drawButton(screen, g.layout.retryButton, "retry puzzle")
	backLabel := "cool levels"
	if g.editorPreview {
		backLabel = "back to editor"
	} else if g.communityPreview {
		backLabel = "back"
	}
	drawButton(screen, g.layout.revealLevelsButton, backLabel)
}

func drawButton(screen *ebiten.Image, r rect, label string) {
	registerButtonRect(r)
	drawRounded(screen, rect{x: r.x + 3, y: r.y + 4, w: r.w, h: r.h}, 8, color.RGBA{147, 137, 122, 130})
	drawRounded(screen, r, 8, colPanelDark)
	drawRounded(screen, inset(r, 4), 6, color.RGBA{237, 228, 208, 255})
	drawCenteredText(screen, label, r, colInk)
}

func drawNoticePopup(screen *ebiten.Image, message string, y float64) {
	if message == "" {
		return
	}
	const maxChars = 52
	lines := wrapTextLines(message, maxChars, 3)
	width := 220.0
	for _, line := range lines {
		lineWidth := float64(text.BoundString(face, line).Dx() + 48)
		if lineWidth > width {
			width = lineWidth
		}
	}
	if width > 500 {
		width = 500
	}
	height := float64(30 + len(lines)*18)
	r := rect{x: (ScreenWidth - width) / 2, y: y - height + 38, w: width, h: height}
	drawRounded(screen, rect{x: r.x + 4, y: r.y + 5, w: r.w, h: r.h}, 8, color.RGBA{80, 72, 62, 110})
	drawRounded(screen, r, 8, colWhite)
	drawRectOutline(screen, r, 2, colAccent)
	textY := int(r.y) + 24
	for _, line := range lines {
		drawCenteredText(screen, line, rect{x: r.x + 16, y: float64(textY - 16), w: r.w - 32, h: 18}, colAccent)
		textY += 18
	}
}

func levelTileRect(index int) rect {
	const cols = 4
	size := 84.0
	gap := 14.0
	startX := 78.0
	startY := 250.0
	col := float64(index % cols)
	row := float64(index / cols)
	return rect{x: startX + col*(size+gap), y: startY + row*(size+gap), w: size, h: size}
}

func drawLevelTile(screen *ebiten.Image, r rect, index int) {
	drawRounded(screen, rect{x: r.x + 4, y: r.y + 5, w: r.w, h: r.h}, 6, color.RGBA{126, 118, 105, 150})
	drawRounded(screen, r, 6, colGridHeavy)
	drawRounded(screen, inset(r, 5), 4, colPanel)
	board := rect{x: r.x + 24, y: r.y + 26, w: r.w - 48, h: r.h - 50}
	vector.DrawFilledRect(screen, float32(board.x), float32(board.y), float32(board.w), float32(board.h), colWhite, false)
	for i := 0; i <= 4; i++ {
		x := float32(board.x + float64(i)*board.w/4)
		y := float32(board.y + float64(i)*board.h/4)
		vector.StrokeLine(screen, x, float32(board.y), x, float32(board.y+board.h), 1, colGrid, false)
		vector.StrokeLine(screen, float32(board.x), y, float32(board.x+board.w), y, 1, colGrid, false)
	}
	if index < len(gameLevels) && gameLevels[index].Available {
		drawCenteredText(screen, gameLevels[index].Label, rect{x: r.x, y: r.y + r.h - 21, w: r.w, h: 16}, colInk)
		return
	}
	drawCenteredText(screen, "LW", rect{x: r.x, y: r.y + r.h - 21, w: r.w, h: 16}, colMuted)
}

func (g *Game) drawLevelTile(screen *ebiten.Image, r rect, index int) {
	drawLevelTile(screen, r, index)
	if index >= len(gameLevels) {
		return
	}
	level := gameLevels[index]
	if !level.Available {
		return
	}
	registerButtonRect(r)
	if best := g.bestTimes[level.ID]; best > 0 {
		board := rect{x: r.x + 18, y: r.y + 25, w: r.w - 36, h: r.h - 46}
		if thumb := g.levelThumbs[level.ID]; len(thumb) > 0 {
			vector.DrawFilledRect(screen, float32(board.x), float32(board.y), float32(board.w), float32(board.h), colWhite, false)
			drawPixelMatrix(screen, thumb, board, 1)
		}
		drawCenteredText(screen, formatTimer(best), rect{x: r.x, y: r.y + 5, w: r.w, h: 14}, colGreen)
	}
}

func drawIconButton(screen *ebiten.Image, r rect) {
	registerButtonRect(r)
	drawRounded(screen, rect{x: r.x + 3, y: r.y + 4, w: r.w, h: r.h}, 8, color.RGBA{147, 137, 122, 130})
	drawRounded(screen, r, 8, colPanelDark)
	drawRounded(screen, inset(r, 5), 6, color.RGBA{237, 228, 208, 255})
}

func drawHomeIcon(dst *ebiten.Image, r rect, c color.Color) {
	cx := float32(r.x + r.w/2)
	top := float32(r.y + 14)
	left := float32(r.x + 20)
	right := float32(r.x + r.w - 20)
	base := float32(r.y + r.h - 15)
	vector.StrokeLine(dst, left, top+15, cx, top, 3, c, false)
	vector.StrokeLine(dst, cx, top, right, top+15, 3, c, false)
	vector.StrokeLine(dst, left+5, top+16, left+5, base, 3, c, false)
	vector.StrokeLine(dst, right-5, top+16, right-5, base, 3, c, false)
	vector.StrokeLine(dst, left+5, base, right-5, base, 3, c, false)
}

func drawGearIcon(dst *ebiten.Image, r rect, c color.Color) {
	cx := float32(r.x + r.w/2)
	cy := float32(r.y + r.h/2)
	vector.StrokeCircle(dst, cx, cy, 10, 3, c, false)
	vector.StrokeCircle(dst, cx, cy, 3, 3, c, false)
	for i := 0; i < 8; i++ {
		a := float64(i) * math.Pi / 4
		x1 := cx + float32(math.Cos(a))*13
		y1 := cy + float32(math.Sin(a))*13
		x2 := cx + float32(math.Cos(a))*17
		y2 := cy + float32(math.Sin(a))*17
		vector.StrokeLine(dst, x1, y1, x2, y2, 3, c, false)
	}
}

func drawPencilIcon(dst *ebiten.Image, r rect, active bool) {
	ink := color.Color(colInk)
	if active {
		ink = colWhite
	}

	x1 := float32(r.x + 17)
	y1 := float32(r.y + r.h - 17)
	x2 := float32(r.x + r.w - 16)
	y2 := float32(r.y + 19)
	offsetX := float32(6)
	offsetY := float32(5)

	vector.StrokeLine(dst, x1, y1, x2, y2, 3, ink, false)
	vector.StrokeLine(dst, x1+offsetX, y1+offsetY, x2+offsetX, y2+offsetY, 3, ink, false)
	vector.StrokeLine(dst, x1, y1, x1+offsetX, y1+offsetY, 3, ink, false)
	vector.StrokeLine(dst, x2, y2, x2+offsetX, y2+offsetY, 3, ink, false)
	vector.StrokeLine(dst, x2+offsetX, y2+offsetY, x2+10, y2+1, 3, ink, false)
	vector.StrokeLine(dst, x2, y2, x2+10, y2+1, 3, ink, false)
	vector.StrokeLine(dst, x1-4, y1+5, x1+5, y1+13, 3, ink, false)
}

func drawEraserIcon(dst *ebiten.Image, r rect, active bool) {
	ink := color.Color(colInk)
	if active {
		ink = colWhite
	}
	x := float32(r.x + 15)
	y := float32(r.y + 23)
	w := float32(28)
	h := float32(20)
	slant := float32(7)
	vector.StrokeLine(dst, x+slant, y, x+w, y, 3, ink, false)
	vector.StrokeLine(dst, x+w, y, x+w-slant, y+h, 3, ink, false)
	vector.StrokeLine(dst, x+w-slant, y+h, x, y+h, 3, ink, false)
	vector.StrokeLine(dst, x, y+h, x+slant, y, 3, ink, false)
	vector.StrokeLine(dst, x+w-9, y+3, x+w-14, y+h-3, 3, ink, false)
	vector.StrokeLine(dst, x+4, y+h+8, x+w-5, y+h+8, 3, ink, false)
}

func (g *Game) drawMainMenu(screen *ebiten.Image) {
	drawMenuBackdrop(screen)
	drawScaledTextCentered(screen, "COMMUNITY NONGRAMS", rect{x: 76, y: 46, w: 388, h: 52}, 2.25, colInk)
	drawButton(screen, mainLevelButton(), "Cool Levels")
	drawGlobalCommunityButton(screen)
	drawButton(screen, mainTipsButton(), "Tips")
	drawButton(screen, mainSettingsButton(), "Settings")
	if time.Now().Before(g.menuNoticeUntil) {
		drawNoticePopup(screen, g.menuNotice, 542)
	}
}

func drawGlobalCommunityButton(screen *ebiten.Image) {
	r := mainCommunityButton()
	registerButtonRect(r)
	drawRounded(screen, rect{x: r.x + 5, y: r.y + 6, w: r.w, h: r.h}, 6, color.RGBA{176, 166, 151, 140})
	drawRounded(screen, r, 6, color.RGBA{14, 18, 27, 255})
	for i := 0; i < 22; i++ {
		x := int(r.x) + 8 + (i*47)%int(r.w-16)
		y := int(r.y) + 6 + (i*19)%int(r.h-12)
		s := float32(1)
		if i%7 == 0 {
			s = 2
		}
		vector.DrawFilledRect(screen, float32(x), float32(y), s, s, colWhite, false)
	}
	drawPixelPlanet(screen, int(r.x+31), int(r.y+23))
	drawPixelAstronaut(screen, int(r.x+r.w-35), int(r.y+23))
	drawCenteredText(screen, "Global Community", rect{x: r.x + 62, y: r.y, w: r.w - 124, h: r.h}, colWhite)
	drawRectOutline(screen, r, 2, colGridHeavy)
}

func registerButtonRect(r rect) {
	renderedButtonRects = append(renderedButtonRects, r)
}

// Pixelized spherical noise adapted from Deep-Fold/PixelPlanets (MIT).
func drawPixelPlanet(screen *ebiten.Image, centerX, centerY int) {
	const cell = 3
	for y := -15; y <= 15; y += cell {
		for x := -15; x <= 15; x += cell {
			if x*x+y*y > 225 {
				continue
			}
			n := perlin2D(float64(x+22)/8, float64(y+31)/8)
			c := color.RGBA{57, 123, 132, 255}
			if n > 0.05 {
				c = color.RGBA{111, 151, 91, 255}
			}
			if x+y > 10 {
				c = mixCommunityColor(c, color.RGBA{24, 35, 49, 255}, 0.48)
			}
			vector.DrawFilledRect(screen, float32(centerX+x), float32(centerY+y), cell, cell, c, false)
		}
	}
}

func drawPixelAstronaut(screen *ebiten.Image, x, y int) {
	white := color.RGBA{239, 235, 220, 255}
	visor := color.RGBA{86, 115, 134, 255}
	shadow := color.RGBA{151, 83, 71, 255}
	vector.DrawFilledRect(screen, float32(x-7), float32(y-13), 14, 10, white, false)
	vector.DrawFilledRect(screen, float32(x-4), float32(y-10), 8, 5, visor, false)
	vector.DrawFilledRect(screen, float32(x-6), float32(y-3), 12, 12, white, false)
	vector.DrawFilledRect(screen, float32(x-3), float32(y), 6, 4, shadow, false)
	vector.DrawFilledRect(screen, float32(x-10), float32(y-1), 4, 9, white, false)
	vector.DrawFilledRect(screen, float32(x+6), float32(y-1), 4, 9, white, false)
	vector.DrawFilledRect(screen, float32(x-6), float32(y+9), 4, 7, white, false)
	vector.DrawFilledRect(screen, float32(x+2), float32(y+9), 4, 7, white, false)
}

func mainLevelButton() rect {
	return rect{x: 128, y: 284, w: 284, h: 46}
}

func mainCommunityButton() rect {
	return rect{x: 128, y: 354, w: 284, h: 46}
}

func mainTipsButton() rect {
	return rect{x: 128, y: 424, w: 284, h: 46}
}

func mainSettingsButton() rect {
	return rect{x: 128, y: 494, w: 284, h: 46}
}

func tipsBackButton() rect {
	return rect{x: 202, y: 674, w: 136, h: 42}
}

func tipsPrevButton() rect {
	return rect{x: 82, y: 618, w: 120, h: 40}
}

func tipsNextButton() rect {
	return rect{x: 338, y: 618, w: 120, h: 40}
}

func tipsDemoSquare() rect {
	return rect{x: 286, y: 344, w: 132, h: 132}
}

func tipsFillToolButton() rect {
	return rect{x: 314, y: 278, w: 46, h: 46}
}

func tipsMarkToolButton() rect {
	return rect{x: 372, y: 278, w: 46, h: 46}
}

func (g *Game) drawTips(screen *ebiten.Image) {
	drawMenuBackdrop(screen)
	drawScaledTextCentered(screen, "HOW TO PLAY", rect{x: 76, y: 46, w: 388, h: 52}, 2.25, colInk)
	panel := rect{x: 70, y: 222, w: 400, h: 356}
	drawRounded(screen, rect{x: panel.x + 8, y: panel.y + 9, w: panel.w, h: panel.h}, 6, color.RGBA{126, 118, 105, 130})
	drawRounded(screen, panel, 4, colGridHeavy)
	drawRounded(screen, inset(panel, 5), 3, colPanel)

	switch g.tipsPage {
	case 0:
		g.drawTipsControls(screen, panel)
	case 1:
		g.drawTipsSolveDemo(screen, panel)
	case 2:
		g.drawTipsComplete(screen, panel)
	default:
		g.drawTipsCommunity(screen, panel)
	}

	drawCenteredText(screen, fmt.Sprintf("%d/%d", g.tipsPage+1, tipsPageCount), rect{x: 0, y: 628, w: ScreenWidth, h: 22}, colMuted)
	if g.tipsPage > 0 {
		drawButton(screen, tipsPrevButton(), "prev")
	}
	if g.tipsPage < tipsPageCount-1 {
		drawButton(screen, tipsNextButton(), "next")
	}
	drawButton(screen, tipsBackButton(), "back")
}

func (g *Game) drawTipsSolveDemo(screen *ebiten.Image, panel rect) {
	drawCenteredText(screen, "WATCH THE CLUES", rect{x: panel.x, y: panel.y + 24, w: panel.w, h: 24}, colInk)
	drawText(screen, "Numbers say which cells get filled.", 100, 280, colInk)
	drawText(screen, "Black cells and X marks show a solve.", 100, 306, colMuted)
	progress := math.Mod(float64(time.Now().UnixMilli()), 6500) / 6500
	g.drawTipsLionBoard(screen, rect{x: 194, y: 396, w: 150, h: 150}, progress, false)
	drawCenteredText(screen, "mark blanks, then fill clue groups", rect{x: 96, y: 552, w: 348, h: 24}, colAccent)
}

func (g *Game) drawTipsComplete(screen *ebiten.Image, panel rect) {
	drawCenteredText(screen, "COMPLETE", rect{x: panel.x, y: panel.y + 24, w: panel.w, h: 24}, colAccent)
	revealProgress := math.Mod(float64(time.Now().UnixMilli()), 4200) / 4200
	portrait := rect{x: panel.x + (panel.w-148)/2, y: 306, w: 148, h: 148}
	drawRounded(screen, rect{x: portrait.x - 6, y: portrait.y - 6, w: portrait.w + 12, h: portrait.h + 12}, 4, colGridHeavy)
	drawRounded(screen, portrait, 3, colWhite)
	if lion := g.levelThumbs["l4"]; len(lion) > 0 {
		if puzzle := g.levelPuzzle["l4"]; puzzle != nil {
			seeds := pixelsFromRaw(puzzle.SkeletonRaw)
			if len(seeds) == 0 {
				seeds = tipsSilhouetteSeeds(puzzle.Solution, lion)
			}
			colorSpread := clamp((revealProgress-0.18)/0.72, 0, 1)
			if colorSpread > 0 {
				drawPixelMatrixSpread(screen, lion, seeds, portrait, colorSpread)
			}
			blackAlpha := 1 - clamp((revealProgress-0.36)/0.42, 0, 1)
			if blackAlpha > 0 {
				drawPixelMatrix(screen, seeds, portrait, blackAlpha)
			}
		} else {
			drawPixelMatrixSpread(screen, lion, lion, portrait, clamp((revealProgress-0.18)/0.72, 0, 1))
		}
	}
	drawCenteredText(screen, "When the board matches the clues,", rect{x: panel.x + 28, y: 480, w: panel.w - 56, h: 24}, colInk)
	drawCenteredText(screen, "the black solve reveals color.", rect{x: panel.x + 28, y: 508, w: panel.w - 56, h: 24}, colInk)
	drawCenteredText(screen, "This one becomes the lion.", rect{x: panel.x + 28, y: 536, w: panel.w - 56, h: 24}, colAccent)
}

func tipsSilhouetteSeeds(solution [][]bool, reveal [][]assets.PixelCell) [][]assets.PixelCell {
	if len(solution) == 0 || len(reveal) == 0 {
		return nil
	}
	rows := min(len(solution), len(reveal))
	seeds := make([][]assets.PixelCell, rows)
	for y := 0; y < rows; y++ {
		cols := min(len(solution[y]), len(reveal[y]))
		seeds[y] = make([]assets.PixelCell, cols)
		for x := 0; x < cols; x++ {
			if solution[y][x] {
				seeds[y][x] = assets.PixelCell{Visible: true, Color: colInk}
			}
		}
	}
	return seeds
}

func (g *Game) drawTipsLionBoard(screen *ebiten.Image, board rect, progress float64, complete bool) {
	matrix := g.levelThumbs["l4"]
	puzzle := g.levelPuzzle["l4"]
	if puzzle == nil || len(puzzle.Solution) == 0 || len(puzzle.Solution[0]) == 0 || len(matrix) == 0 || len(matrix[0]) == 0 {
		drawTipsFallbackBoard(screen, board)
		return
	}
	solution := puzzle.Solution
	rows := len(solution)
	cols := len(solution[0])

	cellSize := math.Floor(math.Min(board.w/float64(cols), board.h/float64(rows)))
	board.w = cellSize * float64(cols)
	board.h = cellSize * float64(rows)
	rowClues := nonogram.RowClues(solution)
	colClues := nonogram.ColumnClues(solution)
	states := make([][]nonogram.CellState, rows)
	for y := range states {
		states[y] = make([]nonogram.CellState, cols)
	}
	moves := tipsLionSolveMoves(solution)
	moveTarget := int(math.Round(clamp(progress, 0, 1) * float64(len(moves))))
	if complete {
		moveTarget = len(moves)
	}
	if moveTarget > len(moves) {
		moveTarget = len(moves)
	}
	for i := 0; i < moveTarget; i++ {
		move := moves[i]
		if move.y >= 0 && move.y < rows && move.x >= 0 && move.x < cols {
			states[move.y][move.x] = move.state
		}
	}
	activeRow := -1
	activeCol := -1
	if !complete && len(moves) > 0 {
		activeMove := moves[min(moveTarget, len(moves)-1)]
		activeRow = activeMove.y
		activeCol = activeMove.x
	}
	if activeRow < 0 || activeRow >= rows {
		activeRow = rows - 1
		activeCol = cols - 1
	}

	clueLeft := board.x - 64
	clueTop := board.y - 62
	drawRounded(screen, rect{x: clueLeft - 4, y: clueTop - 4, w: board.w + 72, h: board.h + 72}, 4, colWhite)
	for y := 0; y < rows; y++ {
		c := color.RGBA{238, 235, 226, 255}
		if y == activeRow && !complete {
			c = color.RGBA{251, 213, 107, 255}
		}
		vector.DrawFilledRect(screen, float32(clueLeft), float32(board.y+float64(y)*cellSize), 58, float32(cellSize), c, false)
	}
	for x := 0; x < cols; x++ {
		c := color.RGBA{238, 235, 226, 255}
		if x == activeCol && !complete {
			c = color.RGBA{251, 213, 107, 255}
		}
		vector.DrawFilledRect(screen, float32(board.x+float64(x)*cellSize), float32(clueTop), float32(cellSize), 58, c, false)
	}

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			cell := rect{x: board.x + float64(x)*cellSize, y: board.y + float64(y)*cellSize, w: cellSize, h: cellSize}
			cellColor := colWhite
			cellIndex := y*cols + x
			if complete && solution[y][x] && cellIndex < int(float64(rows*cols)*clamp(progress*1.25, 0, 1)) {
				if y < len(matrix) && x < len(matrix[y]) && matrix[y][x].Visible {
					cellColor = matrix[y][x].Color
				} else {
					cellColor = colInk
				}
			} else if states[y][x] == nonogram.CellFilled {
				cellColor = colInk
			}
			vector.DrawFilledRect(screen, float32(cell.x), float32(cell.y), float32(cell.w), float32(cell.h), cellColor, false)
			if states[y][x] == nonogram.CellMarked && !complete {
				drawCellX(screen, cell, colAccent)
			}
			if !complete && y == activeRow && x == activeCol {
				drawRectOutline(screen, inset(cell, 2), 2, colAccent)
			}
			drawRectOutline(screen, cell, 1, colGrid)
		}
	}
	drawRectOutline(screen, board, 3, colGridHeavy)

	for y, clues := range rowClues {
		label := clueLabel(clues)
		tx := int(board.x-8) - text.BoundString(face, label).Dx()
		ty := int(board.y + float64(y)*cellSize + cellSize/2 + 5)
		drawText(screen, label, tx, ty, colInk)
	}
	for x, clues := range colClues {
		parts := make([]string, len(clues))
		for i, n := range clues {
			parts[i] = fmt.Sprint(n)
		}
		cx := int(board.x + float64(x)*cellSize + cellSize/2)
		step := columnClueStep(len(parts))
		bottomY := int(board.y - 9)
		startY := bottomY - (len(parts)-1)*step
		for i, part := range parts {
			drawText(screen, part, cx-text.BoundString(face, part).Dx()/2, startY+i*step, colInk)
		}
	}
}

type tipsSolveMove struct {
	x     int
	y     int
	state nonogram.CellState
}

func tipsLionSolveMoves(solution [][]bool) []tipsSolveMove {
	moves := make([]tipsSolveMove, 0, len(solution)*len(solution[0]))
	if len(solution) < 10 || len(solution[0]) < 10 {
		return tipsGenericSolveMoves(solution)
	}
	addMarkedRow := func(y int) {
		for x := range solution[y] {
			moves = append(moves, tipsSolveMove{x: x, y: y, state: nonogram.CellMarked})
		}
	}
	addFilledRun := func(y, start, end int) {
		for x := start; x <= end && x < len(solution[y]); x++ {
			if x >= 0 && solution[y][x] {
				moves = append(moves, tipsSolveMove{x: x, y: y, state: nonogram.CellFilled})
			}
		}
	}
	addMarks := func(y int, xs ...int) {
		for _, x := range xs {
			if y >= 0 && y < len(solution) && x >= 0 && x < len(solution[y]) && !solution[y][x] {
				moves = append(moves, tipsSolveMove{x: x, y: y, state: nonogram.CellMarked})
			}
		}
	}

	addMarkedRow(0)
	addMarkedRow(1)
	addFilledRun(2, 0, 5)
	addMarks(2, 6, 9)
	addFilledRun(2, 7, 8)
	addFilledRun(3, 0, 5)
	addMarks(3, 6, 7, 8)
	addFilledRun(3, 9, 9)
	addFilledRun(4, 0, 6)
	addMarks(4, 7, 9)
	addFilledRun(4, 8, 8)
	addFilledRun(5, 0, 7)
	addMarks(5, 8, 9)
	addFilledRun(6, 0, 7)
	addMarks(6, 8, 9)
	addMarks(7, 0, 8, 9)
	addFilledRun(7, 1, 7)
	addMarks(8, 0, 2, 4, 6, 8, 9)
	addFilledRun(8, 1, 1)
	addFilledRun(8, 3, 3)
	addFilledRun(8, 5, 5)
	addFilledRun(8, 7, 7)
	addMarks(9, 0, 2, 4, 6, 8, 9)
	addFilledRun(9, 1, 1)
	addFilledRun(9, 3, 3)
	addFilledRun(9, 5, 5)
	addFilledRun(9, 7, 7)
	return moves
}

func tipsGenericSolveMoves(solution [][]bool) []tipsSolveMove {
	moves := make([]tipsSolveMove, 0)
	for y, row := range solution {
		for x, filled := range row {
			if filled {
				moves = append(moves, tipsSolveMove{x: x, y: y, state: nonogram.CellFilled})
			} else {
				moves = append(moves, tipsSolveMove{x: x, y: y, state: nonogram.CellMarked})
			}
		}
	}
	return moves
}

func drawTipsFallbackBoard(screen *ebiten.Image, board rect) {
	drawRounded(screen, board, 4, colWhite)
	for i := 0; i <= 5; i++ {
		x := float32(board.x + float64(i)*board.w/5)
		y := float32(board.y + float64(i)*board.h/5)
		vector.StrokeLine(screen, x, float32(board.y), x, float32(board.y+board.h), 1, colGrid, false)
		vector.StrokeLine(screen, float32(board.x), y, float32(board.x+board.w), y, 1, colGrid, false)
	}
	drawRectOutline(screen, board, 3, colGridHeavy)
}

func (g *Game) drawTipsControls(screen *ebiten.Image, panel rect) {
	drawCenteredText(screen, "FILL SQUARES", rect{x: panel.x, y: panel.y + 24, w: panel.w, h: 24}, colInk)
	drawText(screen, "Switch tools:", 96, 292, colInk)
	g.drawTipsToolButtons(screen)
	drawText(screen, "Mouse controls:", 96, 356, colMuted)
	drawText(screen, "left click fills", 96, 386, colAccent)
	drawText(screen, "right click marks X", 96, 416, colAccent)
	drawTipsDemoSquare(screen, tipsDemoSquare(), g.tipsDemoCells)
	drawCenteredText(screen, "try the grid", rect{x: 274, y: 490, w: 156, h: 24}, colMuted)
	drawText(screen, "Settings has assist.", 100, 516, colInk)
	drawText(screen, "It can auto-correct mistakes", 100, 542, colInk)
	drawText(screen, "whenever you want for +10s.", 100, 568, colInk)
}

func (g *Game) drawTipsCommunity(screen *ebiten.Image, panel rect) {
	drawCenteredText(screen, "COMMUNITY FIRST", rect{x: panel.x, y: panel.y + 24, w: panel.w, h: 24}, colInk)
	drawText(screen, "Browse Gallery for art and packs.", 96, 292, colInk)
	drawText(screen, "Open packs to play a set of puzzles.", 96, 326, colInk)
	drawText(screen, "Use chat and likes to react.", 96, 360, colInk)
	drawText(screen, "Publish from My Library when your", 96, 414, colAccent)
	drawText(screen, "own art or pack is ready.", 96, 442, colAccent)
	drawText(screen, "Cool Levels are built-in solo puzzles.", 96, 508, colMuted)
}

func (g *Game) drawTipsToolButtons(screen *ebiten.Image) {
	drawTrigger(screen, tipsFillToolButton(), g.tipsDemoTool == nonogram.ToolFill, colBlue, g.icons.Pencil)
	drawTrigger(screen, tipsMarkToolButton(), g.tipsDemoTool == nonogram.ToolMark, colAccent, g.icons.X)
}

func drawTipsDemoSquare(screen *ebiten.Image, r rect, states [16]nonogram.CellState) {
	registerButtonRect(r)
	drawRounded(screen, rect{x: r.x + 5, y: r.y + 6, w: r.w, h: r.h}, 6, color.RGBA{126, 118, 105, 130})
	drawRounded(screen, r, 4, colGridHeavy)
	grid := inset(r, 8)
	drawRounded(screen, grid, 2, colWhite)
	const cells = 4
	cellSize := grid.w / cells
	for y := 0; y < cells; y++ {
		for x := 0; x < cells; x++ {
			cell := rect{x: grid.x + float64(x)*cellSize, y: grid.y + float64(y)*cellSize, w: cellSize, h: cellSize}
			if (x+y)%2 == 0 {
				vector.DrawFilledRect(screen, float32(cell.x), float32(cell.y), float32(cell.w), float32(cell.h), color.RGBA{246, 241, 228, 255}, false)
			}
			switch states[y*cells+x] {
			case nonogram.CellFilled:
				drawRounded(screen, inset(cell, 4), 2, colInk)
			case nonogram.CellMarked:
				drawCellX(screen, cell, colAccent)
			}
			drawRectOutline(screen, cell, 1, colGrid)
		}
	}
	drawRectOutline(screen, grid, 2, colGridHeavy)
}

func tipsDemoCellAt(px, py int) (int, bool) {
	r := inset(tipsDemoSquare(), 8)
	if !r.Contains(px, py) {
		return 0, false
	}
	const cells = 4
	cellSize := r.w / cells
	x := int((float64(px) - r.x) / cellSize)
	y := int((float64(py) - r.y) / cellSize)
	if x < 0 || y < 0 || x >= cells || y >= cells {
		return 0, false
	}
	return y*cells + x, true
}

func drawCellX(screen *ebiten.Image, cell rect, c color.Color) {
	pad := math.Max(6, cell.w*0.22)
	x1 := float32(cell.x + pad)
	y1 := float32(cell.y + pad)
	x2 := float32(cell.x + cell.w - pad)
	y2 := float32(cell.y + cell.h - pad)
	vector.StrokeLine(screen, x1, y1, x2, y2, 4, c, false)
	vector.StrokeLine(screen, x2, y1, x1, y2, 4, c, false)
}

func (g *Game) drawLevelSelect(screen *ebiten.Image) {
	drawMenuBackdrop(screen)
	drawScaledTextCentered(screen, "COOL LEVELS", rect{x: 56, y: 42, w: 428, h: 54}, 2.1, colInk)
	drawCenteredText(screen, "These are just some built-in nongrams I made.", rect{x: 56, y: 202, w: 428, h: 20}, colMuted)
	drawCenteredText(screen, "I wanted the focus to be the multiplayer though.", rect{x: 56, y: 226, w: 428, h: 20}, colMuted)
	pageStart := g.levelPage * levelSelectPageSize
	for slot := 0; slot < levelSelectPageSize; slot++ {
		g.drawLevelTile(screen, levelTileRect(slot), pageStart+slot)
	}
	if time.Now().Before(g.menuNoticeUntil) {
		drawNoticePopup(screen, g.menuNotice, 636)
	}
	drawCenteredText(screen, fmt.Sprintf("%d/%d", g.levelPage+1, levelSelectPages()), rect{x: 0, y: 650, w: ScreenWidth, h: 26}, colMuted)
	drawButton(screen, g.layout.levelPrevButton, "prev")
	drawButton(screen, g.layout.levelNextButton, "next")
	drawButton(screen, g.layout.levelBackButton, "back")
}

func drawMenuBackdrop(screen *ebiten.Image) {
	screen.Fill(colPanel)
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 186, color.RGBA{214, 205, 189, 255}, false)
	vector.DrawFilledRect(screen, 0, 176, ScreenWidth, 12, colGridHeavy, false)
	for i := 0; i < 10; i++ {
		x := float32(i*64 - 44)
		vector.StrokeLine(screen, x, 0, x+108, 186, 3, color.RGBA{244, 239, 224, 155}, false)
		vector.StrokeLine(screen, x+38, 0, x-70, 186, 3, color.RGBA{176, 166, 151, 105}, false)
	}
	drawRounded(screen, rect{x: 62, y: 32, w: 416, h: 92}, 8, color.RGBA{45, 45, 43, 255})
	drawRounded(screen, rect{x: 76, y: 46, w: 388, h: 52}, 6, colWhite)
}

func (g *Game) drawSettings(screen *ebiten.Image) {
	drawMenuBackdrop(screen)
	drawScaledTextCentered(screen, "SETTINGS", rect{x: 76, y: 46, w: 388, h: 52}, 2.35, colInk)
	panel := rect{x: 106, y: 246, w: 328, h: 312}
	drawRounded(screen, rect{x: panel.x + 5, y: panel.y + 6, w: panel.w, h: panel.h}, 14, color.RGBA{70, 65, 58, 145})
	drawRounded(screen, panel, 14, colPanel)
	drawRectOutline(screen, inset(panel, 10), 3, color.RGBA{98, 92, 84, 255})
	drawCenteredText(screen, "Settings", rect{x: panel.x, y: panel.y + 28, w: panel.w, h: 28}, colInk)
	vector.StrokeLine(screen, float32(panel.x+42), float32(panel.y+72), float32(panel.x+panel.w-42), float32(panel.y+72), 2, colGrid, false)
	drawButton(screen, g.layout.soundButton, toggleLabel("sound", g.audioEnabled))
	drawRectOutline(screen, g.layout.soundButton, 2, color.RGBA{98, 92, 84, 255})
	drawButton(screen, g.layout.autoCorrectButton, autoCorrectLabel(g.autoCorrect))
	drawRectOutline(screen, g.layout.autoCorrectButton, 2, color.RGBA{98, 92, 84, 255})
	drawCenteredText(screen, "stops wrong fills", rect{x: panel.x + 44, y: panel.y + 194, w: panel.w - 88, h: 20}, colMuted)
	drawCenteredText(screen, "mistakes add +10s", rect{x: panel.x + 44, y: panel.y + 216, w: panel.w - 88, h: 20}, colMuted)
	drawButton(screen, g.layout.settingsCloseButton, "back")
	drawRectOutline(screen, g.layout.settingsCloseButton, 2, color.RGBA{98, 92, 84, 255})
}

func drawRounded(dst *ebiten.Image, r rect, radius float32, c color.Color) {
	x, y, w, h := float32(r.x), float32(r.y), float32(r.w), float32(r.h)
	vector.DrawFilledRect(dst, x+radius, y, w-2*radius, h, c, false)
	vector.DrawFilledRect(dst, x, y+radius, w, h-2*radius, c, false)
	vector.DrawFilledCircle(dst, x+radius, y+radius, radius, c, false)
	vector.DrawFilledCircle(dst, x+w-radius, y+radius, radius, c, false)
	vector.DrawFilledCircle(dst, x+radius, y+h-radius, radius, c, false)
	vector.DrawFilledCircle(dst, x+w-radius, y+h-radius, radius, c, false)
}

func drawX(dst *ebiten.Image, r rect, c color.Color) {
	vector.StrokeLine(dst, float32(r.x), float32(r.y), float32(r.x+r.w), float32(r.y+r.h), 3, c, false)
	vector.StrokeLine(dst, float32(r.x+r.w), float32(r.y), float32(r.x), float32(r.y+r.h), 3, c, false)
}

func drawRectOutline(dst *ebiten.Image, r rect, thickness float32, c color.Color) {
	x := float32(r.x)
	y := float32(r.y)
	w := float32(r.w)
	h := float32(r.h)
	vector.StrokeLine(dst, x, y, x+w, y, thickness, c, false)
	vector.StrokeLine(dst, x+w, y, x+w, y+h, thickness, c, false)
	vector.StrokeLine(dst, x+w, y+h, x, y+h, thickness, c, false)
	vector.StrokeLine(dst, x, y+h, x, y, thickness, c, false)
}

func drawText(dst *ebiten.Image, s string, x, y int, c color.Color) {
	text.Draw(dst, s, face, x, y, c)
}

func drawScaledText(dst *ebiten.Image, s string, x, y int, scale float64, c color.Color) {
	b := text.BoundString(face, s)
	img := ebiten.NewImage(b.Dx()+8, b.Dy()+8)
	text.Draw(img, s, face, 4-b.Min.X, 4-b.Min.Y, c)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(img, op)
}

func drawScaledTextCentered(dst *ebiten.Image, s string, r rect, scale float64, c color.Color) {
	b := text.BoundString(face, s)
	w := float64(b.Dx()+8) * scale
	h := float64(b.Dy()+8) * scale
	x := int(r.x + (r.w-w)/2)
	y := int(r.y + (r.h-h)/2)
	drawScaledText(dst, s, x, y, scale, c)
}

func drawCenteredText(dst *ebiten.Image, s string, r rect, c color.Color) {
	b := text.BoundString(face, s)
	x := int(r.x + r.w/2 - float64(b.Dx())/2)
	y := int(r.y + r.h/2 + float64(b.Dy())/2 - 2)
	drawText(dst, s, x, y, c)
}

func drawImageFit(dst *ebiten.Image, img *ebiten.Image, r rect, alpha float64) {
	b := imageBounds(img)
	scale := math.Min(r.w/float64(b.Dx()), r.h/float64(b.Dy()))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(r.x+(r.w-float64(b.Dx())*scale)/2, r.y+(r.h-float64(b.Dy())*scale)/2)
	op.ColorScale.ScaleAlpha(float32(alpha))
	dst.DrawImage(img, op)
}

func drawPixelMatrix(dst *ebiten.Image, matrix [][]assets.PixelCell, r rect, alpha float64) {
	if len(matrix) == 0 || len(matrix[0]) == 0 || alpha <= 0 {
		return
	}
	rows := len(matrix)
	cols := len(matrix[0])
	cellSize := math.Floor(math.Min(r.w/float64(cols), r.h/float64(rows)))
	if cellSize < 1 {
		cellSize = 1
	}
	totalW := cellSize * float64(cols)
	totalH := cellSize * float64(rows)
	startX := math.Floor(r.x + (r.w-totalW)/2)
	startY := math.Floor(r.y + (r.h-totalH)/2)
	img := imageFromMatrix(matrix, alpha, color.RGBA{}, 0)
	drawPixelImage(dst, img, startX, startY, cellSize)
}

func drawPixelMatrixSpread(dst *ebiten.Image, reveal, seeds [][]assets.PixelCell, r rect, progress float64) {
	if len(reveal) == 0 || len(reveal[0]) == 0 || progress <= 0 {
		return
	}
	rows := len(reveal)
	cols := len(reveal[0])
	distances := make([][]int, rows)
	maxDistance := 0
	seedFound := false
	for y := 0; y < rows; y++ {
		distances[y] = make([]int, cols)
		for x := 0; x < cols; x++ {
			best := rows + cols
			for seedY, row := range seeds {
				for seedX, seed := range row {
					if !seed.Visible {
						continue
					}
					seedFound = true
					distance := absInt(x-seedX) + absInt(y-seedY)
					if distance < best {
						best = distance
					}
				}
			}
			if !seedFound {
				best = absInt(x-cols/2) + absInt(y-rows/2)
			}
			distances[y][x] = best
			if reveal[y][x].Visible && best > maxDistance {
				maxDistance = best
			}
		}
	}

	wave := easeInOut(progress) * float64(maxDistance+2)
	img := image.NewRGBA(image.Rect(0, 0, cols, rows))
	for y, row := range reveal {
		for x, cell := range row {
			if !cell.Visible {
				continue
			}
			stagger := float64((x*7+y*11)%5) * 0.12
			alpha := clamp(wave-float64(distances[y][x])-stagger, 0, 1)
			if alpha > 0 {
				img.SetRGBA(x, y, alphaColor(cell.Color, alpha))
			}
		}
	}
	cellSize := math.Floor(math.Min(r.w/float64(cols), r.h/float64(rows)))
	if cellSize < 1 {
		cellSize = 1
	}
	startX := math.Floor(r.x + (r.w-cellSize*float64(cols))/2)
	startY := math.Floor(r.y + (r.h-cellSize*float64(rows))/2)
	drawPixelImage(dst, ebiten.NewImageFromImage(img), startX, startY, cellSize)
}

func drawPixelMatrixTinted(dst *ebiten.Image, matrix [][]assets.PixelCell, r rect, alpha float64, tint color.RGBA, tintAmount float64) {
	if len(matrix) == 0 || len(matrix[0]) == 0 || alpha <= 0 {
		return
	}
	tintAmount = clamp(tintAmount, 0, 1)
	rows := len(matrix)
	cols := len(matrix[0])
	cellSize := math.Floor(math.Min(r.w/float64(cols), r.h/float64(rows)))
	if cellSize < 1 {
		cellSize = 1
	}
	totalW := cellSize * float64(cols)
	totalH := cellSize * float64(rows)
	startX := math.Floor(r.x + (r.w-totalW)/2)
	startY := math.Floor(r.y + (r.h-totalH)/2)
	img := imageFromMatrix(matrix, alpha, tint, tintAmount)
	drawPixelImage(dst, img, startX, startY, cellSize)
}

func imageFromMatrix(matrix [][]assets.PixelCell, alpha float64, tint color.RGBA, tintAmount float64) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, len(matrix[0]), len(matrix)))
	for y, row := range matrix {
		for x, cell := range row {
			if !cell.Visible {
				continue
			}
			c := cell.Color
			if tintAmount > 0 {
				c = mixColor(c, tint, tintAmount)
			}
			img.SetRGBA(x, y, alphaColor(c, alpha))
		}
	}
	return ebiten.NewImageFromImage(img)
}

func drawPixelImage(dst, img *ebiten.Image, x, y, scale float64) {
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	dst.DrawImage(img, op)
}

func mixColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: uint8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

func alphaColor(c color.RGBA, alpha float64) color.RGBA {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	c.A = uint8(float64(c.A) * alpha)
	return c
}

func drawIconImage(dst *ebiten.Image, img *ebiten.Image, r rect, alpha float64) {
	b := imageBounds(img)
	scale := math.Min(r.w/float64(b.Dx()), r.h/float64(b.Dy()))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(r.x+(r.w-float64(b.Dx())*scale)/2, r.y+(r.h-float64(b.Dy())*scale)/2)
	op.ColorScale.ScaleAlpha(float32(alpha))
	dst.DrawImage(img, op)
}

func drawImageCentered(dst *ebiten.Image, img *ebiten.Image, cx, cy int, w, h float64, alpha float64) {
	b := imageBounds(img)
	scale := math.Min(w/float64(b.Dx()), h/float64(b.Dy()))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(cx)-float64(b.Dx())*scale/2, float64(cy)-float64(b.Dy())*scale/2)
	op.ColorScale.ScaleAlpha(float32(alpha))
	dst.DrawImage(img, op)

	ebitenutil.DrawRect(dst, float64(cx)-float64(b.Dx())*scale/2-4, float64(cy)-float64(b.Dy())*scale/2-4, float64(b.Dx())*scale+8, 3, colGridHeavy)
}

func clueLabel(clues []int) string {
	parts := make([]string, len(clues))
	for i, n := range clues {
		parts[i] = fmt.Sprint(n)
	}
	return strings.Join(parts, " ")
}

func inset(r rect, amount float64) rect {
	return rect{x: r.x + amount, y: r.y + amount, w: r.w - amount*2, h: r.h - amount*2}
}

func scaleAround(r rect, scale float64) rect {
	w := r.w * scale
	h := r.h * scale
	return rect{x: r.x + (r.w-w)/2, y: r.y + (r.h-h)/2, w: w, h: h}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func easeOutBack(x float64) float64 {
	c1 := 1.70158
	c3 := c1 + 1
	return 1 + c3*math.Pow(x-1, 3) + c1*math.Pow(x-1, 2)
}

func easeInOut(x float64) float64 {
	if x < 0.5 {
		return 2 * x * x
	}
	return 1 - math.Pow(-2*x+2, 2)/2
}

func (g *Game) hoverCell() (int, int) {
	x, y := ebiten.CursorPosition()
	cellX, cellY, ok := g.layout.CellAt(x, y, g.board.Width, g.board.Height)
	if !ok {
		return -1, -1
	}
	return cellX, cellY
}

func formatTimer(d time.Duration) string {
	total := int(d.Seconds())
	minutes := total / 60
	seconds := total % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func toggleLabel(label string, enabled bool) string {
	if enabled {
		return label + ": on"
	}
	return label + ": off"
}

func autoCorrectLabel(enabled bool) string {
	if enabled {
		return "assist: on"
	}
	return "assist: off"
}

package game

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/alex/nongrampictures/internal/assets"
	"github.com/alex/nongrampictures/internal/nonogram"
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
		drawText(screen, label, tx, ty, colInk)
	}
	for x, clues := range g.colClues {
		parts := make([]string, len(clues))
		for i, n := range clues {
			parts[i] = fmt.Sprint(n)
		}
		cx := int(g.layout.boardX + float64(x)*g.layout.cellSize + g.layout.cellSize/2)
		step := columnClueStep(len(parts))
		bottomY := int(g.layout.boardY - 10)
		startY := bottomY - (len(parts)-1)*step
		for i, part := range parts {
			drawText(screen, part, cx-text.BoundString(face, part).Dx()/2, startY+i*step, colInk)
		}
	}
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
				drawX(screen, inset(cellRect, 8), color.RGBA{245, 139, 17, 255})
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
	drawTrigger(screen, g.layout.markTrigger, g.tool == nonogram.ToolMark, colAccent, g.icons.Eraser)
	drawButton(screen, g.layout.godModeButton, "GOD")
}

func drawTrigger(screen *ebiten.Image, r rect, active bool, c color.RGBA, icon *ebiten.Image) {
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

	scale := 1.0
	if elapsed < 0.45 {
		scale = 0.82 + 0.22*easeOutBack(elapsed/0.45)
	}
	artRect := rect{x: 118, y: 205, w: 330, h: 330}
	displayRect := scaleAround(artRect, scale)
	drawRounded(screen, rect{x: artRect.x - 12, y: artRect.y - 12, w: artRect.w + 24, h: artRect.h + 24}, 8, colGridHeavy)
	drawRounded(screen, rect{x: artRect.x - 6, y: artRect.y - 6, w: artRect.w + 12, h: artRect.h + 12}, 6, colWhite)
	whiteFade := easeInOut(clamp((elapsed-0.55)/1.45, 0, 1))
	drawPixelMatrixTinted(screen, g.skeletonPixels, displayRect, 1, color.RGBA{255, 255, 255, 255}, whiteFade)
	colorFade := easeInOut(clamp((elapsed-2.05)/1.15, 0, 1))
	if colorFade > 0 {
		drawPixelMatrix(screen, g.revealPixels, displayRect, colorFade)
		drawShineSweep(screen, displayRect, clamp((elapsed-2.05)/2.4, 0, 1), colorFade)
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
	drawButton(screen, g.layout.revealLevelsButton, "levels")
}

func drawShineSweep(dst *ebiten.Image, r rect, progress, revealAlpha float64) {
	if progress <= 0 || progress >= 1 || revealAlpha <= 0 {
		return
	}
	progress = easeInOut(progress)
	x0, y0 := r.x, r.y
	x1, y1 := r.x+r.w, r.y+r.h
	minS := x0 + y0
	maxS := x1 + y1
	center := minS - 90 + (maxS-minS+180)*progress
	sweepAlpha := math.Sin(progress * math.Pi)

	for i := -5; i <= 5; i++ {
		offset := float64(i) * 8
		weight := 1 - math.Abs(float64(i))/6
		if weight <= 0 {
			continue
		}
		ax, ay, bx, by, ok := clippedNegativeDiagonal(x0, y0, x1, y1, center+offset)
		if !ok {
			continue
		}
		alpha := uint8(125 * sweepAlpha * weight * clamp(revealAlpha+0.25, 0, 1))
		vector.StrokeLine(dst, float32(ax), float32(ay), float32(bx), float32(by), 5, color.RGBA{255, 252, 224, alpha}, false)
	}
}

func clippedNegativeDiagonal(x0, y0, x1, y1, sum float64) (float64, float64, float64, float64, bool) {
	points := make([][2]float64, 0, 4)
	addPoint := func(x, y float64) {
		const epsilon = 0.001
		if x < x0-epsilon || x > x1+epsilon || y < y0-epsilon || y > y1+epsilon {
			return
		}
		for _, p := range points {
			if math.Abs(p[0]-x) < epsilon && math.Abs(p[1]-y) < epsilon {
				return
			}
		}
		points = append(points, [2]float64{x, y})
	}

	addPoint(x0, sum-x0)
	addPoint(x1, sum-x1)
	addPoint(sum-y0, y0)
	addPoint(sum-y1, y1)
	if len(points) < 2 {
		return 0, 0, 0, 0, false
	}
	return points[0][0], points[0][1], points[1][0], points[1][1], true
}

func drawButton(screen *ebiten.Image, r rect, label string) {
	drawRounded(screen, rect{x: r.x + 3, y: r.y + 4, w: r.w, h: r.h}, 8, color.RGBA{147, 137, 122, 130})
	drawRounded(screen, r, 8, colPanelDark)
	drawRounded(screen, inset(r, 4), 6, color.RGBA{237, 228, 208, 255})
	drawCenteredText(screen, label, r, colInk)
}

func levelTileRect(index int) rect {
	const cols = 4
	size := 84.0
	gap := 14.0
	startX := 78.0
	startY := 206.0
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
	if index < len(gameLevels) {
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
	drawScaledTextCentered(screen, "PIXAROSS", rect{x: 76, y: 46, w: 388, h: 52}, 2.25, colInk)
	drawButton(screen, g.layout.levelSelectButton, "Level Select")
	drawButton(screen, g.layout.mainSettingsButton, "Settings")
	if time.Now().Before(g.menuNoticeUntil) {
		drawCenteredText(screen, g.menuNotice, rect{x: 0, y: 542, w: ScreenWidth, h: 36}, colAccent)
	}
}

func (g *Game) drawLevelSelect(screen *ebiten.Image) {
	drawMenuBackdrop(screen)
	drawScaledTextCentered(screen, "LEVEL SELECT", rect{x: 56, y: 48, w: 428, h: 58}, 2.35, colInk)
	pageStart := g.levelPage * levelSelectPageSize
	for slot := 0; slot < levelSelectPageSize; slot++ {
		g.drawLevelTile(screen, levelTileRect(slot), pageStart+slot)
	}
	if time.Now().Before(g.menuNoticeUntil) {
		drawCenteredText(screen, g.menuNotice, rect{x: 0, y: 648, w: ScreenWidth, h: 34}, colAccent)
	}
	drawCenteredText(screen, fmt.Sprintf("%d/%d", g.levelPage+1, levelSelectPages), rect{x: 0, y: 650, w: ScreenWidth, h: 26}, colMuted)
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
	drawText(screen, "auto on: mistakes +10s", int(panel.x+76), int(panel.y+204), colMuted)
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

	for y, row := range matrix {
		for x, cell := range row {
			if !cell.Visible {
				continue
			}
			c := alphaColor(cell.Color, alpha)
			vector.DrawFilledRect(
				dst,
				float32(startX+float64(x)*cellSize),
				float32(startY+float64(y)*cellSize),
				float32(cellSize),
				float32(cellSize),
				c,
				false,
			)
		}
	}
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

	for y, row := range matrix {
		for x, cell := range row {
			if !cell.Visible {
				continue
			}
			c := mixColor(cell.Color, tint, tintAmount)
			c = alphaColor(c, alpha)
			vector.DrawFilledRect(
				dst,
				float32(startX+float64(x)*cellSize),
				float32(startY+float64(y)*cellSize),
				float32(cellSize),
				float32(cellSize),
				c,
				false,
			)
		}
	}
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
		return "auto: on"
	}
	return "auto: off"
}

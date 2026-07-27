package game

import (
	"fmt"
	"image"
	"math"
	"time"

	"github.com/BakedSoups/community_nongrams/internal/assets"
	"github.com/BakedSoups/community_nongrams/internal/community"
	"github.com/BakedSoups/community_nongrams/internal/nonogram"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 540
	ScreenHeight = 780

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
	communityMusicOn  bool
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

	editor               editorState
	editorUndo           []editorState
	editorPointer        bool
	editorLastX          int
	editorLastY          int
	editorSizeOpen       bool
	editorPreview        bool
	communityPreview     bool
	editorOnionSkin      bool
	editorTitleEditing   bool
	editorTitleDraft     string
	currentDraftID       string
	profileArt           editorState
	profileReturn        editorState
	profileDraftID       string
	editingProfile       bool
	profileBio           string
	profileBioDraft      string
	profileBioEditing    bool
	profileSocial        string
	profileSocialDrafts  [3]string
	profileSocialEditing bool
	profileSocialSlot    int
	profilePalette       string
	profilePaletteSlot   int
	profileColor         string
	profileColorPicking  bool
	profileName          string
	profileNameDraft     string
	profileNameEditing   bool

	communityLibrary         community.Library
	communityView            communityView
	communityPage            int
	communityCatalog         []community.LevelVersion
	activeCommunityPack      string
	packSelection            map[string]bool
	communityEmail           string
	communityCreators        []community.CreatorProfile
	creatorSearch            string
	creatorSearchActive      bool
	selectedCreator          int
	communityPlayReturn      communityView
	communityGallery         []community.GalleryItem
	communityPublished       []community.GalleryItem
	communityCompleted       []community.GalleryItem
	communityChatMessages    []community.ChatMessage
	galleryKind              string
	gallerySort              string
	gallerySortOpen          bool
	selectedGallery          int
	chatKind                 string
	chatID                   string
	chatTitle                string
	chatDraft                string
	chatReturn               communityView
	pendingPublishID         string
	pendingPublishAt         time.Time
	publishAwaitingID        string
	pendingPackPublishID     string
	pendingPackPublishAt     time.Time
	packPublishAwaitingID    string
	pendingDeleteKind        string
	pendingDeleteID          string
	pendingDeleteUntil       time.Time
	pendingUnpublishKind     string
	pendingUnpublishID       string
	publishDraftID           string
	publishTitle             string
	publishDescription       string
	publishTags              string
	publishField             int
	publishSubmitOfficial    bool
	publishRightsConfirmed   bool
	publishPreviewRaw        [][]string
	communityImportRaw       string
	communityImportPack      editorPack
	importTileSize           int
	importVerticalPairs      bool
	newArtTitle              string
	artSearch                string
	artSearchActive          bool
	packSetupID              string
	packSetupTitle           string
	packSetupDescription     string
	packSetupItems           []community.PackItem
	packSetupPreview         int
	packSetupPreviewRaw      [][]string
	packSetupField           int
	publishedEditIndex       int
	publishedEditKind        string
	publishedEditID          string
	publishedEditTitle       string
	publishedEditDescription string
	publishedEditField       int
	publishedEditLevels      []community.LevelVersion
	communityNotice          string
	communityNoticeUntil     time.Time
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
		puzzle:             loaded.Puzzle,
		board:              nonogram.NewBoard(loaded.Puzzle.Width, loaded.Puzzle.Height),
		rowClues:           nonogram.RowClues(loaded.Puzzle.Solution),
		colClues:           nonogram.ColumnClues(loaded.Puzzle.Solution),
		skeleton:           loaded.Skeleton,
		reveal:             loaded.Reveal,
		skeletonPixels:     loaded.SkeletonPixels,
		revealPixels:       loaded.RevealPixels,
		icons:              icons,
		lastCellX:          -1,
		lastCellY:          -1,
		correctFlashX:      -1,
		correctFlashY:      -1,
		startTime:          time.Now(),
		revealStart:        time.Now(),
		audioEnabled:       true,
		autoCorrect:        false,
		mode:               screenMainMenu,
		bestTimes:          loadBestTimes(),
		levelThumbs:        loadLevelThumbs(),
		editor:             initialEditor(),
		profileArt:         initialProfileArt(),
		profileBio:         loadCommunityBio(),
		profileSocial:      loadCommunitySocial(),
		profilePalette:     loadCommunityPalette(),
		profilePaletteSlot: -1,
		profileColor:       loadCommunityFavoriteColor(),
		profileName:        loadCommunityName(),
		communityLibrary:   loadCommunityLibrary(),
		selectedCreator:    -1,
		selectedGallery:    -1,
		galleryKind:        "all",
		gallerySort:        "new",
		importTileSize:     10,
		editorLastX:        -1,
		editorLastY:        -1,
	}
	g.sparkles = makeSparkles()
	g.syncCommunityProfileArt()
	return g, nil
}

func initialEditor() editorState {
	if saved := loadEditorPack(); saved != "" {
		if editor, err := editorFromPackJSON(saved); err == nil {
			return editor
		}
	}
	return newEditorState(10)
}

func initialProfileArt() editorState {
	if saved := loadCommunityProfile(); saved != "" {
		if editor, err := editorFromPackJSON(saved); err == nil && editor.Width == 10 && editor.Height == 10 {
			editor.selectLayer(editorLayerAfter)
			return editor
		}
	}
	profile := newEditorState(10)
	profile.Title = "Profile"
	return profile
}

func loadBestTimes() map[string]time.Duration {
	best := make(map[string]time.Duration, len(gameLevels))
	for _, level := range gameLevels {
		if !level.Available {
			continue
		}
		if d := loadSavedBest(level.ID); d > 0 {
			best[level.ID] = d
		}
	}
	return best
}

func loadLevelThumbs() map[string][][]assets.PixelCell {
	thumbs := make(map[string][][]assets.PixelCell, len(gameLevels))
	for _, level := range gameLevels {
		if !level.Available {
			continue
		}
		loaded, err := assets.LoadPuzzleAssets(level.Path)
		if err == nil {
			thumbs[level.ID] = loaded.RevealPixels
		}
	}
	return thumbs
}

func (g *Game) Update() error {
	g.layout = calculateLayout(g.puzzle.Width, g.puzzle.Height)
	g.updateMusicMode()
	g.updateInput()
	return nil
}

func (g *Game) updateMusicMode() {
	communityOn := g.mode == screenCommunity || g.communityPreview
	if g.communityMusicOn == communityOn {
		return
	}
	g.communityMusicOn = communityOn
	if communityOn {
		setWebMusicMode("community")
	} else {
		setWebMusicMode("default")
	}
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
	g.editorPreview = false
	g.communityPreview = false
	g.mode = screenPuzzle
	return nil
}

func (g *Game) loadEditorPuzzle() {
	_ = saveEditorPack(g.editor.packJSON())
	p := g.editor.puzzle()
	g.puzzle = p
	g.board = nonogram.NewBoard(p.Width, p.Height)
	g.rowClues = nonogram.RowClues(p.Solution)
	g.colClues = nonogram.ColumnClues(p.Solution)
	g.skeleton = nil
	g.reveal = nil
	g.skeletonPixels = g.editor.pixelMatrix(editorLayerBefore)
	g.revealPixels = g.editor.pixelMatrix(editorLayerAfter)
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
	g.editorPreview = true
	g.mode = screenPuzzle
}

func (g *Game) pushEditorUndo() {
	g.editorUndo = append(g.editorUndo, g.editor.clone())
	if len(g.editorUndo) > 40 {
		g.editorUndo = g.editorUndo[1:]
	}
}

func (g *Game) undoEditor() {
	if len(g.editorUndo) == 0 {
		return
	}
	last := g.editorUndo[len(g.editorUndo)-1]
	g.editorUndo = g.editorUndo[:len(g.editorUndo)-1]
	g.editor = last
}

func (g *Game) resetEditor(size int) {
	g.pushEditorUndo()
	g.editor = g.editor.resized(size)
	g.editorPointer = false
	g.editorLastX = -1
	g.editorLastY = -1
}

func (g *Game) saveEditor() {
	if g.saveCurrentDraft(false) == nil {
		g.showMenuNotice("saved")
		return
	}
	g.showMenuNotice("save unavailable")
}

func (g *Game) exportEditor() {
	if exportEditorImage("community_nongrams-art.jpg", g.editor.imageExportJSON()) {
		g.showMenuNotice("exported")
		return
	}
	g.showMenuNotice("export unavailable")
}

func (g *Game) leavePuzzle() {
	if g.editorPreview {
		g.editorPreview = false
		g.mode = screenEditor
		return
	}
	if g.communityPreview {
		g.communityPreview = false
		if g.activeCommunityPack != "" {
			g.communityView = communityPacks
		} else {
			g.communityView = g.communityPlayReturn
		}
		g.mode = screenCommunity
		return
	}
	g.mode = screenMainMenu
}

func (g *Game) leaveReveal() {
	if g.editorPreview {
		g.editorPreview = false
		g.mode = screenEditor
		return
	}
	if g.communityPreview {
		g.communityPreview = false
		if g.activeCommunityPack != "" {
			g.communityView = communityPacks
		} else {
			g.communityView = g.communityPlayReturn
		}
		g.mode = screenCommunity
		return
	}
	g.mode = screenLevelSelect
}

func (g *Game) loadLevel(index int) error {
	if index < 0 || index >= len(gameLevels) {
		return fmt.Errorf("level %d is not available", index+1)
	}
	level := gameLevels[index]
	if !level.Available {
		return fmt.Errorf("level %d is not available", index+1)
	}
	return g.loadPuzzle(level.Path)
}

func (g *Game) prevLevelPage() {
	if g.levelPage > 0 {
		g.levelPage--
	}
}

func (g *Game) nextLevelPage() {
	if g.levelPage < levelSelectPages()-1 {
		g.levelPage++
	}
}

func levelSelectPages() int {
	pages := (len(gameLevels) + levelSelectPageSize - 1) / levelSelectPageSize
	if pages < 1 {
		return 1
	}
	return pages
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
	g.editorPreview = false
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
		if g.communityPlayReturn == communityBrowse || g.communityPlayReturn == communityGalleryPack || g.communityPlayReturn == communityCreatorProfile {
			recordCommunityPlay(g.puzzle.ID, true)
			g.markCommunityLevelCompleted(g.puzzle.ID)
			requestCommunityCompleted()
		}
	}
	g.mode = screenReveal
	if g.editorPreview {
		_ = g.saveCurrentDraft(true)
	}
	if g.activeCommunityPack != "" && g.puzzle != nil {
		g.completeCommunityPackLevel(g.puzzle.ID)
	}
	g.revealStart = time.Now()
	playWebSFX("complete")
}

func (g *Game) showMenuNotice(text string) {
	g.menuNotice = text
	g.menuNoticeUntil = time.Now().Add(1800 * time.Millisecond)
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

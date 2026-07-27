package game

import (
	"encoding/json"
	"image"
	"strings"
	"time"

	"github.com/BakedSoups/community_nongrams/internal/community"
	"github.com/BakedSoups/community_nongrams/internal/nonogram"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) updateInput() {
	x, y, _, justPressed, _ := pointerState()
	if justPressed {
		for _, button := range renderedButtonRects {
			if button.Contains(x, y) {
				playWebSFX("menu")
				break
			}
		}
	}

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
	if g.mode == screenEditor {
		g.updateEditorInput()
		return
	}
	if g.mode == screenCommunity {
		g.updateCommunityInput()
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
		g.leavePuzzle()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.godModeFill()
		return
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		x, y := ebiten.CursorPosition()
		cellX, cellY, ok := g.layout.CellAt(x, y, g.board.Width, g.board.Height)
		if ok {
			g.pushUndo()
			next := nonogram.CellMarked
			if g.board.Cells[cellY][cellX] == nonogram.CellMarked {
				next = nonogram.CellEmpty
			}
			next, corrected := g.correctedStrokeState(cellX, cellY, next)
			if g.board.SetCell(cellX, cellY, next) {
				if corrected {
					g.timePenalty += 10 * time.Second
					g.penaltyFlashUntil = time.Now().Add(900 * time.Millisecond)
					g.correctFlashUntil = time.Now().Add(850 * time.Millisecond)
					g.correctFlashX, g.correctFlashY = cellX, cellY
					playWebSFX("correct")
				} else {
					playWebSFX("eraser")
				}
				if nonogram.IsSolved(g.board, g.puzzle.Solution) {
					g.completePuzzle()
				}
			}
		}
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
			g.leavePuzzle()
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

func updateTextField(value string, maxLen int, allow func(rune) bool) (string, bool) {
	changed := false
	if !textPasteShortcutPressed() {
		for _, char := range ebiten.AppendInputChars(nil) {
			var appended bool
			value, appended = appendAllowedText(value, string(char), maxLen, allow)
			changed = changed || appended
		}
	}
	if pasted := takeTextPaste(); pasted != "" {
		var appended bool
		value, appended = appendAllowedText(value, pasted, maxLen, allow)
		changed = changed || appended
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(value) > 0 {
		value = value[:len(value)-1]
		changed = true
	}
	return value, changed
}

func appendAllowedText(value, text string, maxLen int, allow func(rune) bool) (string, bool) {
	if len(value) >= maxLen {
		return value, false
	}
	changed := false
	for _, char := range text {
		if char == '\n' || char == '\r' || char == '\t' {
			char = ' '
		}
		if !allow(char) || len(value) >= maxLen {
			continue
		}
		value += string(char)
		changed = true
	}
	return value, changed
}

func allowPrintableText(char rune) bool {
	return char >= 32 && char <= 126
}

func allowEmailText(char rune) bool {
	return char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || strings.ContainsRune("@._+-", char)
}

func textPasteShortcutPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyV) && (ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta))
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
		g.leaveReveal()
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
		g.leaveReveal()
	}
}

func (g *Game) updateMainMenuInput() {
	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	switch {
	case mainLevelButton().Contains(x, y):
		g.mode = screenLevelSelect
	case mainCommunityButton().Contains(x, y):
		g.communityView = communityHome
		g.mode = screenCommunity
	case mainSettingsButton().Contains(x, y):
		g.mode = screenSettings
	}
}

func (g *Game) consumeCommunityPayload(take func() string, load func(string) error, errorMessage string) {
	raw := take()
	if raw == "" {
		return
	}
	if err := load(raw); err != nil {
		g.showCommunityNotice(errorMessage)
	}
}

func (g *Game) updateCommunityInput() {
	if g.pendingPackPublishID != "" && !time.Now().Before(g.pendingPackPublishAt) {
		id := g.pendingPackPublishID
		g.pendingPackPublishID = ""
		g.publishLocalPack(id)
		return
	}
	if g.pendingPublishID != "" && !time.Now().Before(g.pendingPublishAt) {
		id := g.pendingPublishID
		g.pendingPublishID = ""
		g.publishCommunityDraft(id)
		return
	}
	if id := takeCommunityPublishedID(); id != "" {
		g.markCommunityDraftPublished(id)
	}
	if id := takeCommunityPublishedPackID(); id != "" {
		g.markCommunityPackPublished(id)
	}
	g.consumeCommunityPayload(takeCommunityGallery, g.loadCommunityGallery, "could not load gallery")
	g.consumeCommunityPayload(takeCommunityChat, g.loadCommunityChat, "could not load chat")
	g.consumeCommunityPayload(takeCommunityPublished, g.loadCommunityPublished, "could not load published work")
	g.consumeCommunityPayload(takeCommunityCompleted, g.loadCommunityCompleted, "could not load completed levels")
	g.consumeCommunityPayload(takeCommunityCreators, g.loadCommunityCreators, "could not load creators")
	g.consumeCommunityPayload(takeCommunityCloudDrafts, g.mergeCloudDrafts, "could not sync cloud drafts")
	g.consumeCommunityPayload(takeCommunityCatalog, g.loadCommunityCatalog, "could not load community levels")
	if result := takeCommunityResult(); result != "" {
		g.publishAwaitingID = ""
		g.packPublishAwaitingID = ""
		if strings.HasSuffix(result, " unpublished") {
			g.markCommunityItemUnpublished(g.pendingUnpublishKind, g.pendingUnpublishID)
		}
		g.showCommunityNotice(result)
	}
	if raw := takeCommunityImport(); raw != "" {
		if err := g.loadCommunityImportPreview(raw); err != nil {
			g.showCommunityNotice("import failed: " + err.Error())
		}
	}
	if message := takeCommunityImportError(); message != "" {
		g.showCommunityNotice("import failed: " + message)
	}
	if raw := takeCommunityCoverImport(); raw != "" {
		var preview [][]string
		if json.Unmarshal([]byte(raw), &preview) == nil {
			if g.communityView == communityPublishSetup {
				g.publishPreviewRaw = preview
			} else if g.communityView == communityPackSetup {
				g.packSetupPreviewRaw = preview
				g.packSetupPreview = -1
			}
		}
	}
	if raw := takeEditorColorPicker(); raw != "" && g.communityView == communitySignIn && communitySignedIn() {
		g.applyProfilePickedColor(raw)
	}
	if g.communityView == communitySignIn && !communitySignedIn() {
		g.communityEmail, _ = updateTextField(g.communityEmail, 80, allowEmailText)
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.submitCommunitySignIn()
			return
		}
	}
	if g.communityView == communitySignIn && communitySignedIn() && g.profileBioEditing {
		g.profileBioDraft, _ = updateTextField(g.profileBioDraft, 50, allowPrintableText)
	}
	if g.communityView == communitySignIn && communitySignedIn() && g.profileNameEditing {
		g.profileNameDraft, _ = updateTextField(g.profileNameDraft, 40, allowPrintableText)
	}
	if g.communityView == communitySignIn && communitySignedIn() && g.profileSocialEditing && g.profileSocialSlot >= 0 && g.profileSocialSlot < len(g.profileSocialDrafts) {
		g.profileSocialDrafts[g.profileSocialSlot], _ = updateTextField(g.profileSocialDrafts[g.profileSocialSlot], 120, allowPrintableText)
	}
	if g.communityView == communityChat {
		g.chatDraft, _ = updateTextField(g.chatDraft, 220, allowPrintableText)
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.sendCommunityChat()
			return
		}
	}
	if g.communityView == communityPublishSetup {
		switch g.publishField {
		case 0:
			g.publishTitle, _ = updateTextField(g.publishTitle, 80, allowPrintableText)
		case 1:
			g.publishDescription, _ = updateTextField(g.publishDescription, 160, allowPrintableText)
		case 2:
			g.publishTags, _ = updateTextField(g.publishTags, 100, allowPrintableText)
		}
	}
	if g.communityView == communityNewArtSetup {
		g.newArtTitle, _ = updateTextField(g.newArtTitle, 80, allowPrintableText)
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.startNewCommunityArt()
			return
		}
	}
	if g.communityView == communityPackSetup {
		if g.packSetupField == 0 {
			g.packSetupTitle, _ = updateTextField(g.packSetupTitle, 80, allowPrintableText)
		}
		if g.packSetupField == 1 {
			g.packSetupDescription, _ = updateTextField(g.packSetupDescription, 200, allowPrintableText)
		}
	}
	if g.communityView == communityPublishedEdit {
		if g.publishedEditField == 0 {
			g.publishedEditTitle, _ = updateTextField(g.publishedEditTitle, 80, allowPrintableText)
		}
		if g.publishedEditField == 1 {
			g.publishedEditDescription, _ = updateTextField(g.publishedEditDescription, 200, allowPrintableText)
		}
	}
	if g.communityView == communityMyArt && g.artSearchActive {
		var changed bool
		g.artSearch, changed = updateTextField(g.artSearch, 60, allowPrintableText)
		if changed {
			g.communityPage = 0
		}
	}
	if g.communityView == communityCreators && g.creatorSearchActive {
		var changed bool
		g.creatorSearch, changed = updateTextField(g.creatorSearch, 60, allowPrintableText)
		if changed {
			g.communityPage = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.communityBack()
		return
	}
	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	if communityBackButton().Contains(x, y) {
		g.communityBack()
		return
	}
	if g.communityView == communityHome && communityProfileBadgeButton().Contains(x, y) {
		if communitySignedIn() {
			g.openProfileEditor()
		} else {
			g.communityView = communitySignIn
		}
		return
	}
	if g.communityView == communityHome && communityAccountButton().Contains(x, y) {
		g.profileBioDraft = truncateText(g.profileBio, 50)
		g.profileNameDraft = g.profileName
		g.profileSocialDrafts = splitProfileSocials(g.profileSocial)
		g.profileSocialSlot = -1
		g.communityView = communitySignIn
		requestCommunityCompleted()
		return
	}
	switch g.communityView {
	case communityHome:
		switch {
		case communityLevelsButton().Contains(x, y):
			g.communityView = communityBrowse
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
		case communityMyArtButton().Contains(x, y):
			g.communityView = communityMyArt
			g.communityPage = 0
			g.syncLocalDrafts()
			requestCommunityCloudDrafts()
		case communityCreatorsButton().Contains(x, y):
			g.communityView = communityCreators
			g.selectedCreator = -1
			g.creatorSearchActive = false
			g.communityPage = 0
			g.syncCommunityProfileArt()
			requestCommunityCreators()
		}
	case communityCreate:
		switch {
		case communityNewButton().Contains(x, y):
			g.openNewArtSetup()
		case communityImportButton().Contains(x, y):
			g.communityView = communityImportSetup
		case communityImportHelpButton().Contains(x, y):
			g.communityView = communityImportHelp
		}
	case communityMyArt:
		if communityArtSearchField().Contains(x, y) {
			g.artSearchActive = true
			return
		}
		g.artSearchActive = false
		if communityLibraryPublishedTab().Contains(x, y) {
			g.communityView = communityPublished
			g.communityPage = 0
			requestCommunityPublished()
			return
		}
		if communityLibraryPacksTab().Contains(x, y) {
			g.communityView = communityPacks
			g.communityPage = 0
			return
		}
		if communityArtCreateButton().Contains(x, y) {
			g.communityView = communityCreate
			return
		}
		indexes := g.filteredCommunityDraftIndexes()
		start := g.communityPage * communityDraftsPerPage
		for slot := 0; slot < communityDraftsPerPage; slot++ {
			position := start + slot
			if position >= len(indexes) {
				break
			}
			index := indexes[position]
			if communityDraftEditButton(slot).Contains(x, y) {
				g.editCommunityDraft(index)
				return
			}
			if communityDraftPublishButton(slot).Contains(x, y) {
				draft := g.communityLibrary.Drafts[index]
				if draft.Status == community.LevelPublishedStatus {
					g.showPublishedManagementNotice("art")
					return
				}
				g.queueCommunityDraftPublish(index)
				return
			}
			if communityDraftDeleteButton(slot).Contains(x, y) {
				g.deleteCommunityDraft(index)
				return
			}
		}
		if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
			g.communityPage--
		}
		if communityNextButton().Contains(x, y) && (g.communityPage+1)*communityDraftsPerPage < len(indexes) {
			g.communityPage++
		}
	case communityBrowse:
		switch {
		case communityGalleryAllButton().Contains(x, y):
			g.galleryKind = "all"
			g.gallerySortOpen = false
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
			return
		case communityGalleryArtButton().Contains(x, y):
			g.galleryKind = "art"
			g.gallerySortOpen = false
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
			return
		case communityGalleryPacksButton().Contains(x, y):
			g.galleryKind = "pack"
			g.gallerySortOpen = false
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
			return
		case communityGallerySortButton().Contains(x, y):
			g.gallerySortOpen = !g.gallerySortOpen
			return
		case g.gallerySortOpen && communityGalleryNewButton().Contains(x, y):
			g.gallerySort = "new"
			g.gallerySortOpen = false
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
			return
		case g.gallerySortOpen && communityGalleryPlayedButton().Contains(x, y):
			g.gallerySort = "played"
			g.gallerySortOpen = false
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
			return
		case g.gallerySortOpen && communityGalleryTopButton().Contains(x, y):
			g.gallerySort = "top"
			g.gallerySortOpen = false
			g.communityPage = 0
			requestCommunityGallery(g.galleryKind, g.gallerySort)
			return
		}
		g.gallerySortOpen = false
		start := g.communityPage * communityCatalogPerPage
		for slot := 0; slot < communityCatalogPerPage; slot++ {
			index := start + slot
			if index >= len(g.communityGallery) {
				continue
			}
			item := g.communityGallery[index]
			if communityGalleryOpenButton(slot).Contains(x, y) {
				if item.Kind == "pack" {
					g.selectedGallery = index
					g.communityView = communityGalleryPack
				} else {
					g.playGalleryLevel(index)
				}
				return
			}
			if communityGalleryChatButton(slot).Contains(x, y) {
				g.openCommunityChat(item.Kind, item.ID, item.Title, communityBrowse)
				return
			}
			if communityGalleryLikeButton(slot).Contains(x, y) {
				if !communitySignedIn() {
					g.communityView = communitySignIn
				} else {
					toggleCommunityLike(item.Kind, item.ID)
				}
				return
			}
		}
		if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
			g.communityPage--
		}
		if communityNextButton().Contains(x, y) && (g.communityPage+1)*communityCatalogPerPage < len(g.communityGallery) {
			g.communityPage++
		}
	case communityGalleryPack:
		if g.selectedGallery >= 0 && g.selectedGallery < len(g.communityGallery) {
			item := g.communityGallery[g.selectedGallery]
			if communityGalleryPackChatButton().Contains(x, y) {
				g.openCommunityChat(item.Kind, item.ID, item.Title, communityGalleryPack)
				return
			}
			for slot := 0; slot < 6 && slot < len(g.communityGallery[g.selectedGallery].Levels); slot++ {
				if communityGalleryPackLevelButton(slot).Contains(x, y) {
					g.playGalleryPackLevel(slot)
					return
				}
			}
		}
	case communityChat:
		start := len(g.communityChatMessages) - 5
		if start < 0 {
			start = 0
		}
		for slot, msg := range g.communityChatMessages[start:] {
			if communityChatMessageButton(slot).Contains(x, y) {
				g.openChatAuthorProfile(msg.AuthorID)
				return
			}
		}
		if communityChatSendButton().Contains(x, y) {
			g.sendCommunityChat()
			return
		}
	case communityPacks:
		if communityLibraryPublishedTab().Contains(x, y) {
			g.communityView = communityPublished
			g.communityPage = 0
			requestCommunityPublished()
			return
		}
		if communityLibraryArtTab().Contains(x, y) {
			g.communityView = communityMyArt
			g.communityPage = 0
			return
		}
		if communityPackCreateButton().Contains(x, y) {
			g.openPackBuilder()
			return
		}
		for slot := 0; slot < 4; slot++ {
			if communityPackPlayButton(slot).Contains(x, y) {
				g.playLocalPack(slot)
				return
			}
			if communityPackPublishButton(slot).Contains(x, y) {
				if slot < len(g.communityLibrary.Packs) && g.communityLibrary.Packs[slot].Status == community.LevelPublishedStatus {
					g.showPublishedManagementNotice("pack")
					return
				}
				g.queueLocalPackPublish(slot)
				return
			}
			if communityPackDeleteButton(slot).Contains(x, y) {
				g.deleteCommunityPack(slot)
				return
			}
		}
	case communityPublished:
		if communityLibraryArtTab().Contains(x, y) {
			g.communityView = communityMyArt
			g.communityPage = 0
			return
		}
		if communityLibraryPacksTab().Contains(x, y) {
			g.communityView = communityPacks
			g.communityPage = 0
			return
		}
		for slot := range g.communityPublished {
			if slot >= 4 {
				break
			}
			if communityPublishedRemoveButton(slot).Contains(x, y) {
				g.openPublishedEditor(slot)
				return
			}
		}
	case communityPublishedEdit:
		if g.publishedEditKind == "pack" {
			for slot := 0; slot < len(g.publishedEditLevels) && slot < 8; slot++ {
				if communityPublishedEditLevelRemoveButton(slot).Contains(x, y) {
					g.removePublishedEditPackLevel(slot)
					return
				}
			}
			if communityPublishedEditAddLevelButton().Contains(x, y) {
				g.openPublishedPackAdd()
				return
			}
		}
		switch {
		case communityPublishedEditTitleField().Contains(x, y):
			g.publishedEditField = 0
		case communityPublishedEditDescriptionField().Contains(x, y):
			g.publishedEditField = 1
		case communityPublishedApplyButton().Contains(x, y):
			g.applyPublishedEdit()
		case communityPublishedUnpublishButton().Contains(x, y):
			g.pendingUnpublishKind = g.publishedEditKind
			g.pendingUnpublishID = g.publishedEditID
			unpublishCommunityItem(g.publishedEditKind, g.publishedEditID)
		}
	case communityPublishedPackAdd:
		start := g.communityPage * communityPackDraftsPerPage
		for slot := 0; slot < communityPackDraftsPerPage; slot++ {
			if communityPackDraftButton(slot).Contains(x, y) {
				g.addPublishedEditPackDraft(start + slot)
				return
			}
		}
		if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
			g.communityPage--
		}
		if communityNextButton().Contains(x, y) && (g.communityPage+1)*communityPackDraftsPerPage < len(g.communityLibrary.Drafts) {
			g.communityPage++
		}
	case communityPublishSetup:
		switch {
		case communityPublishCoverButton().Contains(x, y):
			requestCommunityCoverImport(10)
		case communityPublishFinalCoverButton().Contains(x, y):
			g.publishPreviewRaw = nil
		case communityPublishQuestionCoverButton().Contains(x, y):
			g.publishPreviewRaw = communityQuestionCover()
		case communityPublishTitleField().Contains(x, y):
			g.publishField = 0
		case communityPublishDescriptionField().Contains(x, y):
			g.publishField = 1
		case communityPublishTagsField().Contains(x, y):
			g.publishField = 2
		case communityPublishOfficialButton().Contains(x, y):
			g.publishSubmitOfficial = !g.publishSubmitOfficial
			if !g.publishSubmitOfficial {
				g.publishRightsConfirmed = false
			} else {
				g.showCommunityNotice("main game review request added")
			}
		case communityPublishRightsButton().Contains(x, y) && g.publishSubmitOfficial:
			g.publishRightsConfirmed = !g.publishRightsConfirmed
		case communityPublishConfirmButton().Contains(x, y):
			if g.publishAwaitingID == "" {
				g.submitCommunityDraftPublish()
			}
		}
	case communityImportPreview:
		if communityImportConfirmButton().Contains(x, y) {
			if err := g.importCommunityPack(g.communityImportRaw); err != nil {
				g.showCommunityNotice("import failed: " + err.Error())
			}
		}
	case communityImportSetup:
		for index, size := range []int{8, 10, 15, 20, 32} {
			if communityImportSizeButton(index).Contains(x, y) {
				g.importTileSize = size
				return
			}
		}
		switch {
		case communityImportHorizontalButton().Contains(x, y):
			g.importVerticalPairs = false
		case communityImportVerticalButton().Contains(x, y):
			g.importVerticalPairs = true
		case communityImportChooseButton().Contains(x, y):
			if !requestCommunityImport(g.importTileSize, g.importVerticalPairs) {
				g.showCommunityNotice("import is available in the web build")
			}
		}
	case communityPackBuild:
		start := g.communityPage * communityPackDraftsPerPage
		for slot := 0; slot < communityPackDraftsPerPage; slot++ {
			if communityPackDraftButton(slot).Contains(x, y) {
				g.togglePackDraft(start + slot)
				return
			}
		}
		if communityPackDoneButton().Contains(x, y) {
			g.openNewPackSetup()
			return
		}
		if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
			g.communityPage--
		}
		if communityNextButton().Contains(x, y) && (g.communityPage+1)*communityPackDraftsPerPage < len(g.communityLibrary.Drafts) {
			g.communityPage++
		}
	case communityNewArtSetup:
		if communityNewArtTitleField().Contains(x, y) {
			return
		}
		if communityNewArtStartButton().Contains(x, y) {
			g.startNewCommunityArt()
		}
	case communityPackSetup:
		if communityPackUploadCoverButton().Contains(x, y) {
			requestCommunityCoverImport(10)
			return
		}
		if communityPackQuestionCoverButton().Contains(x, y) {
			g.packSetupPreviewRaw = communityQuestionCover()
			g.packSetupPreview = -1
			return
		}
		for art := 0; art < len(g.packSetupItems) && art < 8; art++ {
			if communityPackSetupPreview(art).Contains(x, y) {
				g.packSetupPreview = art
				g.packSetupPreviewRaw = nil
				return
			}
		}
		switch {
		case communityPackTitleField().Contains(x, y):
			g.packSetupField = 0
		case communityPackDescriptionField().Contains(x, y):
			g.packSetupField = 1
		case communityPackSaveDraftButton().Contains(x, y):
			g.savePackSetup(false)
		case communityPackSetupPublishButton().Contains(x, y):
			if g.packPublishAwaitingID == "" {
				g.savePackSetup(true)
			}
		}
	case communitySignIn:
		if communitySignedIn() {
			for index := 0; index < 4; index++ {
				if communityAccountPaletteOptionButton(index).Contains(x, y) {
					g.profilePaletteSlot = index
					g.profileColorPicking = false
					g.profileNameEditing, g.profileBioEditing, g.profileSocialEditing = false, false, false
					g.profileSocialSlot = -1
					if !requestEditorColorPicker(profilePaletteColorInitial(g.profilePalette, index)) {
						g.showCommunityNotice("color picker unavailable")
					}
					return
				}
			}
			switch {
			case communityAccountNameField().Contains(x, y):
				g.profileNameEditing, g.profileBioEditing, g.profileSocialEditing = true, false, false
				g.profileSocialSlot = -1
				g.profilePaletteSlot = -1
				g.profileColorPicking = false
			case communityAccountBioField().Contains(x, y):
				g.profileBioEditing, g.profileNameEditing, g.profileSocialEditing = true, false, false
				g.profilePaletteSlot = -1
				g.profileColorPicking = false
			case communityAccountColorButton().Contains(x, y):
				g.profileColorPicking = true
				g.profilePaletteSlot = -1
				g.profileNameEditing, g.profileBioEditing, g.profileSocialEditing = false, false, false
				g.profileSocialSlot = -1
				if !requestEditorColorPicker(profileColorInitial(g.profileColor)) {
					g.showCommunityNotice("color picker unavailable")
				}
			case communityAccountBioSaveButton().Contains(x, y):
				name := strings.TrimSpace(g.profileNameDraft)
				if name == "" {
					g.showCommunityNotice("name is required")
					return
				}
				social, ok := normalizeProfileSocialList(g.profileSocialDrafts)
				if !ok {
					g.showCommunityNotice("social: supported sites only")
					return
				}
				g.profileName = name
				g.profileBio = truncateText(strings.TrimSpace(g.profileBioDraft), 50)
				g.profileSocial = social
				g.saveCommunityProfileDetails()
				if raw, err := json.Marshal(g.profileArt.puzzle()); err == nil {
					syncCommunityProfile(string(raw), g.profileBio, g.profileName, g.profileSocial, g.profilePalette, g.profileColor)
				}
				g.profileBioEditing = false
				g.profileNameEditing = false
				g.profileSocialEditing = false
				g.profileSocialSlot = -1
				g.showCommunityNotice("profile saved")
			case communitySignOutButton().Contains(x, y):
				requestCommunitySignOut()
				g.communityView = communityHome
			}
			for slot := range g.profileSocialDrafts {
				if communityAccountSocialField(slot).Contains(x, y) {
					g.profileSocialEditing, g.profileNameEditing, g.profileBioEditing = true, false, false
					g.profileSocialSlot = slot
					return
				}
			}
		} else {
			switch {
			case communityGoogleButton().Contains(x, y):
				if !requestCommunityGoogleSignIn() {
					g.showCommunityNotice("Google sign in is unavailable")
				}
			case communitySendLinkButton().Contains(x, y):
				g.submitCommunitySignIn()
			}
		}
	case communityCreators:
		if communityCreatorSearchField().Contains(x, y) {
			g.creatorSearchActive = true
			return
		}
		g.creatorSearchActive = false
		indexes := g.filteredCommunityCreatorIndexes()
		start := g.communityPage * communityCreatorsPerPage
		for slot := 0; slot < communityCreatorsPerPage && start+slot < len(indexes); slot++ {
			if communityCreatorButton(slot).Contains(x, y) {
				g.selectedCreator = indexes[start+slot]
				g.communityPage = 0
				g.communityView = communityCreatorProfile
				return
			}
		}
		if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
			g.communityPage--
		}
		if communityNextButton().Contains(x, y) && (g.communityPage+1)*communityCreatorsPerPage < len(indexes) {
			g.communityPage++
		}
	case communityCreatorProfile:
		if g.selectedCreator >= 0 && g.selectedCreator < len(g.communityCreators) {
			creator := g.communityCreators[g.selectedCreator]
			levels := creator.Levels
			contentY := communityCreatorProfileLevelsY(creator.Social, creator.Bio, creator.Palette, creator.FavoriteColor, len(creator.Featured) > 0)
			start := g.communityPage * 4
			for slot := 0; slot < 4 && start+slot < len(levels); slot++ {
				if communityCreatorLevelButtonAt(slot, contentY).Contains(x, y) {
					g.playCreatorLevel(start + slot)
					return
				}
			}
			if communityPrevButton().Contains(x, y) && g.communityPage > 0 {
				g.communityPage--
			}
			if communityNextButton().Contains(x, y) && (g.communityPage+1)*4 < len(levels) {
				g.communityPage++
			}
		}
	}
}

func (g *Game) submitCommunitySignIn() {
	email := strings.TrimSpace(g.communityEmail)
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		g.showCommunityNotice("enter a valid email address")
		return
	}
	if !requestCommunitySignIn(email) {
		g.showCommunityNotice("sign in is available in the web build")
	}
}

func (g *Game) communityBack() {
	if g.communityView == communityNewArtSetup {
		g.communityView = communityCreate
		return
	}
	if g.communityView == communityPackSetup {
		if g.packSetupID == "" {
			g.communityView = communityPackBuild
		} else {
			g.communityView = communityPacks
		}
		return
	}
	if g.communityView == communityChat {
		if g.chatReturn != 0 {
			g.communityView = g.chatReturn
		} else {
			g.communityView = communityBrowse
		}
		return
	}
	if g.communityView == communityPublishSetup {
		g.communityView = communityMyArt
		return
	}
	if g.communityView == communityPublishedEdit {
		g.communityView = communityPublished
		return
	}
	if g.communityView == communityPublishedPackAdd {
		g.communityView = communityPublishedEdit
		return
	}
	if g.communityView == communityImportPreview {
		g.communityView = communityCreate
		return
	}
	if g.communityView == communityImportSetup {
		g.communityView = communityCreate
		return
	}
	if g.communityView == communityImportHelp {
		g.communityView = communityCreate
		return
	}
	if g.communityView == communityCreate {
		g.communityView = communityMyArt
		return
	}
	if g.communityView == communityGalleryPack {
		g.communityView = communityBrowse
		return
	}
	if g.communityView == communityCreatorProfile {
		g.communityView = communityCreators
		return
	}
	if g.communityView == communityPackBuild {
		g.communityView = communityPacks
		return
	}
	if g.communityView == communityHome {
		g.mode = screenMainMenu
		return
	}
	g.communityView = communityHome
}

func (g *Game) updateEditorInput() {
	if g.editorTitleEditing {
		g.updateEditorTitleInput()
		return
	}
	if raw := takeEditorColorPicker(); raw != "" {
		if c, ok := parseEditorHexColor(raw); ok {
			g.editor.selectPaintColor(c)
			g.editor.Tool = editorToolPencil
		}
	}
	if raw := takeEditorImageImport(); raw != "" {
		g.pushEditorUndo()
		if err := g.editor.importPayload(raw); err != nil {
			g.showMenuNotice("import failed")
		} else {
			g.showMenuNotice("image imported")
		}
	}
	if raw := takeEditorPackImport(); raw != "" {
		g.pushEditorUndo()
		if editor, err := editorFromPackJSON(raw); err != nil {
			g.showMenuNotice("pack failed")
		} else {
			g.editor = editor
			g.showMenuNotice("pack imported")
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.editorSizeOpen {
			g.editorSizeOpen = false
			return
		}
		if g.editingProfile {
			g.closeProfileEditor(false)
			return
		}
		_ = g.saveCurrentDraft(false)
		g.communityView = communityMyArt
		g.mode = screenCommunity
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.undoEditor()
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.editor.Tool = editorToolPencil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		g.editor.Tool = editorToolEraser
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.editor.Tool = editorToolFill
	}

	x, y, down, justPressed, justReleased := pointerState()
	if justReleased {
		g.editorPointer = false
		g.editorLastX = -1
		g.editorLastY = -1
	}
	if justPressed {
		if g.editorSizeOpen {
			if editorSizeButton().Contains(x, y) {
				g.editorSizeOpen = false
				return
			}
			for _, size := range editorSizes {
				if editorSizeOption(size).Contains(x, y) {
					g.resetEditor(size)
					g.editorSizeOpen = false
					return
				}
			}
			g.editorSizeOpen = false
			return
		}
		if g.handleEditorButton(x, y) {
			return
		}
		for i, c := range editorPalette {
			if editorPaletteRect(i).Contains(x, y) {
				g.editor.selectPaintColor(c)
				g.editor.Tool = editorToolPencil
				return
			}
		}
		cellX, cellY, ok := editorCellAt(g.editor, x, y)
		if ok {
			g.pushEditorUndo()
			g.editor.apply(cellX, cellY)
			g.editorPointer = true
			g.editorLastX = cellX
			g.editorLastY = cellY
			return
		}
	}
	if !down || !g.editorPointer || g.editor.Tool == editorToolFill || g.editor.Tool == editorToolEyedropper {
		return
	}
	cellX, cellY, ok := editorCellAt(g.editor, x, y)
	if !ok || (cellX == g.editorLastX && cellY == g.editorLastY) {
		return
	}
	g.editor.applyLine(g.editorLastX, g.editorLastY, cellX, cellY)
	g.editorLastX = cellX
	g.editorLastY = cellY
}

func isEditorNoticeText(title string) bool {
	switch title {
	case "saved", "save unavailable", "exported", "export unavailable":
		return true
	default:
		return false
	}
}

func normalizeProfileSocial(value string) (string, bool) {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		return "", true
	}
	lower := strings.ToLower(value)
	if strings.Contains(lower, "http://") || strings.Contains(lower, "https://") || strings.Contains(lower, "www.") || strings.ContainsAny(value, `/\`) {
		normalized, ok := normalizeProfileSocialLink(value)
		if !ok {
			return "", false
		}
		return normalized, true
	}
	if len(value) > 80 {
		value = value[:80]
	}
	return value, true
}

func splitProfileSocials(value string) [3]string {
	var socials [3]string
	parts := strings.Split(value, " | ")
	for i := range socials {
		if i >= len(parts) {
			break
		}
		socials[i] = strings.TrimSpace(parts[i])
	}
	return socials
}

func countProfileSocials(value string) int {
	count := 0
	for _, social := range splitProfileSocials(value) {
		if social != "" {
			count++
		}
	}
	return count
}

func normalizeProfileSocialList(values [3]string) (string, bool) {
	socials := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		normalized, ok := normalizeProfileSocial(value)
		if !ok {
			return "", false
		}
		if normalized == "" {
			continue
		}
		platform := socialPlatform(normalized)
		if platform != "" {
			if seen[platform] {
				continue
			}
			seen[platform] = true
		}
		socials = append(socials, normalized)
	}
	result := strings.Join(socials, " | ")
	if len(result) > 160 {
		result = result[:160]
	}
	return result, true
}

func (g *Game) applyProfilePickedColor(raw string) {
	c, ok := parseEditorHexColor(raw)
	if !ok {
		return
	}
	value := editorColorHex(c)
	if g.profilePaletteSlot >= 0 {
		g.profilePalette = setProfilePaletteSlot(g.profilePalette, g.profilePaletteSlot, value)
		g.profilePaletteSlot = -1
		return
	}
	if g.profileColorPicking {
		g.profileColor = value
		g.profileColorPicking = false
	}
}

func socialPlatform(value string) string {
	prefix, _, ok := strings.Cut(strings.ToLower(strings.TrimSpace(value)), ":")
	if !ok {
		return ""
	}
	switch prefix {
	case "x", "github", "instagram", "tiktok", "youtube", "twitch", "bluesky", "threads", "mastodon", "linkedin":
		return prefix
	default:
		return ""
	}
}

func normalizeProfileSocialLink(value string) (string, bool) {
	lower := strings.ToLower(strings.TrimSpace(value))
	lower = strings.TrimPrefix(lower, "https://")
	lower = strings.TrimPrefix(lower, "http://")
	lower = strings.TrimPrefix(lower, "www.")
	lower = strings.TrimRight(lower, "/")
	parts := strings.Split(lower, "/")
	if len(parts) < 2 || parts[1] == "" {
		return "", false
	}
	domain := parts[0]
	handle := strings.TrimPrefix(parts[1], "@")
	if strings.ContainsAny(handle, `?#\`) {
		fields := strings.FieldsFunc(handle, func(r rune) bool {
			return r == '?' || r == '#' || r == '\\'
		})
		if len(fields) == 0 {
			return "", false
		}
		handle = fields[0]
	}
	platform := ""
	switch domain {
	case "twitter.com", "x.com":
		platform = "x"
	case "github.com":
		platform = "github"
	case "instagram.com":
		platform = "instagram"
	case "tiktok.com":
		platform = "tiktok"
	case "youtube.com", "youtu.be":
		platform = "youtube"
	case "twitch.tv":
		platform = "twitch"
	case "bsky.app":
		if len(parts) >= 3 && parts[1] == "profile" {
			handle = strings.TrimPrefix(parts[2], "@")
		}
		platform = "bluesky"
	case "threads.net":
		platform = "threads"
	case "mastodon.social":
		platform = "mastodon"
	case "linkedin.com":
		if len(parts) >= 3 && (parts[1] == "in" || parts[1] == "company") {
			handle = parts[2]
		}
		platform = "linkedin"
	default:
		return "", false
	}
	if handle == "" {
		return "", false
	}
	value = platform + ": " + handle
	if len(value) > 80 {
		value = value[:80]
	}
	return value, true
}

func (g *Game) openEditorTitleDialog() {
	g.editorTitleDraft = strings.TrimSpace(g.editor.Title)
	g.editorTitleEditing = true
	g.editorSizeOpen = false
}

func (g *Game) closeEditorTitleDialog(save bool) {
	if save {
		title := strings.TrimSpace(g.editorTitleDraft)
		if title == "" {
			title = "Untitled"
		}
		g.editor.Title = title
		_ = g.saveCurrentDraft(false)
	}
	g.editorTitleEditing = false
}

func (g *Game) updateEditorTitleInput() {
	g.editorTitleDraft, _ = updateTextField(g.editorTitleDraft, 80, allowPrintableText)
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.closeEditorTitleDialog(true)
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.closeEditorTitleDialog(false)
		return
	}
	x, y, _, justPressed, _ := pointerState()
	if !justPressed {
		return
	}
	switch {
	case editorTitleDialogSaveButton().Contains(x, y):
		g.closeEditorTitleDialog(true)
	case editorTitleDialogCancelButton().Contains(x, y):
		g.closeEditorTitleDialog(false)
	}
}

func (g *Game) handleEditorButton(x, y int) bool {
	if g.editingProfile {
		switch {
		case editorBackButton().Contains(x, y):
			g.closeProfileEditor(false)
			return true
		case profileSaveButton().Contains(x, y):
			g.closeProfileEditor(true)
			return true
		}
	}
	switch {
	case editorBackButton().Contains(x, y):
		_ = g.saveCurrentDraft(false)
		g.communityView = communityMyArt
		g.mode = screenCommunity
	case editorUndoButton().Contains(x, y):
		g.undoEditor()
	case editorSizeButton().Contains(x, y):
		g.editorSizeOpen = !g.editorSizeOpen
	case editorTitleButton().Contains(x, y):
		g.openEditorTitleDialog()
	case editorPreviewButton().Contains(x, y):
		g.loadEditorPuzzle()
	case editorSaveButton().Contains(x, y):
		g.saveEditor()
	case editorExportButton().Contains(x, y):
		g.exportEditor()
	case editorPencilButton().Contains(x, y):
		g.editor.Tool = editorToolPencil
	case editorEraserButton().Contains(x, y):
		g.editor.Tool = editorToolEraser
	case editorFillButton().Contains(x, y):
		g.editor.Tool = editorToolFill
	case editorEyeButton().Contains(x, y):
		g.editor.Tool = editorToolEyedropper
	case editorBeforeButton().Contains(x, y):
		g.editor.selectLayer(editorLayerBefore)
	case editorAfterButton().Contains(x, y):
		g.editor.selectLayer(editorLayerAfter)
	case editorLayerPreviewButton().Contains(x, y) && g.editor.Layer == editorLayerAfter:
		g.editorOnionSkin = !g.editorOnionSkin
	case editorRainbowRect().Contains(x, y):
		if g.editor.Layer == editorLayerBefore {
			g.editor.selectPaintColor(editorBeforeColor)
			break
		}
		if !requestEditorColorPicker(editorColorHex(g.editor.PaintColor)) {
			g.showMenuNotice("color picker unavailable")
		}
	default:
		return false
	}
	return true
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
			if levelIndex < len(gameLevels) && gameLevels[levelIndex].Available {
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

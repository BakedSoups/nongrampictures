package game

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/BakedSoups/community_nongrams/internal/assets"
	"github.com/BakedSoups/community_nongrams/internal/community"
	"github.com/BakedSoups/community_nongrams/internal/nonogram"
)

type communityView uint8

const (
	communityHome communityView = iota
	communityBrowse
	communityPacks
	communityCreate
	communityMyArt
	communitySignIn
	communityPackBuild
	communityCreators
	communityCreatorProfile
	communityGalleryPack
	communityImportHelp
	communityPublished
	communityPublishSetup
	communityImportPreview
	communityImportSetup
	communityNewArtSetup
	communityPackSetup
	communityChat
	communityPublishedEdit
	communityPublishedPackAdd
)

func loadCommunityLibrary() community.Library {
	var library community.Library
	raw := loadCommunityData()
	if raw == "" || json.Unmarshal([]byte(raw), &library) != nil {
		return community.Library{}
	}
	return library
}

func (g *Game) saveCommunityLibrary() bool {
	raw, err := json.Marshal(g.communityLibrary)
	return err == nil && saveCommunityData(string(raw))
}

func (g *Game) newCommunityDraft(size int, title string) {
	g.editor = newEditorState(size)
	g.editor.Title = strings.TrimSpace(title)
	g.currentDraftID = newLocalID("level")
	g.editorUndo = nil
	g.editorSizeOpen = false
	g.editorOnionSkin = false
	g.mode = screenEditor
	_ = g.saveCurrentDraft(false)
}

func (g *Game) openNewArtSetup() {
	g.newArtTitle = ""
	g.communityView = communityNewArtSetup
}

func (g *Game) startNewCommunityArt() {
	title := strings.TrimSpace(g.newArtTitle)
	if title == "" {
		g.showCommunityNotice("title is required")
		return
	}
	g.newCommunityDraft(10, title)
}

func communityQuestionCover() [][]string {
	rows := make([][]string, 10)
	for y := range rows {
		rows[y] = make([]string, 10)
	}
	for _, point := range [][2]int{{3, 2}, {4, 2}, {5, 2}, {6, 2}, {2, 3}, {7, 3}, {6, 4}, {5, 5}, {5, 6}, {5, 8}} {
		rows[point[1]][point[0]] = "#484343FF"
	}
	return rows
}

func (g *Game) openProfileEditor() {
	g.profileReturn = g.editor.clone()
	g.profileDraftID = g.currentDraftID
	g.editor = g.profileArt.clone()
	g.editor.Title = "Profile"
	g.editor.selectLayer(editorLayerAfter)
	g.editorUndo = nil
	g.editorSizeOpen = false
	g.editorOnionSkin = false
	g.editingProfile = true
	g.mode = screenEditor
}

func (g *Game) closeProfileEditor(save bool) {
	if save {
		g.editor.Title = "Profile"
		g.editor.selectLayer(editorLayerAfter)
		g.profileArt = g.editor.clone()
		saveCommunityProfile(g.profileArt.packJSON())
		g.saveCommunityProfileDetails()
		if raw, err := json.Marshal(g.profileArt.puzzle()); err == nil {
			syncCommunityProfile(string(raw), g.profileBio, g.profileName, g.profileSocial, g.profilePalette, g.profileColor)
		}
		g.showCommunityNotice("profile saved")
	}
	g.editor = g.profileReturn
	g.currentDraftID = g.profileDraftID
	g.editingProfile = false
	g.editorUndo = nil
	g.communityView = communityHome
	g.mode = screenCommunity
}

func (g *Game) syncCommunityProfileArt() {
	if !communitySignedIn() {
		return
	}
	if raw, err := json.Marshal(g.profileArt.puzzle()); err == nil {
		syncCommunityProfile(string(raw), g.profileBio, g.profileName, g.profileSocial, g.profilePalette, g.profileColor)
	}
}

func (g *Game) saveCommunityProfileDetails() {
	saveCommunityBio(g.profileBio)
	saveCommunitySocial(g.profileSocial)
	saveCommunityName(g.profileName)
	saveCommunityPalette(g.profilePalette)
	saveCommunityFavoriteColor(g.profileColor)
}

func (g *Game) editCommunityDraft(index int) {
	if index < 0 || index >= len(g.communityLibrary.Drafts) {
		return
	}
	draft := g.communityLibrary.Drafts[index]
	raw, err := json.Marshal(editorPack{Levels: []*nonogram.Puzzle{draft.Puzzle}})
	if err != nil {
		return
	}
	editor, err := editorFromPackJSON(string(raw))
	if err != nil {
		g.showCommunityNotice("draft could not be opened")
		return
	}
	g.editor = editor
	g.editor.Title = draft.Title
	g.currentDraftID = draft.ID
	g.editorUndo = nil
	g.mode = screenEditor
}

func (g *Game) saveCurrentDraft(playtested bool) error {
	puzzle := g.editor.puzzle()
	id := g.currentDraftID
	if id == "" {
		id = newLocalID("level")
		g.currentDraftID = id
	}
	puzzle.ID = id
	title := strings.TrimSpace(g.editor.Title)
	if title == "" {
		title = "Untitled"
	}
	puzzle.Title = title
	draft := community.NewDraft(id, puzzle)
	if existing, ok := g.communityLibrary.Draft(id); ok {
		draft.OwnerID = existing.OwnerID
		draft.Description = existing.Description
		draft.Tags = append([]string(nil), existing.Tags...)
		draft.Visibility = existing.Visibility
		draft.Status = existing.Status
		draft.Version = existing.Version
		draft.Playtested = existing.Playtested || playtested
		draft.Stats = existing.Stats
	}
	if err := draft.ValidateForSave(); err != nil {
		return err
	}
	g.communityLibrary.UpsertDraft(draft)
	if !g.saveCommunityLibrary() {
		return fmt.Errorf("local storage is unavailable")
	}
	if raw, err := json.Marshal(draft); err == nil {
		syncCommunityDraft(string(raw))
	}
	_ = saveEditorPack(g.editor.packJSON())
	return nil
}

func (g *Game) mergeCloudDrafts(raw string) error {
	var drafts []community.LevelDraft
	if err := json.Unmarshal([]byte(raw), &drafts); err != nil {
		return err
	}
	for _, draft := range drafts {
		if draft.Puzzle == nil || draft.ValidateForSave() != nil {
			continue
		}
		existing, ok := g.communityLibrary.Draft(draft.ID)
		if !ok || existing.UpdatedAt < draft.UpdatedAt {
			g.communityLibrary.UpsertDraft(draft)
		}
	}
	g.saveCommunityLibrary()
	return nil
}

func (g *Game) syncLocalDrafts() {
	for _, draft := range g.communityLibrary.Drafts {
		if raw, err := json.Marshal(draft); err == nil {
			syncCommunityDraft(string(raw))
		}
	}
}

func (g *Game) filteredCommunityDraftIndexes() []int {
	query := strings.ToLower(strings.TrimSpace(g.artSearch))
	indexes := make([]int, 0, len(g.communityLibrary.Drafts))
	for i, draft := range g.communityLibrary.Drafts {
		if query == "" || strings.Contains(strings.ToLower(draft.Title), query) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func (g *Game) importCommunityPack(raw string) error {
	var pack editorPack
	if err := json.Unmarshal([]byte(raw), &pack); err != nil {
		return err
	}
	if len(pack.Levels) == 0 {
		return fmt.Errorf("no paired levels were found")
	}
	imported := 0
	for _, puzzle := range pack.Levels {
		if puzzle == nil {
			continue
		}
		if err := puzzle.ParseSolution(); err != nil {
			continue
		}
		id := newLocalID(fmt.Sprintf("import-%d", imported+1))
		puzzle.ID = id
		draft := community.NewDraft(id, puzzle)
		if err := draft.ValidateForSave(); err != nil {
			continue
		}
		g.communityLibrary.UpsertDraft(draft)
		imported++
	}
	if imported == 0 {
		return fmt.Errorf("no valid levels were found")
	}
	if !g.saveCommunityLibrary() {
		return fmt.Errorf("local storage is unavailable")
	}
	g.communityView = communityMyArt
	g.communityPage = 0
	g.showCommunityNotice(fmt.Sprintf("imported %d level(s)", imported))
	return nil
}

func (g *Game) loadCommunityImportPreview(raw string) error {
	var pack editorPack
	if err := json.Unmarshal([]byte(raw), &pack); err != nil {
		return err
	}
	if len(pack.Levels) == 0 {
		return fmt.Errorf("no paired levels were found")
	}
	for _, puzzle := range pack.Levels {
		if puzzle == nil || parsePuzzle(puzzle) != nil {
			return fmt.Errorf("import contains invalid artwork")
		}
	}
	g.communityImportRaw = raw
	g.communityImportPack = pack
	g.communityView = communityImportPreview
	return nil
}

func (g *Game) queueCommunityDraftPublish(index int) {
	if index < 0 || index >= len(g.communityLibrary.Drafts) {
		return
	}
	draft := g.communityLibrary.Drafts[index]
	if draft.Status == community.LevelPublishedStatus {
		g.showPublishedManagementNotice("art")
		return
	}
	if err := draft.ValidateForPublish(); err != nil {
		g.showCommunityNotice(err.Error())
		return
	}
	if !communitySignedIn() {
		g.communityView = communitySignIn
		g.showCommunityNotice("sign in to publish")
		return
	}
	g.publishDraftID = draft.ID
	g.publishTitle = draft.Title
	g.publishDescription = draft.Description
	g.publishTags = strings.Join(draft.Tags, ", ")
	g.publishSubmitOfficial = false
	g.publishRightsConfirmed = false
	g.publishPreviewRaw = communityQuestionCover()
	g.publishField = 0
	g.communityView = communityPublishSetup
}

func (g *Game) showPublishedManagementNotice(kind string) {
	g.showCommunityNotice("to change or unpublish this " + kind + ", use the Published tab")
}

func (g *Game) openPublishedEditor(slot int) {
	if slot < 0 || slot >= len(g.communityPublished) {
		return
	}
	item := g.communityPublished[slot]
	g.publishedEditIndex = slot
	g.publishedEditKind = item.Kind
	g.publishedEditID = item.ID
	g.publishedEditTitle = item.Title
	g.publishedEditDescription = item.Description
	g.publishedEditField = 0
	g.publishedEditLevels = append([]community.LevelVersion(nil), item.Levels...)
	g.communityView = communityPublishedEdit
}

func (g *Game) openPublishedPackAdd() {
	if g.publishedEditKind != "pack" {
		return
	}
	if len(g.publishedEditLevels) >= community.MaxPackItems {
		g.showCommunityNotice(fmt.Sprintf("packs can have up to %d levels", community.MaxPackItems))
		return
	}
	g.communityPage = 0
	g.communityView = communityPublishedPackAdd
}

func (g *Game) applyPublishedEdit() {
	title := strings.TrimSpace(g.publishedEditTitle)
	if title == "" {
		g.showCommunityNotice("name is required")
		return
	}
	description := strings.TrimSpace(g.publishedEditDescription)
	levelsRaw := ""
	if g.publishedEditKind == "pack" {
		if len(g.publishedEditLevels) == 0 {
			g.showCommunityNotice("pack needs at least one level")
			return
		}
		raw, err := json.Marshal(g.publishedEditLevelsPayload())
		if err != nil {
			g.showCommunityNotice("could not update pack contents")
			return
		}
		levelsRaw = string(raw)
	}
	if !updateCommunityPublishedItem(g.publishedEditKind, g.publishedEditID, title, description, levelsRaw) {
		g.showCommunityNotice("published edits are available in the web build")
		return
	}
	if g.publishedEditIndex >= 0 && g.publishedEditIndex < len(g.communityPublished) {
		g.communityPublished[g.publishedEditIndex].Title = title
		g.communityPublished[g.publishedEditIndex].Description = description
		if g.publishedEditKind == "pack" {
			g.communityPublished[g.publishedEditIndex].Levels = append([]community.LevelVersion(nil), g.publishedEditLevels...)
		}
	}
	g.showCommunityNotice("applying published changes")
}

func (g *Game) publishedEditLevelsPayload() []community.LevelDraft {
	levels := make([]community.LevelDraft, 0, len(g.publishedEditLevels))
	for _, level := range g.publishedEditLevels {
		if level.Puzzle == nil {
			continue
		}
		localID := strings.TrimSpace(level.LocalID)
		if localID == "" {
			localID = strings.TrimSpace(level.LevelID)
		}
		levels = append(levels, community.LevelDraft{
			ID:          localID,
			Title:       level.Title,
			Description: level.Description,
			Tags:        append([]string(nil), level.Tags...),
			Puzzle:      level.Puzzle,
		})
	}
	return levels
}

func (g *Game) removePublishedEditPackLevel(index int) {
	if index < 0 || index >= len(g.publishedEditLevels) {
		return
	}
	if len(g.publishedEditLevels) == 1 {
		g.showCommunityNotice("pack needs at least one level")
		return
	}
	g.publishedEditLevels = append(g.publishedEditLevels[:index], g.publishedEditLevels[index+1:]...)
}

func (g *Game) addPublishedEditPackLevel() {
	if len(g.publishedEditLevels) >= community.MaxPackItems {
		g.showCommunityNotice(fmt.Sprintf("packs can have up to %d levels", community.MaxPackItems))
		return
	}
	g.openPublishedPackAdd()
}

func (g *Game) addPublishedEditPackDraft(index int) {
	if index < 0 || index >= len(g.communityLibrary.Drafts) {
		return
	}
	if len(g.publishedEditLevels) >= community.MaxPackItems {
		g.showCommunityNotice(fmt.Sprintf("packs can have up to %d levels", community.MaxPackItems))
		return
	}
	used := make(map[string]bool, len(g.publishedEditLevels))
	for _, level := range g.publishedEditLevels {
		used[level.LocalID] = true
		used[level.LevelID] = true
	}
	draft := g.communityLibrary.Drafts[index]
	if draft.Puzzle == nil || used[draft.ID] {
		g.showCommunityNotice("that art is already in this pack")
		return
	}
	g.publishedEditLevels = append(g.publishedEditLevels, community.LevelVersion{
		LevelID:     draft.ID,
		LocalID:     draft.ID,
		Title:       draft.Title,
		Description: draft.Description,
		Tags:        append([]string(nil), draft.Tags...),
		Puzzle:      draft.Puzzle,
	})
	g.communityView = communityPublishedEdit
}

func (g *Game) submitCommunityDraftPublish() {
	draft, ok := g.communityLibrary.Draft(g.publishDraftID)
	if !ok {
		return
	}
	title := strings.TrimSpace(g.publishTitle)
	if title == "" {
		g.showCommunityNotice("name is required")
		return
	}
	if g.publishSubmitOfficial && !g.publishRightsConfirmed {
		g.showCommunityNotice("confirm your rights first")
		return
	}
	draft.Title = title
	draft.Description = strings.TrimSpace(g.publishDescription)
	draft.Tags = nil
	for _, tag := range strings.Split(g.publishTags, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" && len(draft.Tags) < 8 {
			draft.Tags = append(draft.Tags, tag)
		}
	}
	g.saveCommunityLibrary()
	g.pendingPublishID = draft.ID
	g.pendingPublishAt = time.Now().Add(100 * time.Millisecond)
	g.publishAwaitingID = draft.ID
	g.showCommunityNotice("publishing " + draft.Title + "...")
}

func (g *Game) publishCommunityDraft(id string) {
	draft, ok := g.communityLibrary.Draft(id)
	if !ok {
		return
	}
	raw, err := json.Marshal(draft)
	preview := ""
	if len(g.publishPreviewRaw) > 0 {
		if encoded, encodeErr := json.Marshal(g.publishPreviewRaw); encodeErr == nil {
			preview = string(encoded)
		}
	}
	if err != nil || !requestCommunityPublish(string(raw), g.publishSubmitOfficial, g.publishRightsConfirmed, preview) {
		g.showCommunityNotice("publishing is available in the web build")
	}
}

func (g *Game) deleteCommunityDraft(index int) {
	if index < 0 || index >= len(g.communityLibrary.Drafts) {
		return
	}
	draft := g.communityLibrary.Drafts[index]
	for _, pack := range g.communityLibrary.Packs {
		for _, item := range pack.Items {
			if item.LevelID == draft.ID {
				g.showCommunityNotice("remove this art from its pack first")
				return
			}
		}
	}
	if !g.confirmCommunityDelete("art", draft.ID, draft.Title) {
		return
	}
	g.communityLibrary.Drafts = append(g.communityLibrary.Drafts[:index], g.communityLibrary.Drafts[index+1:]...)
	g.saveCommunityLibrary()
	deleteCommunityCloudDraft(draft.ID)
	if g.communityPage > 0 && g.communityPage*communityDraftsPerPage >= len(g.communityLibrary.Drafts) {
		g.communityPage--
	}
	g.showCommunityNotice("art deleted")
}

func (g *Game) deleteCommunityPack(index int) {
	if index < 0 || index >= len(g.communityLibrary.Packs) {
		return
	}
	pack := g.communityLibrary.Packs[index]
	if !g.confirmCommunityDelete("pack", pack.ID, pack.Title) {
		return
	}
	g.communityLibrary.Packs = append(g.communityLibrary.Packs[:index], g.communityLibrary.Packs[index+1:]...)
	g.saveCommunityLibrary()
	g.showCommunityNotice("pack deleted")
}

func (g *Game) confirmCommunityDelete(kind, id, title string) bool {
	now := time.Now()
	if g.pendingDeleteKind == kind && g.pendingDeleteID == id && now.Before(g.pendingDeleteUntil) {
		g.pendingDeleteKind = ""
		g.pendingDeleteID = ""
		return true
	}
	g.pendingDeleteKind = kind
	g.pendingDeleteID = id
	g.pendingDeleteUntil = now.Add(4 * time.Second)
	g.showCommunityNotice("click x again to delete " + title)
	return false
}

func (g *Game) markCommunityDraftPublished(id string) {
	for i := range g.communityLibrary.Drafts {
		if g.communityLibrary.Drafts[i].ID != id {
			continue
		}
		g.communityLibrary.Drafts[i].Status = community.LevelPublishedStatus
		g.communityLibrary.Drafts[i].Visibility = community.VisibilityPublic
		g.communityLibrary.Drafts[i].Version++
		g.communityLibrary.Drafts[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		g.saveCommunityLibrary()
		g.publishAwaitingID = ""
		g.communityView = communityMyArt
		return
	}
}

func (g *Game) markCommunityItemUnpublished(kind, id string) {
	if kind == "" || id == "" {
		return
	}
	changed := false
	switch kind {
	case "art":
		localID := id
		for _, item := range g.communityPublished {
			if item.Kind == kind && item.ID == id && item.LocalID != "" {
				localID = item.LocalID
				break
			}
		}
		for i := range g.communityLibrary.Drafts {
			if g.communityLibrary.Drafts[i].ID != localID {
				continue
			}
			oldID := g.communityLibrary.Drafts[i].ID
			newID := newLocalID("level")
			g.communityLibrary.Drafts[i].ID = newID
			if g.communityLibrary.Drafts[i].Puzzle != nil {
				g.communityLibrary.Drafts[i].Puzzle.ID = newID
			}
			g.communityLibrary.Drafts[i].Status = community.LevelDraftStatus
			g.communityLibrary.Drafts[i].Visibility = community.VisibilityDraft
			g.communityLibrary.Drafts[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			for packIndex := range g.communityLibrary.Packs {
				for itemIndex := range g.communityLibrary.Packs[packIndex].Items {
					if g.communityLibrary.Packs[packIndex].Items[itemIndex].LevelID == oldID {
						g.communityLibrary.Packs[packIndex].Items[itemIndex].LevelID = newID
					}
				}
			}
			changed = true
			break
		}
	case "pack":
		for i := range g.communityLibrary.Packs {
			if g.communityLibrary.Packs[i].ID != id {
				continue
			}
			g.communityLibrary.Packs[i].Status = community.LevelDraftStatus
			g.communityLibrary.Packs[i].Visibility = community.VisibilityDraft
			g.communityLibrary.Packs[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			changed = true
			break
		}
	}
	for i := 0; i < len(g.communityPublished); i++ {
		if g.communityPublished[i].Kind == kind && g.communityPublished[i].ID == id {
			g.communityPublished = append(g.communityPublished[:i], g.communityPublished[i+1:]...)
			i--
		}
	}
	if changed {
		g.saveCommunityLibrary()
	}
	g.pendingUnpublishKind = ""
	g.pendingUnpublishID = ""
}

func (g *Game) loadCommunityCatalog(raw string) error {
	var versions []community.LevelVersion
	if err := json.Unmarshal([]byte(raw), &versions); err != nil {
		return err
	}
	if err := parseLevelVersionPuzzles(versions); err != nil {
		return err
	}
	g.communityCatalog = versions
	return nil
}

func decodeCommunityGallery(raw string) ([]community.GalleryItem, error) {
	var items []community.GalleryItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, err
	}
	if err := parseGalleryItemPuzzles(items); err != nil {
		return nil, err
	}
	return items, nil
}

func (g *Game) loadCommunityGallery(raw string) error {
	items, err := decodeCommunityGallery(raw)
	g.communityGalleryLoading = false
	if err != nil {
		return err
	}
	g.communityGallery = items
	return nil
}

func (g *Game) requestCommunityGallery(kind, sort string) {
	g.communityGalleryLoading = true
	g.communityGallery = nil
	if !requestCommunityGallery(kind, sort) {
		g.communityGalleryLoading = false
		g.showCommunityNotice("gallery is available in the web build")
	}
}

func (g *Game) loadCommunityChat(raw string) error {
	var messages []community.ChatMessage
	if err := json.Unmarshal([]byte(raw), &messages); err != nil {
		return err
	}
	for i := range messages {
		if err := parsePuzzle(messages[i].AvatarPuzzle); err != nil {
			return fmt.Errorf("chat message %d avatar: %w", i+1, err)
		}
	}
	g.communityChatMessages = messages
	return nil
}

func (g *Game) openCommunityChat(kind, id, title string, back communityView) {
	g.chatKind = kind
	g.chatID = id
	g.chatTitle = title
	g.chatDraft = ""
	g.chatReturn = back
	g.communityChatMessages = nil
	g.communityView = communityChat
	requestCommunityCreators()
	if !requestCommunityChat(kind, id) {
		g.showCommunityNotice("chat is available in the web build")
	}
}

func (g *Game) openChatAuthorProfile(authorID string) bool {
	for i := range g.communityCreators {
		if g.communityCreators[i].ID != authorID {
			continue
		}
		g.selectedCreator = i
		g.communityPage = 0
		g.communityView = communityCreatorProfile
		return true
	}
	requestCommunityCreators()
	g.showCommunityNotice("loading profile")
	return false
}

func (g *Game) sendCommunityChat() {
	body := strings.TrimSpace(g.chatDraft)
	if body == "" {
		return
	}
	if !communitySignedIn() {
		g.communityView = communitySignIn
		g.showCommunityNotice("sign in to chat")
		return
	}
	if !postCommunityChat(g.chatKind, g.chatID, body) {
		g.showCommunityNotice("chat is available in the web build")
		return
	}
	g.chatDraft = ""
}

func (g *Game) loadCommunityPublished(raw string) error {
	items, err := decodeCommunityGallery(raw)
	if err != nil {
		return err
	}
	g.communityPublished = items
	return nil
}

func (g *Game) loadCommunityCompleted(raw string) error {
	items, err := decodeCommunityGallery(raw)
	if err != nil {
		return err
	}
	for i := range items {
		items[i].Completed = true
	}
	g.communityCompleted = items
	for _, item := range items {
		g.markCommunityLevelCompleted(item.ID)
	}
	return nil
}

func (g *Game) communityLevelCompleted(levelID string, cloudCompleted bool) bool {
	return cloudCompleted || (levelID != "" && g.bestTimes[levelID] > 0)
}

func (g *Game) markCommunityLevelCompleted(levelID string) {
	if levelID == "" {
		return
	}
	for i := range g.communityGallery {
		if g.communityGallery[i].ID == levelID {
			g.communityGallery[i].Completed = true
		}
		for j := range g.communityGallery[i].Levels {
			if g.communityGallery[i].Levels[j].LevelID == levelID {
				g.communityGallery[i].Levels[j].Completed = true
			}
		}
	}
	for i := range g.communityCreators {
		for j := range g.communityCreators[i].Levels {
			if g.communityCreators[i].Levels[j].LevelID == levelID {
				g.communityCreators[i].Levels[j].Completed = true
			}
		}
		for j := range g.communityCreators[i].Featured {
			if g.communityCreators[i].Featured[j].ID == levelID {
				g.communityCreators[i].Featured[j].Completed = true
			}
			for k := range g.communityCreators[i].Featured[j].Levels {
				if g.communityCreators[i].Featured[j].Levels[k].LevelID == levelID {
					g.communityCreators[i].Featured[j].Levels[k].Completed = true
				}
			}
		}
	}
	for i := range g.communityCompleted {
		if g.communityCompleted[i].ID == levelID {
			g.communityCompleted[i].Completed = true
		}
	}
}

func (g *Game) playGalleryLevel(index int) {
	if index < 0 || index >= len(g.communityGallery) || g.communityGallery[index].Puzzle == nil {
		return
	}
	item := g.communityGallery[index]
	g.playCommunityLevel(item.Puzzle, item.ID, communityBrowse)
}

func (g *Game) playGalleryPackLevel(index int) {
	if g.selectedGallery < 0 || g.selectedGallery >= len(g.communityGallery) {
		return
	}
	levels := g.communityGallery[g.selectedGallery].Levels
	if index < 0 || index >= len(levels) || levels[index].Puzzle == nil {
		return
	}
	g.playCommunityLevel(levels[index].Puzzle, levels[index].LevelID, communityGalleryPack)
}

func (g *Game) loadCommunityCreators(raw string) error {
	var creators []community.CreatorProfile
	if err := json.Unmarshal([]byte(raw), &creators); err != nil {
		return err
	}
	for i := range creators {
		if err := parsePuzzle(creators[i].AvatarPuzzle); err != nil {
			return fmt.Errorf("creator %d avatar: %w", i+1, err)
		}
		if err := parseLevelVersionPuzzles(creators[i].Levels); err != nil {
			return fmt.Errorf("creator %d levels: %w", i+1, err)
		}
		if err := parseGalleryItemPuzzles(creators[i].Featured); err != nil {
			return fmt.Errorf("creator %d featured work: %w", i+1, err)
		}
	}
	g.communityCreators = creators
	return nil
}

func parsePuzzle(puzzle *nonogram.Puzzle) error {
	if puzzle == nil {
		return nil
	}
	return puzzle.ParseSolution()
}

func parseLevelVersionPuzzles(levels []community.LevelVersion) error {
	for i := range levels {
		if err := parsePuzzle(levels[i].Puzzle); err != nil {
			return fmt.Errorf("level %d: %w", i+1, err)
		}
	}
	return nil
}

func parseGalleryItemPuzzles(items []community.GalleryItem) error {
	for i := range items {
		if err := parsePuzzle(items[i].AvatarPuzzle); err != nil {
			return fmt.Errorf("gallery item %d avatar: %w", i+1, err)
		}
		if err := parsePuzzle(items[i].Puzzle); err != nil {
			return fmt.Errorf("gallery item %d puzzle: %w", i+1, err)
		}
		if err := parseLevelVersionPuzzles(items[i].Levels); err != nil {
			return fmt.Errorf("gallery item %d: %w", i+1, err)
		}
	}
	return nil
}

func (g *Game) filteredCommunityCreatorIndexes() []int {
	query := strings.ToLower(strings.TrimSpace(g.creatorSearch))
	indexes := make([]int, 0, len(g.communityCreators))
	for i, creator := range g.communityCreators {
		if query == "" ||
			strings.Contains(strings.ToLower(creator.DisplayName), query) ||
			strings.Contains(strings.ToLower(creator.Bio), query) ||
			strings.Contains(strings.ToLower(creator.Social), query) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func (g *Game) playCreatorLevel(index int) {
	if g.selectedCreator < 0 || g.selectedCreator >= len(g.communityCreators) {
		return
	}
	levels := g.communityCreators[g.selectedCreator].Levels
	if index < 0 || index >= len(levels) || levels[index].Puzzle == nil {
		return
	}
	g.playCommunityLevel(levels[index].Puzzle, levels[index].LevelID, communityCreatorProfile)
}

func (g *Game) playCommunityLevel(source *nonogram.Puzzle, levelID string, returnView communityView) {
	recordCommunityPlay(levelID, false)
	puzzle := *source
	puzzle.ID = levelID
	g.activeCommunityPack = ""
	g.communityPlayReturn = returnView
	g.loadCommunityPuzzle(&puzzle)
}

func (g *Game) playCommunityVersion(index int) {
	if index < 0 || index >= len(g.communityCatalog) || g.communityCatalog[index].Puzzle == nil {
		return
	}
	p := g.communityCatalog[index].Puzzle
	g.activeCommunityPack = ""
	g.communityPlayReturn = communityBrowse
	g.loadCommunityPuzzle(p)
}

func (g *Game) loadCommunityPuzzle(p *nonogram.Puzzle) {
	g.puzzle = p
	g.board = nonogram.NewBoard(p.Width, p.Height)
	g.rowClues = nonogram.RowClues(p.Solution)
	g.colClues = nonogram.ColumnClues(p.Solution)
	g.skeleton = nil
	g.reveal = nil
	g.skeletonPixels = pixelsFromRaw(p.SkeletonRaw)
	g.revealPixels = pixelsFromRaw(p.RevealRaw)
	g.undoStack = nil
	g.startTime = time.Now()
	g.editorPreview = false
	g.communityPreview = true
	g.mode = screenPuzzle
}

func pixelsFromRaw(raw [][]string) [][]assets.PixelCell {
	result := make([][]assets.PixelCell, len(raw))
	for y, row := range raw {
		result[y] = make([]assets.PixelCell, len(row))
		for x, value := range row {
			if c, ok := parseEditorHexColor(value); ok && c.A > 0 {
				result[y][x] = assets.PixelCell{Color: c, Visible: true}
			}
		}
	}
	return result
}

func (g *Game) openNewPackSetup() {
	items := make([]community.PackItem, 0, len(g.packSelection))
	for _, draft := range g.communityLibrary.Drafts {
		if !g.packSelection[draft.ID] {
			continue
		}
		items = append(items, community.PackItem{LevelID: draft.ID, Position: len(items)})
		if len(items) == 20 {
			break
		}
	}
	if len(items) == 0 {
		g.showCommunityNotice("create at least one level first")
		return
	}
	g.packSetupID = ""
	g.packSetupTitle = fmt.Sprintf("My Pack %d", len(g.communityLibrary.Packs)+1)
	g.packSetupDescription = ""
	g.packSetupItems = items
	g.packSetupPreview = -1
	g.packSetupPreviewRaw = communityQuestionCover()
	g.packSetupField = 0
	g.packSelection = nil
	g.communityView = communityPackSetup
}

func (g *Game) openPackBuilder() {
	g.packSelection = make(map[string]bool)
	g.communityPage = 0
	g.communityView = communityPackBuild
}

func (g *Game) togglePackDraft(index int) {
	if index < 0 || index >= len(g.communityLibrary.Drafts) {
		return
	}
	id := g.communityLibrary.Drafts[index].ID
	if !g.packSelection[id] && len(g.packSelection) >= community.MaxPackItems {
		g.showCommunityNotice(fmt.Sprintf("packs can contain up to %d levels", community.MaxPackItems))
		return
	}
	if g.packSelection[id] {
		delete(g.packSelection, id)
	} else {
		g.packSelection[id] = true
	}
}

func (g *Game) queueLocalPackPublish(index int) {
	if index < 0 || index >= len(g.communityLibrary.Packs) {
		return
	}
	pack := g.communityLibrary.Packs[index]
	if pack.Status == community.LevelPublishedStatus {
		g.showPublishedManagementNotice("pack")
		return
	}
	g.packSetupID = pack.ID
	g.packSetupTitle = pack.Title
	g.packSetupDescription = pack.Description
	g.packSetupItems = append([]community.PackItem(nil), pack.Items...)
	g.packSetupPreview = -1
	g.packSetupPreviewRaw = communityQuestionCover()
	g.packSetupField = 0
	g.communityView = communityPackSetup
}

func (g *Game) savePackSetup(publish bool) {
	title := strings.TrimSpace(g.packSetupTitle)
	if title == "" {
		g.showCommunityNotice("pack title is required")
		return
	}
	items := append([]community.PackItem(nil), g.packSetupItems...)
	if g.packSetupPreview > 0 && g.packSetupPreview < len(items) {
		cover := items[g.packSetupPreview]
		items = append([]community.PackItem{cover}, append(items[:g.packSetupPreview], items[g.packSetupPreview+1:]...)...)
	}
	for i := range items {
		items[i].Position = i
	}
	var pack *community.Pack
	for i := range g.communityLibrary.Packs {
		if g.communityLibrary.Packs[i].ID == g.packSetupID {
			pack = &g.communityLibrary.Packs[i]
			break
		}
	}
	if pack == nil {
		created := community.Pack{ID: newLocalID("pack"), Visibility: community.VisibilityDraft, Status: community.LevelDraftStatus, Version: 1, Items: items}
		g.communityLibrary.Packs = append([]community.Pack{created}, g.communityLibrary.Packs...)
		pack = &g.communityLibrary.Packs[0]
	}
	pack.Title = title
	pack.Description = strings.TrimSpace(g.packSetupDescription)
	pack.Items = items
	pack.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := pack.Validate(); err != nil {
		g.showCommunityNotice(err.Error())
		return
	}
	g.saveCommunityLibrary()
	if !publish {
		g.communityView = communityPacks
		g.showCommunityNotice("pack saved as draft")
		return
	}
	if !communitySignedIn() {
		g.communityView = communitySignIn
		g.showCommunityNotice("sign in to publish")
		return
	}
	g.pendingPackPublishID = pack.ID
	g.pendingPackPublishAt = time.Now().Add(100 * time.Millisecond)
	g.packPublishAwaitingID = pack.ID
	g.showCommunityNotice("publishing " + pack.Title + "...")
}

func (g *Game) publishLocalPack(id string) {
	var pack *community.Pack
	for i := range g.communityLibrary.Packs {
		if g.communityLibrary.Packs[i].ID == id {
			pack = &g.communityLibrary.Packs[i]
			break
		}
	}
	if pack == nil {
		return
	}
	levels := make([]community.LevelDraft, 0, len(pack.Items))
	for _, item := range pack.Items {
		if draft, ok := g.communityLibrary.Draft(item.LevelID); ok {
			levels = append(levels, *draft)
		}
	}
	raw, err := json.Marshal(struct {
		Pack   *community.Pack        `json:"pack"`
		Levels []community.LevelDraft `json:"levels"`
	}{Pack: pack, Levels: levels})
	preview := ""
	if len(g.packSetupPreviewRaw) > 0 {
		if encoded, encodeErr := json.Marshal(g.packSetupPreviewRaw); encodeErr == nil {
			preview = string(encoded)
		}
	}
	if err != nil || !requestCommunityPackPublish(string(raw), preview) {
		g.showCommunityNotice("pack publishing is available in the web build")
	}
}

func (g *Game) markCommunityPackPublished(id string) {
	for i := range g.communityLibrary.Packs {
		if g.communityLibrary.Packs[i].ID != id {
			continue
		}
		g.communityLibrary.Packs[i].Status = community.LevelPublishedStatus
		g.communityLibrary.Packs[i].Visibility = community.VisibilityPublic
		g.communityLibrary.Packs[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		g.saveCommunityLibrary()
		g.packPublishAwaitingID = ""
		g.communityView = communityPacks
		return
	}
}

func (g *Game) playLocalPack(index int) {
	if index < 0 || index >= len(g.communityLibrary.Packs) {
		return
	}
	pack := &g.communityLibrary.Packs[index]
	completed := make(map[string]bool, len(pack.Progress.CompletedLevelIDs))
	for _, id := range pack.Progress.CompletedLevelIDs {
		completed[id] = true
	}
	for _, item := range pack.Items {
		if completed[item.LevelID] {
			continue
		}
		if draft, ok := g.communityLibrary.Draft(item.LevelID); ok && draft.Puzzle != nil {
			g.activeCommunityPack = pack.ID
			g.loadCommunityPuzzle(draft.Puzzle)
			return
		}
	}
	g.showCommunityNotice("pack complete")
}

func (g *Game) completeCommunityPackLevel(levelID string) {
	for i := range g.communityLibrary.Packs {
		pack := &g.communityLibrary.Packs[i]
		if pack.ID != g.activeCommunityPack {
			continue
		}
		for _, id := range pack.Progress.CompletedLevelIDs {
			if id == levelID {
				return
			}
		}
		pack.Progress.CompletedLevelIDs = append(pack.Progress.CompletedLevelIDs, levelID)
		g.saveCommunityLibrary()
		return
	}
}

func newLocalID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func (g *Game) showCommunityNotice(text string) {
	g.communityNotice = text
	g.communityNoticeUntil = time.Now().Add(2200 * time.Millisecond)
}

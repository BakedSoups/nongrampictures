package game

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/BakedSoups/community_nongrams/internal/community"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	communityDraftsPerPage     = 3
	communityCatalogPerPage    = 4
	communityPackDraftsPerPage = 4
	communityCreatorsPerPage   = 3
)

func (g *Game) drawCommunity(screen *ebiten.Image) {
	drawCommunityBackdrop(screen)
	drawScaledTextCentered(screen, "GLOBAL COMMUNITY", rect{x: 76, y: 42, w: 388, h: 52}, 1.65, colInk)
	if g.communityView == communityHome {
		g.drawCommunityAccount(screen)
	}
	switch g.communityView {
	case communityBrowse:
		g.drawCommunityBrowse(screen)
	case communityPacks:
		g.drawCommunityPacks(screen)
	case communityCreate:
		g.drawCommunityCreate(screen)
	case communityMyArt:
		g.drawCommunityMyArt(screen)
	case communitySignIn:
		g.drawCommunitySignIn(screen)
	case communityPackBuild:
		g.drawCommunityPackBuilder(screen)
	case communityCreators:
		g.drawCommunityCreators(screen)
	case communityCreatorProfile:
		g.drawCommunityCreatorProfile(screen)
	case communityGalleryPack:
		g.drawCommunityGalleryPack(screen)
	case communityImportHelp:
		g.drawCommunityImportHelp(screen)
	case communityPublished:
		g.drawCommunityPublished(screen)
	case communityPublishedEdit:
		g.drawCommunityPublishedEdit(screen)
	case communityPublishedPackAdd:
		g.drawCommunityPublishedPackAdd(screen)
	case communityPublishSetup:
		g.drawCommunityPublishSetup(screen)
	case communityImportPreview:
		g.drawCommunityImportPreview(screen)
	case communityImportSetup:
		g.drawCommunityImportSetup(screen)
	case communityNewArtSetup:
		g.drawCommunityNewArtSetup(screen)
	case communityPackSetup:
		g.drawCommunityPackSetup(screen)
	case communityChat:
		g.drawCommunityChat(screen)
	default:
		g.drawCommunityHome(screen)
	}
	drawButton(screen, communityBackButton(), "back")
	if time.Now().Before(g.communityNoticeUntil) {
		drawNoticePopup(screen, g.communityNotice, 610)
	}
}

func (g *Game) drawCommunityBrowse(screen *ebiten.Image) {
	drawCenteredText(screen, "GALLERY", rect{x: 100, y: 190, w: 340, h: 28}, colInk)
	drawSelectedButton(screen, communityGalleryAllButton(), "All", g.galleryKind == "all")
	drawSelectedButton(screen, communityGalleryArtButton(), "Art", g.galleryKind == "art")
	drawSelectedButton(screen, communityGalleryPacksButton(), "Packs", g.galleryKind == "pack")
	drawSelectedButton(screen, communityGallerySortButton(), communityGallerySortLabel(g.gallerySort), g.gallerySortOpen)
	if g.communityGalleryLoading {
		drawCenteredText(screen, "loading art...", rect{x: 60, y: 346, w: 420, h: 40}, colMuted)
		if g.gallerySortOpen {
			drawCommunitySortMenu(screen, g.gallerySort)
		}
		return
	}
	if len(g.communityGallery) == 0 {
		drawCenteredText(screen, "No published work yet", rect{x: 60, y: 346, w: 420, h: 40}, colMuted)
		if g.gallerySortOpen {
			drawCommunitySortMenu(screen, g.gallerySort)
		}
		return
	}
	start := g.communityPage * communityCatalogPerPage
	for slot := 0; slot < communityCatalogPerPage && start+slot < len(g.communityGallery); slot++ {
		item := g.communityGallery[start+slot]
		r := communityGalleryCard(slot)
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		drawCommunityArtThumbnail(screen, g.communityGalleryPreviewPixels(item), rect{x: r.x + 8, y: r.y + 8, w: 54, h: 54})
		drawText(screen, truncateText(item.Title, 28), int(r.x+72), int(r.y+18), colInk)
		description := strings.TrimSpace(item.Description)
		if description == "" && item.Kind == "pack" {
			description = fmt.Sprintf("%d levels", len(item.Levels))
		}
		drawWrappedText(screen, description, int(r.x+72), int(r.y+38), 30, 2, colMuted)
		creator := item.CreatorName
		if len(creator) > 15 {
			creator = creator[:15]
		}
		avatar := defaultCommunityProfilePixels
		if item.AvatarPuzzle != nil {
			avatar = item.AvatarPuzzle.RevealRaw
		}
		avatarRect := rect{x: r.x + r.w - 42, y: r.y + 8, w: 32, h: 26}
		drawText(screen, fmt.Sprintf("%s  %d plays", creator, item.Plays), int(r.x+72), int(r.y+76), colMuted)
		drawCommunityArtThumbnail(screen, avatar, avatarRect)
		drawButton(screen, communityGalleryOpenButton(slot), map[bool]string{true: "open", false: "play"}[item.Kind == "pack"])
		drawChatIconButton(screen, communityGalleryChatButton(slot))
		g.drawThumbLikeButton(screen, communityGalleryLikeButton(slot), item.Likes)
	}
	if g.gallerySortOpen {
		drawCommunitySortMenu(screen, g.gallerySort)
	}
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*communityCatalogPerPage < len(g.communityGallery) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func drawSelectedButton(screen *ebiten.Image, r rect, label string, selected bool) {
	if selected {
		drawRounded(screen, inset(r, -3), 9, colAccent)
	}
	drawButton(screen, r, label)
}

func drawCommunitySortMenu(screen *ebiten.Image, selected string) {
	menu := communityGallerySortMenu()
	drawRounded(screen, menu, 7, colWhite)
	drawRectOutline(screen, menu, 2, colGridHeavy)
	drawCommunitySortMenuOption(screen, communityGalleryNewButton(), "sort by new", selected == "new")
	drawCommunitySortMenuOption(screen, communityGalleryPlayedButton(), "most played", selected == "played")
	drawCommunitySortMenuOption(screen, communityGalleryTopButton(), "top rated", selected == "top")
}

func drawCommunitySortMenuOption(screen *ebiten.Image, r rect, label string, selected bool) {
	if selected {
		drawRounded(screen, inset(r, 4), 5, colPanel)
	}
	drawCenteredText(screen, label, r, colInk)
}

func (g *Game) drawThumbLikeButton(screen *ebiten.Image, r rect, likes int) {
	drawButton(screen, r, "")
	if g.icons != nil && g.icons.Thumbsup != nil {
		drawPixelIconImageSized(screen, g.icons.Thumbsup, rect{x: r.x + 6, y: r.y + 6, w: 16, h: 16}, 16)
	}
	drawText(screen, fmt.Sprintf("%d", likes), int(r.x+28), int(r.y+20), colInk)
}

func drawChatIconButton(screen *ebiten.Image, r rect) {
	drawButton(screen, r, "")
	drawPixelChatIcon(screen, inset(r, 8), colInk)
}

func drawPixelChatIcon(screen *ebiten.Image, r rect, c color.Color) {
	x := float32(r.x)
	y := float32(r.y)
	w := float32(r.w)
	h := float32(r.h)
	vector.DrawFilledRect(screen, x+w*0.18, y+h*0.20, w*0.64, h*0.46, c, false)
	vector.DrawFilledRect(screen, x+w*0.26, y+h*0.66, w*0.18, h*0.16, c, false)
	vector.DrawFilledRect(screen, x+w*0.30, y+h*0.30, w*0.10, h*0.12, colWhite, false)
	vector.DrawFilledRect(screen, x+w*0.46, y+h*0.30, w*0.10, h*0.12, colWhite, false)
	vector.DrawFilledRect(screen, x+w*0.62, y+h*0.30, w*0.10, h*0.12, colWhite, false)
}

func communityGallerySortLabel(sort string) string {
	switch sort {
	case "played":
		return "most played"
	case "top":
		return "top rated"
	default:
		return "sort by new"
	}
}

func (g *Game) communityGalleryPreviewPixels(item community.GalleryItem) [][]string {
	if item.Kind == "art" && g.communityLevelCompleted(item.ID, item.Completed) && item.Puzzle != nil {
		return item.Puzzle.RevealRaw
	}
	if len(item.PreviewRaw) > 0 {
		return item.PreviewRaw
	}
	return communityQuestionCover()
}

func (g *Game) drawCommunityGalleryPack(screen *ebiten.Image) {
	if g.selectedGallery < 0 || g.selectedGallery >= len(g.communityGallery) {
		return
	}
	item := g.communityGallery[g.selectedGallery]
	drawCenteredText(screen, item.Title, rect{x: 70, y: 194, w: 400, h: 30}, colInk)
	drawCenteredText(screen, fmt.Sprintf("by %s   %d plays   %d likes", item.CreatorName, item.Plays, item.Likes), rect{x: 70, y: 224, w: 332, h: 24}, colMuted)
	if bio := strings.TrimSpace(item.CreatorBio); bio != "" {
		drawCenteredText(screen, truncateText(bio, 42), rect{x: 70, y: 248, w: 400, h: 22}, colMuted)
	}
	drawChatIconButton(screen, communityGalleryPackChatButton())
	for slot := 0; slot < 6 && slot < len(item.Levels); slot++ {
		level := item.Levels[slot]
		r := communityGalleryPackLevelButton(slot)
		drawRounded(screen, r, 5, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		preview := communityQuestionCover()
		if g.communityLevelCompleted(level.LevelID, level.Completed) && level.Puzzle != nil {
			preview = level.Puzzle.RevealRaw
		}
		drawCommunityArtThumbnail(screen, preview, rect{x: r.x + 7, y: r.y + 7, w: 58, h: 58})
		title := level.Title
		if len(title) > 12 {
			title = title[:12]
		}
		drawCenteredText(screen, title, rect{x: r.x + 70, y: r.y + 8, w: r.w - 76, h: 28}, colInk)
		drawCenteredText(screen, "play", rect{x: r.x + 70, y: r.y + 38, w: r.w - 76, h: 26}, colAccent)
	}
}

func (g *Game) drawCommunityChat(screen *ebiten.Image) {
	drawCenteredText(screen, "CHAT", rect{x: 100, y: 190, w: 340, h: 28}, colInk)
	drawCenteredText(screen, truncateText(g.chatTitle, 34), rect{x: 70, y: 218, w: 400, h: 24}, colMuted)
	panel := rect{x: 48, y: 252, w: 444, h: 288}
	drawRounded(screen, panel, 6, colWhite)
	drawRectOutline(screen, panel, 2, colGridHeavy)
	if len(g.communityChatMessages) == 0 {
		drawCenteredText(screen, "No messages yet", rect{x: 80, y: 360, w: 380, h: 30}, colMuted)
	} else {
		start := len(g.communityChatMessages) - 5
		if start < 0 {
			start = 0
		}
		y := 278
		for slot, msg := range g.communityChatMessages[start:] {
			row := communityChatMessageButton(slot)
			avatar := defaultCommunityProfilePixels
			if msg.AvatarPuzzle != nil {
				avatar = msg.AvatarPuzzle.RevealRaw
			}
			drawCommunityArtThumbnail(screen, avatar, rect{x: row.x + 8, y: row.y + 5, w: 36, h: 36})
			drawText(screen, truncateText(msg.AuthorName, 18), int(row.x+54), y, colAccent)
			drawText(screen, "view profile", int(row.x+270), y, colMuted)
			drawWrappedText(screen, msg.Body, int(row.x+54), y+22, 38, 2, colInk)
			y += 50
		}
	}
	drawPublishField(screen, communityChatInputField(), "Message", g.chatDraft, true)
	drawButton(screen, communityChatSendButton(), "send")
}

func (g *Game) drawCommunityPacks(screen *ebiten.Image) {
	drawCenteredText(screen, "MY LIBRARY", rect{x: 100, y: 190, w: 340, h: 30}, colInk)
	drawLibraryTabs(screen, false)
	drawButton(screen, communityPackCreateButton(), "Create Pack")
	for i, pack := range g.communityLibrary.Packs {
		if i >= 4 {
			break
		}
		r := communityPackRect(i)
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		drawText(screen, truncateText(pack.Title, 30), int(r.x+10), int(r.y+18), colInk)
		description := strings.TrimSpace(pack.Description)
		if description == "" {
			description = "No description"
		}
		drawText(screen, truncateText(description, 30), int(r.x+10), int(r.y+38), colMuted)
		packStatus := fmt.Sprintf("%d art", len(pack.Items))
		if string(pack.Status) == "published" {
			packStatus = "published"
		}
		if g.pendingPackPublishID == pack.ID {
			packStatus = "publishing"
		}
		drawText(screen, packStatus, int(r.x+10), int(r.y+58), colAccent)
		for art, item := range pack.Items {
			if art >= community.MaxPackItems {
				break
			}
			draft, ok := g.communityLibrary.Draft(item.LevelID)
			if !ok || draft.Puzzle == nil {
				continue
			}
			drawCommunityArtThumbnail(screen, draft.Puzzle.RevealRaw, communityPackArtPreview(i, art, len(pack.Items)))
		}
		drawButton(screen, communityPackPlayButton(i), "play")
		publishLabel := "edit"
		if pack.Status == community.LevelPublishedStatus {
			publishLabel = "published"
		}
		if g.pendingPackPublishID == pack.ID {
			publishLabel = "..."
		}
		drawButton(screen, communityPackPublishButton(i), publishLabel)
		drawButton(screen, communityPackDeleteButton(i), "x")
	}
}

func (g *Game) drawCommunityHome(screen *ebiten.Image) {
	drawButton(screen, communityLevelsButton(), "Community Gallery")
	drawButton(screen, communityMyArtButton(), "My Library")
	drawButton(screen, communityCreatorsButton(), "Creators")
}

func (g *Game) drawCommunityCreators(screen *ebiten.Image) {
	drawCenteredText(screen, "CREATORS", rect{x: 100, y: 194, w: 340, h: 30}, colInk)
	drawPublishField(screen, communityCreatorSearchField(), "Search", g.creatorSearch, g.creatorSearchActive)
	indexes := g.filteredCommunityCreatorIndexes()
	if len(indexes) == 0 {
		drawCenteredText(screen, "No creators yet", rect{x: 80, y: 350, w: 380, h: 32}, colMuted)
		return
	}
	start := g.communityPage * communityCreatorsPerPage
	for slot := 0; slot < communityCreatorsPerPage && start+slot < len(indexes); slot++ {
		creator := g.communityCreators[indexes[start+slot]]
		r := communityCreatorButton(slot)
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		avatar := defaultCommunityProfilePixels
		if creator.AvatarPuzzle != nil {
			avatar = creator.AvatarPuzzle.RevealRaw
		}
		drawCommunityArtThumbnail(screen, avatar, rect{x: r.x + 8, y: r.y + 8, w: 56, h: 56})
		name := creator.DisplayName
		if len(name) > 20 {
			name = name[:20]
		}
		drawText(screen, name, int(r.x+76), int(r.y+25), colInk)
		bio := strings.TrimSpace(creator.Bio)
		if bio == "" {
			bio = "No bio yet"
		}
		drawWrappedText(screen, bio, int(r.x+76), int(r.y+48), 28, 4, colMuted)
		for art := 0; art < 3 && art < len(creator.Levels); art++ {
			if creator.Levels[art].Puzzle != nil {
				drawCommunityArtThumbnail(screen, creator.Levels[art].Puzzle.RevealRaw, communityCreatorPreviewRect(slot, art))
			}
		}
	}
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*communityCreatorsPerPage < len(indexes) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func (g *Game) drawCommunityCreatorProfile(screen *ebiten.Image) {
	if g.selectedCreator < 0 || g.selectedCreator >= len(g.communityCreators) {
		drawCenteredText(screen, "CREATOR", rect{x: 100, y: 194, w: 340, h: 30}, colInk)
		return
	}
	creator := g.communityCreators[g.selectedCreator]
	avatar := defaultCommunityProfilePixels
	if creator.AvatarPuzzle != nil {
		avatar = creator.AvatarPuzzle.RevealRaw
	}
	drawCommunityArtThumbnail(screen, avatar, rect{x: 54, y: 206, w: 82, h: 82})
	drawText(screen, creator.DisplayName, 154, 242, colInk)
	drawText(screen, fmt.Sprintf("%d published", len(creator.Levels)), 154, 269, colMuted)
	infoY := 296
	if creator.Bio != "" {
		drawWrappedText(screen, creator.Bio, 154, infoY, 38, 3, colMuted)
		infoY += wrappedTextLineCount(creator.Bio, 38, 3)*20 + 14
	}
	socials := splitProfileSocials(creator.Social)
	for slot, social := range socials {
		if social == "" {
			continue
		}
		drawSocialChip(screen, rect{x: 154, y: float64(infoY + slot*24), w: 270, h: 20}, social)
	}
	prefsY := infoY + countProfileSocials(creator.Social)*24
	if prefsY != infoY {
		prefsY += 6
	}
	if creator.Palette != "" || creator.FavoriteColor != "" {
		drawProfilePreferenceRow(screen, 154, prefsY, creator.Palette, creator.FavoriteColor)
		prefsY += 26
	}
	contentY := communityCreatorProfileBaseContentY(creator.Social, creator.Bio, creator.Palette, creator.FavoriteColor)
	if len(creator.Featured) > 0 {
		featured := creator.Featured[0]
		headerY := int(contentY)
		drawPixelFavoriteStar(screen, 48, headerY)
		drawText(screen, "FAVORITE PICROSS", 72, headerY+6, colAccent)
		r := communityCreatorFeaturedButtonAt(float64(headerY + 30))
		drawRounded(screen, r, 5, colWhite)
		drawRectOutline(screen, r, 2, colAccent)
		if featured.Puzzle != nil {
			drawCommunityArtThumbnail(screen, featured.Puzzle.RevealRaw, rect{x: r.x + 7, y: r.y + 7, w: 58, h: 58})
		} else if len(featured.Levels) > 0 && featured.Levels[0].Puzzle != nil {
			drawCommunityArtThumbnail(screen, featured.Levels[0].Puzzle.RevealRaw, rect{x: r.x + 7, y: r.y + 7, w: 58, h: 58})
		}
		drawText(screen, featured.Title, int(r.x+76), int(r.y+23), colInk)
		likeLabel := "LIKES"
		if featured.Likes == 1 {
			likeLabel = "LIKE"
		}
		drawText(screen, fmt.Sprintf("%s   %d %s", strings.ToUpper(featured.Kind), featured.Likes, likeLabel), int(r.x+76), int(r.y+48), colMuted)
		contentY = communityCreatorProfileLevelsY(creator.Social, creator.Bio, creator.Palette, creator.FavoriteColor, true)
	}
	start := g.communityPage * 4
	for slot := 0; slot < 4 && start+slot < len(creator.Levels); slot++ {
		level := creator.Levels[start+slot]
		r := communityCreatorLevelButtonAt(slot, contentY)
		drawRounded(screen, r, 5, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		if level.Puzzle != nil {
			drawCommunityArtThumbnail(screen, level.Puzzle.RevealRaw, rect{x: r.x + 7, y: r.y + 7, w: 58, h: 58})
		}
		title := level.Title
		if len(title) > 12 {
			title = title[:12]
		}
		drawCenteredText(screen, title, rect{x: r.x + 70, y: r.y + 8, w: r.w - 76, h: 28}, colInk)
		drawCenteredText(screen, "play", rect{x: r.x + 70, y: r.y + 38, w: r.w - 76, h: 26}, colAccent)
	}
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*4 < len(creator.Levels) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func drawPixelFavoriteStar(screen *ebiten.Image, x, y int) {
	vector.DrawFilledRect(screen, float32(x+8), float32(y), 6, 22, colAccent, false)
	vector.DrawFilledRect(screen, float32(x), float32(y+8), 22, 6, colAccent, false)
	vector.DrawFilledRect(screen, float32(x+4), float32(y+4), 14, 14, colAccent, false)
}

func (g *Game) drawCommunityAccount(screen *ebiten.Image) {
	drawButton(screen, communityAccountButton(), communityAccountLabel())
	pixels := g.profilePixelsOrDefault()
	if !communitySignedIn() {
		pixels = defaultCommunityProfilePixels
	}
	registerButtonRect(communityProfileBadgeButton())
	drawCommunityArtThumbnail(screen, pixels, communityProfileBadgeButton())
}

func (g *Game) profilePixelsOrDefault() [][]string {
	if communitySignedIn() {
		if saved := g.profileArt.rawPixels(editorLayerAfter); len(saved) > 0 {
			for _, row := range saved {
				for _, value := range row {
					if c, ok := parseEditorHexColor(value); ok && c.A > 0 {
						return saved
					}
				}
			}
		}
	}
	return defaultCommunityProfilePixels
}

func drawSupportedSocialIcons(screen *ebiten.Image) {
	supported := []string{"github", "x", "instagram", "tiktok", "youtube", "twitch", "bluesky", "threads", "mastodon", "linkedin"}
	for index, platform := range supported {
		r := rect{x: 134 + float64(index)*27, y: 584, w: 22, h: 20}
		drawRounded(screen, r, 3, colWhite)
		drawRectOutline(screen, r, 1, colGridHeavy)
		drawSocialIcon(screen, r, platform+":")
	}
}

func drawProfilePreferenceRow(screen *ebiten.Image, x, y int, palette, favoriteColor string) {
	if palette != "" {
		drawText(screen, "favorite palette", x, y+16, colMuted)
		slots := profilePaletteSlots(palette)
		for index, value := range slots {
			if c, ok := parseEditorHexColor(value); ok {
				sw := rect{x: float64(x + 148 + index*18), y: float64(y + 1), w: 14, h: 14}
				drawRounded(screen, sw, 2, c)
				drawRectOutline(screen, sw, 1, colGridHeavy)
			}
		}
	}
	if favoriteColor != "" {
		drawText(screen, "favorite color", x, y+40, colMuted)
		if c, ok := parseEditorHexColor(favoriteColor); ok {
			sw := rect{x: float64(x + 126), y: float64(y + 25), w: 18, h: 18}
			drawRounded(screen, sw, 3, c)
			drawRectOutline(screen, sw, 2, colGridHeavy)
		}
	}
}

func (g *Game) drawProfilePreferenceButtons(screen *ebiten.Image, palette, favoriteColor string) {
	drawText(screen, "Palette", int(communityAccountPaletteOptionButton(0).x), 404, colMuted)
	slots := profilePaletteSlots(palette)
	for index, value := range slots {
		drawPaletteColorButton(screen, communityAccountPaletteOptionButton(index), value, g.profilePaletteSlot == index)
	}
	drawText(screen, "Favorite color", int(communityAccountColorButton().x), 404, colMuted)
	r := communityAccountColorButton()
	drawPaletteColorButton(screen, r, favoriteColor, g.profileColorPicking)
}

func drawPaletteColorButton(screen *ebiten.Image, r rect, value string, selected bool) {
	registerButtonRect(r)
	fill := colWhite
	if selected {
		fill = colPanel
	}
	drawRounded(screen, r, 5, fill)
	outline := colGrid
	if selected {
		outline = colAccent
	}
	drawRectOutline(screen, r, 3, outline)
	if c, ok := parseEditorHexColor(value); ok {
		sw := rect{x: r.x + 8, y: r.y + 6, w: r.w - 16, h: r.h - 12}
		drawRounded(screen, sw, 2, c)
		drawRectOutline(screen, sw, 1, colGridHeavy)
	}
}

func profilePaletteSlots(palette string) [4]string {
	var slots [4]string
	parts := strings.FieldsFunc(palette, func(r rune) bool {
		return r == ',' || r == '|' || r == ' '
	})
	index := 0
	for _, part := range parts {
		if index >= len(slots) {
			break
		}
		if _, ok := parseEditorHexColor(part); ok {
			slots[index] = strings.ToUpper(part)
			index++
		}
	}
	return slots
}

func setProfilePaletteSlot(palette string, slot int, value string) string {
	slots := profilePaletteSlots(palette)
	if slot >= 0 && slot < len(slots) {
		slots[slot] = strings.ToUpper(value)
	}
	return strings.Join(slots[:], ",")
}

func profilePaletteColorInitial(palette string, slot int) string {
	slots := profilePaletteSlots(palette)
	if slot >= 0 && slot < len(slots) {
		return profileColorInitial(slots[slot])
	}
	return profileColorInitial("")
}

func profileColorInitial(value string) string {
	if c, ok := parseEditorHexColor(value); ok {
		return editorColorHex(c)
	}
	return "#A35A4D"
}

func drawSocialField(screen *ebiten.Image, r rect, value string, active bool) {
	drawRounded(screen, r, 4, colPanel)
	outline := colGridHeavy
	if active {
		outline = colAccent
	}
	drawRectOutline(screen, r, 2, outline)
	badge := rect{x: r.x + 7, y: r.y + 7, w: 36, h: 24}
	drawRounded(screen, badge, 3, colWhite)
	drawRectOutline(screen, badge, 2, colGridHeavy)
	drawSocialIcon(screen, badge, value)
	display := value
	if display == "" {
		display = "social link or handle"
		drawText(screen, display, int(r.x+52), int(r.y+25), colMuted)
		if active {
			drawTextCaret(screen, r.x+52, r.y+8, 20)
		}
		return
	}
	shown := truncateText(display, 26)
	drawText(screen, shown, int(r.x+52), int(r.y+25), colInk)
	if active {
		drawTextCaret(screen, r.x+52+float64(len(shown))*8, r.y+8, 20)
	}
}

func drawSocialChip(screen *ebiten.Image, r rect, value string) {
	badge := rect{x: r.x, y: r.y, w: 30, h: r.h}
	drawRounded(screen, badge, 3, colWhite)
	drawRectOutline(screen, badge, 1, colAccent)
	drawSocialIcon(screen, badge, value)
	drawText(screen, truncateText(value, 25), int(r.x+38), int(r.y+16), colAccent)
}

func drawTextCaret(screen *ebiten.Image, x, y, h float64) {
	if time.Now().UnixMilli()/450%2 == 1 {
		return
	}
	vector.DrawFilledRect(screen, float32(x+1), float32(y), 2, float32(h), colAccent, false)
}

func drawSocialIcon(screen *ebiten.Image, r rect, value string) {
	platform := socialPlatform(value)
	if platform == "" {
		if normalized, ok := normalizeProfileSocial(value); ok {
			platform = socialPlatform(normalized)
		}
	}
	x := float32(r.x)
	y := float32(r.y)
	w := float32(r.w)
	h := float32(r.h)
	c := colAccent
	switch platform {
	case "github":
		vector.DrawFilledCircle(screen, x+w/2, y+h/2+1, h*0.30, c, false)
		vector.DrawFilledRect(screen, x+w*0.28, y+h*0.22, w*0.12, h*0.16, c, false)
		vector.DrawFilledRect(screen, x+w*0.60, y+h*0.22, w*0.12, h*0.16, c, false)
		vector.DrawFilledRect(screen, x+w*0.38, y+h*0.72, w*0.24, h*0.12, c, false)
	case "x":
		vector.StrokeLine(screen, x+w*0.30, y+h*0.25, x+w*0.70, y+h*0.75, 4, c, false)
		vector.StrokeLine(screen, x+w*0.70, y+h*0.25, x+w*0.30, y+h*0.75, 4, c, false)
	case "instagram":
		vector.DrawFilledRect(screen, x+w*0.28, y+h*0.25, w*0.46, h*0.50, c, false)
		vector.DrawFilledRect(screen, x+w*0.34, y+h*0.35, w*0.34, h*0.30, colWhite, false)
		vector.DrawFilledCircle(screen, x+w*0.51, y+h*0.50, h*0.09, c, false)
		vector.DrawFilledRect(screen, x+w*0.61, y+h*0.31, w*0.07, h*0.07, c, false)
	case "tiktok":
		vector.DrawFilledRect(screen, x+w*0.50, y+h*0.22, w*0.12, h*0.42, c, false)
		vector.DrawFilledRect(screen, x+w*0.58, y+h*0.30, w*0.20, h*0.10, c, false)
		vector.DrawFilledCircle(screen, x+w*0.42, y+h*0.68, h*0.14, c, false)
	case "youtube":
		vector.DrawFilledRect(screen, x+w*0.25, y+h*0.32, w*0.50, h*0.36, c, false)
		vector.DrawFilledRect(screen, x+w*0.48, y+h*0.40, w*0.09, h*0.20, colWhite, false)
	case "twitch":
		vector.DrawFilledRect(screen, x+w*0.25, y+h*0.22, w*0.50, h*0.48, c, false)
		vector.DrawFilledRect(screen, x+w*0.35, y+h*0.34, w*0.08, h*0.16, colWhite, false)
		vector.DrawFilledRect(screen, x+w*0.56, y+h*0.34, w*0.08, h*0.16, colWhite, false)
		vector.DrawFilledRect(screen, x+w*0.36, y+h*0.70, w*0.22, h*0.10, c, false)
	case "bluesky":
		vector.DrawFilledCircle(screen, x+w*0.40, y+h*0.42, h*0.16, c, false)
		vector.DrawFilledCircle(screen, x+w*0.60, y+h*0.42, h*0.16, c, false)
		vector.DrawFilledCircle(screen, x+w*0.50, y+h*0.62, h*0.14, c, false)
	case "threads":
		vector.DrawFilledCircle(screen, x+w*0.50, y+h*0.50, h*0.24, c, false)
		vector.DrawFilledCircle(screen, x+w*0.50, y+h*0.50, h*0.15, colWhite, false)
		vector.DrawFilledRect(screen, x+w*0.50, y+h*0.22, w*0.16, h*0.18, c, false)
	case "mastodon":
		vector.DrawFilledRect(screen, x+w*0.24, y+h*0.28, w*0.52, h*0.40, c, false)
		vector.DrawFilledRect(screen, x+w*0.34, y+h*0.42, w*0.10, h*0.20, colWhite, false)
		vector.DrawFilledRect(screen, x+w*0.56, y+h*0.42, w*0.10, h*0.20, colWhite, false)
	case "linkedin":
		vector.DrawFilledRect(screen, x+w*0.28, y+h*0.40, w*0.10, h*0.28, c, false)
		vector.DrawFilledRect(screen, x+w*0.45, y+h*0.40, w*0.10, h*0.28, c, false)
		vector.DrawFilledRect(screen, x+w*0.56, y+h*0.48, w*0.10, h*0.20, c, false)
		vector.DrawFilledRect(screen, x+w*0.28, y+h*0.26, w*0.10, h*0.10, c, false)
	default:
		vector.DrawFilledRect(screen, x+w*0.34, y+h*0.46, w*0.32, h*0.08, c, false)
		vector.DrawFilledRect(screen, x+w*0.46, y+h*0.32, w*0.08, h*0.32, c, false)
	}
}

func (g *Game) drawCommunityCreate(screen *ebiten.Image) {
	drawCenteredText(screen, "CREATE", rect{x: 100, y: 202, w: 340, h: 34}, colInk)
	drawButton(screen, communityNewButton(), "Create New Art")
	drawButton(screen, communityImportButton(), "Import Sprite Sheet")
	drawButton(screen, communityImportHelpButton(), "?")
}

func (g *Game) drawCommunityImportHelp(screen *ebiten.Image) {
	drawCenteredText(screen, "SPRITE SHEET IMPORT", rect{x: 48, y: 190, w: 444, h: 34}, colInk)
	panel := rect{x: 48, y: 230, w: 444, h: 370}
	drawRounded(screen, panel, 6, colWhite)
	drawRectOutline(screen, panel, 2, colGridHeavy)
	drawCenteredText(screen, "ONE LEVEL = TWO FRAMES", rect{x: 72, y: 246, w: 236, h: 24}, colMuted)
	drawCenteredText(screen, "1 BEFORE", rect{x: 76, y: 270, w: 112, h: 24}, colMuted)
	drawCenteredText(screen, "2 AFTER", rect{x: 198, y: 270, w: 112, h: 24}, colMuted)
	drawCommunityArtThumbnail(screen, importExampleBefore, rect{x: 88, y: 296, w: 72, h: 72})
	drawCommunityArtThumbnail(screen, importExampleAfter, rect{x: 210, y: 296, w: 72, h: 72})
	drawText(screen, "first frame: before", 316, 300, colInk)
	drawText(screen, "second frame: after", 316, 328, colInk)
	drawText(screen, "repeat pairs for more art", 316, 356, colMuted)
	drawText(screen, "PNG SHEETS", 70, 392, colAccent)
	drawText(screen, "8, 10, 15, 20, or 32 px PNG frames", 70, 422, colInk)
	drawText(screen, "Pairs can run left-right or up-down", 70, 450, colInk)
	drawText(screen, "ASEPRITE", 70, 492, colAccent)
	drawText(screen, "Export PNG + JSON (array data)", 70, 522, colInk)
	drawText(screen, "flower_before / flower_after", 70, 550, colInk)
	drawText(screen, "Each pair imports as one My Art item.", 70, 578, colMuted)
}

func (g *Game) drawCommunityImportPreview(screen *ebiten.Image) {
	drawCenteredText(screen, "IMPORT PREVIEW", rect{x: 70, y: 190, w: 400, h: 30}, colInk)
	drawCenteredText(screen, "Before", rect{x: 82, y: 224, w: 92, h: 20}, colMuted)
	drawCenteredText(screen, "After", rect{x: 196, y: 224, w: 92, h: 20}, colMuted)
	for slot, puzzle := range g.communityImportPack.Levels {
		if slot >= 3 || puzzle == nil {
			break
		}
		before := puzzle.BeforeRaw
		if len(before) == 0 {
			before = puzzle.SkeletonRaw
		}
		drawCommunityArtThumbnail(screen, before, communityImportBeforePreviewRect(slot))
		drawCommunityArtThumbnail(screen, puzzle.RevealRaw, communityImportAfterPreviewRect(slot))
	}
	drawCenteredText(screen, fmt.Sprintf("%d artwork", len(g.communityImportPack.Levels)), rect{x: 90, y: 494, w: 360, h: 30}, colMuted)
	drawButton(screen, communityImportConfirmButton(), "Import to My Art")
}

func (g *Game) drawCommunityImportSetup(screen *ebiten.Image) {
	drawCenteredText(screen, "IMPORT PNG SHEET", rect{x: 70, y: 190, w: 400, h: 30}, colInk)
	drawCenteredText(screen, "Choose the PNG tile size and pair layout.", rect{x: 52, y: 224, w: 436, h: 24}, colMuted)
	drawText(screen, "Tile size", 88, 274, colMuted)
	for index, size := range []int{8, 10, 15, 20, 32} {
		r := communityImportSizeButton(index)
		drawSelectedButton(screen, r, fmt.Sprintf("%d", size), g.importTileSize == size)
	}
	drawText(screen, "Pairs", 88, 370, colMuted)
	drawSelectedButton(screen, communityImportHorizontalButton(), "Before -> After", !g.importVerticalPairs)
	drawSelectedButton(screen, communityImportVerticalButton(), "Before / After", g.importVerticalPairs)
	drawButton(screen, communityImportChooseButton(), "Choose PNG")
}

func (g *Game) drawCommunitySignIn(screen *ebiten.Image) {
	drawCenteredText(screen, "ACCOUNT", rect{x: 100, y: 190, w: 340, h: 32}, colInk)
	panel := rect{x: 72, y: 232, w: 396, h: 486}
	drawRounded(screen, panel, 6, colWhite)
	drawRectOutline(screen, panel, 3, colGridHeavy)
	if communitySignedIn() {
		drawCommunityArtThumbnail(screen, g.profilePixelsOrDefault(), rect{x: 102, y: 262, w: 82, h: 82})
		drawPublishField(screen, communityAccountNameField(), "Name", g.profileNameDraft, g.profileNameEditing)
		drawPublishField(screen, communityAccountBioField(), "Bio", g.profileBioDraft, g.profileBioEditing)
		g.drawProfilePreferenceButtons(screen, g.profilePalette, g.profileColor)
		for slot, social := range g.profileSocialDrafts {
			drawSocialField(screen, communityAccountSocialField(slot), social, g.profileSocialEditing && g.profileSocialSlot == slot)
		}
		drawSupportedSocialIcons(screen)
		g.drawAccountCompletedLevels(screen)
		drawButton(screen, communityAccountBioSaveButton(), "save profile")
		drawButton(screen, communitySignOutButton(), "sign out")
		return
	}
	drawCommunityArtThumbnail(screen, defaultCommunityProfilePixels, rect{x: 230, y: 270, w: 80, h: 80})
	drawButton(screen, communityGoogleButton(), "continue with Google")
	input := communityEmailInput()
	drawRounded(screen, input, 4, colPanel)
	drawRectOutline(screen, input, 3, colGridHeavy)
	email := g.communityEmail
	cursorLength := len(email)
	if email == "" {
		email = "email address"
		drawText(screen, email, int(input.x+12), int(input.y+26), colMuted)
	} else {
		if len(email) > 34 {
			email = email[len(email)-34:]
		}
		drawText(screen, email, int(input.x+12), int(input.y+26), colInk)
	}
	if time.Now().UnixMilli()/500%2 == 0 {
		cursorX := input.x + 12 + float64(cursorLength)*8
		vector.DrawFilledRect(screen, float32(cursorX), float32(input.y+12), 2, 20, colAccent, false)
	}
	drawButton(screen, communitySendLinkButton(), "email sign-in link")
}

func (g *Game) drawAccountCompletedLevels(screen *ebiten.Image) {
	if len(g.communityCompleted) == 0 {
		drawText(screen, "beaten before: none yet", 102, 634, colMuted)
		return
	}
	drawText(screen, "beaten before", 102, 626, colMuted)
	for i := 0; i < len(g.communityCompleted) && i < 5; i++ {
		item := g.communityCompleted[i]
		if item.Puzzle == nil {
			continue
		}
		drawCommunityArtThumbnail(screen, item.Puzzle.RevealRaw, rect{x: 102 + float64(i)*36, y: 638, w: 28, h: 28})
	}
}

func (g *Game) drawCommunityPackBuilder(screen *ebiten.Image) {
	drawCenteredText(screen, fmt.Sprintf("SELECT ART  %d/%d", len(g.packSelection), community.MaxPackItems), rect{x: 80, y: 204, w: 380, h: 32}, colInk)
	start := g.communityPage * communityPackDraftsPerPage
	if start >= len(g.communityLibrary.Drafts) {
		drawCenteredText(screen, "Create some art first", rect{x: 80, y: 360, w: 380, h: 30}, colMuted)
	} else {
		for slot := 0; slot < communityPackDraftsPerPage && start+slot < len(g.communityLibrary.Drafts); slot++ {
			draft := g.communityLibrary.Drafts[start+slot]
			r := communityPackDraftButton(slot)
			drawRounded(screen, r, 5, colWhite)
			drawRectOutline(screen, r, 2, colGridHeavy)
			drawCommunityArtThumbnail(screen, draft.Puzzle.RevealRaw, rect{x: r.x + 8, y: r.y + 7, w: 50, h: 50})
			title := draft.Title
			if len(title) > 24 {
				title = title[:24]
			}
			drawText(screen, title, int(r.x+72), int(r.y+26), colInk)
			box := rect{x: r.x + r.w - 48, y: r.y + 13, w: 36, h: 36}
			drawRounded(screen, box, 3, colPanel)
			drawRectOutline(screen, box, 2, colGridHeavy)
			if g.packSelection[draft.ID] {
				drawCenteredText(screen, "x", box, colAccent)
			}
		}
	}
	drawButton(screen, communityPackDoneButton(), fmt.Sprintf("create pack  %d", len(g.packSelection)))
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*communityPackDraftsPerPage < len(g.communityLibrary.Drafts) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func (g *Game) drawCommunityMyArt(screen *ebiten.Image) {
	drawCenteredText(screen, "MY LIBRARY", rect{x: 100, y: 190, w: 340, h: 30}, colInk)
	drawLibraryTabs(screen, true)
	drawPublishField(screen, communityArtSearchField(), "Search", g.artSearch, g.artSearchActive)
	drawButton(screen, communityArtCreateButton(), "Create")
	indexes := g.filteredCommunityDraftIndexes()
	start := g.communityPage * communityDraftsPerPage
	if start >= len(indexes) {
		drawCenteredText(screen, "No matching art", rect{x: 80, y: 350, w: 380, h: 32}, colMuted)
		return
	}
	for slot := 0; slot < communityDraftsPerPage && start+slot < len(indexes); slot++ {
		draft := g.communityLibrary.Drafts[indexes[start+slot]]
		r := communityMyArtRect(slot)
		drawRounded(screen, r, 6, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
		drawCommunityArtThumbnail(screen, draft.Puzzle.SkeletonRaw, communityMyArtBeforePreviewRect(slot))
		drawCommunityArtThumbnail(screen, draft.Puzzle.RevealRaw, communityMyArtAfterPreviewRect(slot))
		title := draft.Title
		if len(title) > 16 {
			title = title[:16]
		}
		drawText(screen, title, int(r.x+158), int(r.y+27), colInk)
		status := string(draft.Status)
		if status == "" {
			status = "draft"
		}
		statusText := fmt.Sprintf("%dx%d %s", draft.Puzzle.Width, draft.Puzzle.Height, status)
		if g.pendingPublishID == draft.ID {
			statusText = "publishing"
		}
		drawText(screen, statusText, int(r.x+158), int(r.y+54), colMuted)
		drawButton(screen, communityDraftEditButton(slot), "edit")
		publishLabel := "pub"
		if draft.Status == community.LevelPublishedStatus {
			publishLabel = "online"
		}
		if g.pendingPublishID == draft.ID {
			publishLabel = "..."
		}
		drawButton(screen, communityDraftPublishButton(slot), publishLabel)
		drawButton(screen, communityDraftDeleteButton(slot), "x")
	}
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*communityDraftsPerPage < len(indexes) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func (g *Game) drawCommunityNewArtSetup(screen *ebiten.Image) {
	drawCenteredText(screen, "TITLE YOUR ART", rect{x: 80, y: 214, w: 380, h: 34}, colInk)
	drawPublishField(screen, communityNewArtTitleField(), "Title", g.newArtTitle, true)
	drawButton(screen, communityNewArtStartButton(), "Start Drawing")
}

func (g *Game) drawCommunityPackSetup(screen *ebiten.Image) {
	drawCenteredText(screen, "PACK DETAILS", rect{x: 80, y: 190, w: 380, h: 30}, colInk)
	status := "draft"
	if pack, ok := g.communityLibrary.Pack(g.packSetupID); ok && pack.Status == community.LevelPublishedStatus {
		status = "published"
	}
	drawCenteredText(screen, "? is the default cover; art covers are optional   status: "+status, rect{x: 40, y: 216, w: 460, h: 20}, colMuted)
	drawText(screen, "Pack art", 76, 248, colMuted)
	for art, item := range g.packSetupItems {
		if art >= 8 {
			break
		}
		if draft, ok := g.communityLibrary.Draft(item.LevelID); ok && draft.Puzzle != nil {
			if art == g.packSetupPreview {
				drawRounded(screen, inset(communityPackSetupPreview(art), -4), 7, colAccent)
			}
			drawCommunityArtThumbnail(screen, draft.Puzzle.RevealRaw, communityPackSetupPreview(art))
		}
	}
	drawText(screen, "Cover", 382, 248, colMuted)
	if cover := g.packSetupCoverPixels(); len(cover) > 0 {
		if g.packSetupPreview < 0 {
			drawRounded(screen, inset(communityPackSetupCoverPreview(), -4), 7, colAccent)
		}
		drawCommunityArtThumbnail(screen, cover, communityPackSetupCoverPreview())
	}
	drawButton(screen, communityPackUploadCoverButton(), "or upload cover")
	drawButton(screen, communityPackQuestionCoverButton(), "use ?")
	drawPublishField(screen, communityPackTitleField(), "Title", g.packSetupTitle, g.packSetupField == 0)
	drawPublishField(screen, communityPackDescriptionField(), "Description", g.packSetupDescription, g.packSetupField == 1)
	drawButton(screen, communityPackSaveDraftButton(), "Save Draft")
	publishLabel := "Publish"
	if g.packPublishAwaitingID != "" {
		publishLabel = "Publishing..."
	}
	drawButton(screen, communityPackSetupPublishButton(), publishLabel)
}

func (g *Game) packSetupCoverPixels() [][]string {
	if len(g.packSetupPreviewRaw) > 0 {
		return g.packSetupPreviewRaw
	}
	if g.packSetupPreview >= 0 && g.packSetupPreview < len(g.packSetupItems) {
		item := g.packSetupItems[g.packSetupPreview]
		if draft, ok := g.communityLibrary.Draft(item.LevelID); ok && draft.Puzzle != nil {
			return draft.Puzzle.RevealRaw
		}
	}
	return nil
}

func drawCommunityBackdrop(screen *ebiten.Image) {
	screen.Fill(colPanel)
	const pixelSize = 6
	elapsed := float64(time.Now().UnixMilli()%240000) / 1000
	for y := 0; y < 186; y += pixelSize {
		for x := 0; x < ScreenWidth; x += pixelSize {
			n := perlin2D(float64(x)/92+elapsed*0.025, float64(y)/72+elapsed*0.009)
			cloud := math.Max(0, math.Abs(n)-0.08) / 0.55
			cloud = math.Min(1, cloud)
			base := color.RGBA{14, 18, 27, 255}
			if n >= 0 {
				base = mixCommunityColor(base, color.RGBA{43, 100, 101, 255}, cloud)
			} else {
				base = mixCommunityColor(base, color.RGBA{116, 62, 80, 255}, cloud)
			}
			vector.DrawFilledRect(screen, float32(x), float32(y), pixelSize, pixelSize, base, false)
			star := communityStarHash(x/pixelSize, y/pixelSize)
			if star%113 == 0 {
				size := float32(2)
				if star%7 == 0 {
					size = 3
				}
				vector.DrawFilledRect(screen, float32(x+2), float32(y+2), size, size, colWhite, false)
			}
		}
	}
	drawCommunityShip(screen, shipPosition(elapsed, 17, 620), 20, false, color.RGBA{239, 235, 220, 255})
	drawCommunityShip(screen, shipPosition(elapsed, 10, 700), 146, true, color.RGBA{235, 107, 86, 255})
	drawCommunityShip(screen, shipPosition(elapsed+31, 7, 760), 118, false, color.RGBA{75, 143, 140, 255})
	vector.DrawFilledRect(screen, 0, 176, ScreenWidth, 12, colGridHeavy, false)
	drawRounded(screen, rect{x: 62, y: 32, w: 416, h: 92}, 8, color.RGBA{45, 45, 43, 255})
	drawRounded(screen, rect{x: 76, y: 46, w: 388, h: 52}, 6, colWhite)
}

func shipPosition(elapsed, speed, distance float64) int {
	return int(math.Mod(elapsed*speed, distance)) - 60
}

func drawCommunityShip(screen *ebiten.Image, x, y int, reverse bool, hull color.RGBA) {
	direction := 1
	if reverse {
		x = ScreenWidth - x
		direction = -1
	}
	px := func(offsetX, offsetY, width, height int, c color.Color) {
		if direction < 0 {
			offsetX = -offsetX - width
		}
		vector.DrawFilledRect(screen, float32(x+offsetX), float32(y+offsetY), float32(width), float32(height), c, false)
	}
	px(-18, -3, 12, 6, color.RGBA{244, 201, 93, 210})
	px(-10, -5, 24, 10, hull)
	px(8, -9, 12, 18, hull)
	px(12, -4, 8, 8, color.RGBA{86, 115, 134, 255})
	px(-2, -10, 8, 5, hull)
	px(-2, 5, 8, 5, hull)
}

func perlin2D(x, y float64) float64 {
	x0 := math.Floor(x)
	y0 := math.Floor(y)
	sx := perlinFade(x - x0)
	sy := perlinFade(y - y0)
	n00 := perlinGradient(int(x0), int(y0), x-x0, y-y0)
	n10 := perlinGradient(int(x0)+1, int(y0), x-x0-1, y-y0)
	n01 := perlinGradient(int(x0), int(y0)+1, x-x0, y-y0-1)
	n11 := perlinGradient(int(x0)+1, int(y0)+1, x-x0-1, y-y0-1)
	return perlinLerp(perlinLerp(n00, n10, sx), perlinLerp(n01, n11, sx), sy)
}

func perlinFade(t float64) float64       { return t * t * t * (t*(t*6-15) + 10) }
func perlinLerp(a, b, t float64) float64 { return a + (b-a)*t }

func perlinGradient(ix, iy int, x, y float64) float64 {
	switch communityStarHash(ix, iy) & 7 {
	case 0:
		return x + y
	case 1:
		return -x + y
	case 2:
		return x - y
	case 3:
		return -x - y
	case 4:
		return x
	case 5:
		return -x
	case 6:
		return y
	default:
		return -y
	}
}

func communityStarHash(x, y int) uint32 {
	h := uint32(x)*0x8da6b343 ^ uint32(y)*0xd8163841
	h ^= h >> 13
	h *= 0x85ebca6b
	return h ^ (h >> 16)
}

func mixCommunityColor(a, b color.RGBA, amount float64) color.RGBA {
	mix := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*amount) }
	return color.RGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), 255}
}

func drawLibraryTabs(screen *ebiten.Image, art bool) {
	drawSelectedButton(screen, communityLibraryArtTab(), "Art", art)
	drawSelectedButton(screen, communityLibraryPacksTab(), "Packs", !art)
	drawSelectedButton(screen, communityLibraryPublishedTab(), "Published", false)
}

func (g *Game) drawCommunityPublished(screen *ebiten.Image) {
	drawCenteredText(screen, "MY LIBRARY", rect{x: 100, y: 190, w: 340, h: 30}, colInk)
	drawSelectedButton(screen, communityLibraryArtTab(), "Art", false)
	drawSelectedButton(screen, communityLibraryPacksTab(), "Packs", false)
	drawSelectedButton(screen, communityLibraryPublishedTab(), "Published", true)
	if len(g.communityPublished) == 0 {
		drawCenteredText(screen, "Nothing published yet", rect{x: 70, y: 360, w: 400, h: 32}, colMuted)
		return
	}
	for slot, item := range g.communityPublished {
		if slot >= 4 {
			break
		}
		r := communityPublishedRect(slot)
		fill := colWhite
		outline := colGridHeavy
		if item.Kind == "pack" {
			fill = colPanel
			outline = colAccent
		}
		drawRounded(screen, r, 6, fill)
		drawRectOutline(screen, r, 2, outline)
		if item.Kind == "pack" {
			for art := 0; art < 3 && art < len(item.Levels); art++ {
				drawCommunityArtThumbnail(screen, communityQuestionCover(), rect{x: r.x + 8 + float64(art)*30, y: r.y + 12, w: 28, h: 28})
			}
			if len(item.Levels) == 0 {
				drawCommunityArtThumbnail(screen, communityQuestionCover(), rect{x: r.x + 8, y: r.y + 8, w: 54, h: 54})
			}
		} else if item.Puzzle != nil {
			drawCommunityArtThumbnail(screen, item.Puzzle.RevealRaw, rect{x: r.x + 8, y: r.y + 8, w: 54, h: 54})
		}
		title := item.Title
		if len(title) > 17 {
			title = title[:17]
		}
		textX := int(r.x + 72)
		if item.Kind == "pack" {
			textX = int(r.x + 112)
		}
		drawText(screen, title, textX, int(r.y+24), colInk)
		if item.Kind == "pack" {
			drawText(screen, fmt.Sprintf("PACK  %d art", len(item.Levels)), textX, int(r.y+49), colAccent)
		} else {
			drawText(screen, "ART", textX, int(r.y+49), colMuted)
		}
		drawButton(screen, communityPublishedRemoveButton(slot), "edit")
	}
}

func (g *Game) drawCommunityPublishedEdit(screen *ebiten.Image) {
	drawCenteredText(screen, "EDIT PUBLISHED", rect{x: 80, y: 190, w: 380, h: 30}, colInk)
	drawCenteredText(screen, strings.ToUpper(g.publishedEditKind)+" is published", rect{x: 80, y: 220, w: 380, h: 24}, colMuted)
	if g.publishedEditIndex >= 0 && g.publishedEditIndex < len(g.communityPublished) {
		item := g.communityPublished[g.publishedEditIndex]
		if item.Kind == "pack" {
			drawCenteredText(screen, "PACK CONTENTS", rect{x: 74, y: 252, w: 392, h: 24}, colMuted)
			if len(item.PreviewRaw) > 0 {
				drawText(screen, "Cover", 416, 258, colMuted)
				drawCommunityArtThumbnail(screen, item.PreviewRaw, rect{x: 412, y: 274, w: 42, h: 42})
			}
			for slot := 0; slot < len(g.publishedEditLevels) && slot < 8; slot++ {
				level := g.publishedEditLevels[slot]
				r := communityPublishedEditLevelButton(slot)
				drawRounded(screen, r, 4, colPanel)
				drawRectOutline(screen, r, 2, colGridHeavy)
				if level.Puzzle != nil {
					drawCommunityArtThumbnail(screen, level.Puzzle.RevealRaw, rect{x: r.x + 5, y: r.y + 5, w: 36, h: 36})
				} else {
					drawCommunityArtThumbnail(screen, communityQuestionCover(), rect{x: r.x + 5, y: r.y + 5, w: 36, h: 36})
				}
				drawText(screen, truncateText(level.Title, 18), int(r.x+50), int(r.y+28), colInk)
				drawButton(screen, communityPublishedEditLevelRemoveButton(slot), "x")
			}
			drawButton(screen, communityPublishedEditAddLevelButton(), fmt.Sprintf("add art  %d/%d", len(g.publishedEditLevels), community.MaxPackItems))
		} else if item.Puzzle != nil {
			drawText(screen, "Before", 124, 260, colMuted)
			drawText(screen, "After", 308, 260, colMuted)
			drawCommunityArtThumbnail(screen, item.Puzzle.SkeletonRaw, rect{x: 102, y: 270, w: 82, h: 82})
			drawCommunityArtThumbnail(screen, item.Puzzle.RevealRaw, rect{x: 286, y: 270, w: 82, h: 82})
		} else {
			drawText(screen, "Before", 124, 260, colMuted)
			drawText(screen, "After", 308, 260, colMuted)
			drawCommunityArtThumbnail(screen, communityQuestionCover(), rect{x: 102, y: 270, w: 82, h: 82})
			if len(item.PreviewRaw) > 0 {
				drawCommunityArtThumbnail(screen, item.PreviewRaw, rect{x: 286, y: 270, w: 82, h: 82})
			} else {
				drawCommunityArtThumbnail(screen, communityQuestionCover(), rect{x: 286, y: 270, w: 82, h: 82})
			}
		}
	}
	drawPublishField(screen, communityPublishedEditTitleField(), "Name", g.publishedEditTitle, g.publishedEditField == 0)
	drawPublishField(screen, communityPublishedEditDescriptionField(), "Bio / description", g.publishedEditDescription, g.publishedEditField == 1)
	drawButton(screen, communityPublishedApplyButton(), "apply changes")
	drawButton(screen, communityPublishedUnpublishButton(), "unpublish")
}

func (g *Game) drawCommunityPublishedPackAdd(screen *ebiten.Image) {
	drawCenteredText(screen, fmt.Sprintf("ADD ART TO PACK  %d/%d", len(g.publishedEditLevels), community.MaxPackItems), rect{x: 80, y: 204, w: 380, h: 32}, colInk)
	used := make(map[string]bool, len(g.publishedEditLevels))
	for _, level := range g.publishedEditLevels {
		used[level.LocalID] = true
		used[level.LevelID] = true
	}
	start := g.communityPage * communityPackDraftsPerPage
	if start >= len(g.communityLibrary.Drafts) {
		drawCenteredText(screen, "No local art yet", rect{x: 80, y: 360, w: 380, h: 30}, colMuted)
	} else {
		for slot := 0; slot < communityPackDraftsPerPage && start+slot < len(g.communityLibrary.Drafts); slot++ {
			draft := g.communityLibrary.Drafts[start+slot]
			r := communityPackDraftButton(slot)
			drawRounded(screen, r, 5, colWhite)
			drawRectOutline(screen, r, 2, colGridHeavy)
			drawCommunityArtThumbnail(screen, draft.Puzzle.RevealRaw, rect{x: r.x + 8, y: r.y + 7, w: 50, h: 50})
			drawText(screen, truncateText(draft.Title, 24), int(r.x+72), int(r.y+26), colInk)
			label := "add"
			if used[draft.ID] {
				label = "in pack"
			}
			drawButton(screen, rect{x: r.x + r.w - 96, y: r.y + 18, w: 82, h: 28}, label)
		}
	}
	if g.communityPage > 0 {
		drawButton(screen, communityPrevButton(), "prev")
	}
	if (g.communityPage+1)*communityPackDraftsPerPage < len(g.communityLibrary.Drafts) {
		drawButton(screen, communityNextButton(), "next")
	}
}

func (g *Game) drawCommunityPublishSetup(screen *ebiten.Image) {
	draft, ok := g.communityLibrary.Draft(g.publishDraftID)
	if !ok {
		return
	}
	drawCenteredText(screen, "PUBLISH ART", rect{x: 80, y: 190, w: 380, h: 30}, colInk)
	drawText(screen, "Before", 76, 232, colMuted)
	drawText(screen, "After", 178, 232, colMuted)
	drawText(screen, "Cover", 82, 406, colMuted)
	drawCommunityArtThumbnail(screen, draft.Puzzle.SkeletonRaw, rect{x: 52, y: 238, w: 94, h: 94})
	drawCommunityArtThumbnail(screen, draft.Puzzle.RevealRaw, rect{x: 154, y: 238, w: 94, h: 94})
	cover := draft.Puzzle.RevealRaw
	if len(g.publishPreviewRaw) > 0 {
		cover = g.publishPreviewRaw
	}
	drawCommunityArtThumbnail(screen, cover, rect{x: 52, y: 422, w: 94, h: 94})
	drawText(screen, "Cover options", 166, 412, colMuted)
	drawButton(screen, communityPublishCoverButton(), "upload")
	drawButton(screen, communityPublishFinalCoverButton(), "use final")
	drawButton(screen, communityPublishQuestionCoverButton(), "use ?")
	drawPublishField(screen, communityPublishTitleField(), "Name - hint at what it is!", g.publishTitle, g.publishField == 0)
	drawPublishField(screen, communityPublishDescriptionField(), "Description", g.publishDescription, g.publishField == 1)
	drawPublishField(screen, communityPublishTagsField(), "Tags", g.publishTags, g.publishField == 2)
	drawSelectedButton(screen, communityPublishOfficialButton(), checkboxLabel(g.publishSubmitOfficial, "Main game review"), g.publishSubmitOfficial)
	if g.publishSubmitOfficial {
		drawSelectedButton(screen, communityPublishRightsButton(), checkboxLabel(g.publishRightsConfirmed, "I own this art"), g.publishRightsConfirmed)
	}
	publishLabel := "Publish"
	if g.publishAwaitingID != "" {
		publishLabel = "Publishing..."
	}
	drawButton(screen, communityPublishConfirmButton(), publishLabel)
}

func drawPublishField(screen *ebiten.Image, r rect, label, value string, active bool) {
	drawRounded(screen, r, 4, colWhite)
	drawRectOutline(screen, r, 2, colGridHeavy)
	if active {
		drawRounded(screen, inset(r, -3), 7, colAccent)
		drawRounded(screen, r, 4, colWhite)
		drawRectOutline(screen, r, 2, colGridHeavy)
	}
	drawText(screen, label, int(r.x), int(r.y-8), colMuted)
	shown := value
	max := int((r.w - 16) / 8)
	if len(shown) > max {
		shown = shown[len(shown)-max:]
	}
	drawText(screen, shown, int(r.x+8), int(r.y+25), colInk)
	if active {
		drawTextCaret(screen, r.x+8+float64(len(shown))*8, r.y+10, 22)
	}
}

func checkboxLabel(checked bool, text string) string {
	if checked {
		return "x  " + text
	}
	return "o  " + text
}

func drawCommunityArtThumbnail(screen *ebiten.Image, pixels [][]string, frame rect) {
	drawRounded(screen, frame, 4, colPanel)
	drawRectOutline(screen, frame, 2, colGridHeavy)
	if len(pixels) == 0 {
		return
	}
	height := len(pixels)
	width := len(pixels[0])
	if width == 0 {
		return
	}
	cell := minFloat((frame.w-6)/float64(width), (frame.h-6)/float64(height))
	originX := frame.x + (frame.w-cell*float64(width))/2
	originY := frame.y + (frame.h-cell*float64(height))/2
	for y, row := range pixels {
		for x, value := range row {
			c, ok := parseEditorHexColor(value)
			if !ok || c.A == 0 {
				continue
			}
			vector.DrawFilledRect(screen, float32(originX+float64(x)*cell), float32(originY+float64(y)*cell), float32(cell+0.25), float32(cell+0.25), c, false)
		}
	}
}

func (g *Game) drawCommunityEmpty(screen *ebiten.Image, title, message string) {
	drawCenteredText(screen, title, rect{x: 100, y: 212, w: 340, h: 34}, colInk)
	drawCenteredText(screen, message, rect{x: 60, y: 326, w: 420, h: 70}, colMuted)
}

func communityBackButton() rect          { return rect{x: 202, y: 724, w: 136, h: 42} }
func communityAccountButton() rect       { return rect{x: 322, y: 208, w: 98, h: 40} }
func communityProfileBadgeButton() rect  { return rect{x: 430, y: 198, w: 58, h: 58} }
func communityLevelsButton() rect        { return rect{x: 108, y: 314, w: 324, h: 50} }
func communityPacksButton() rect         { return rect{x: 128, y: 348, w: 284, h: 46} }
func communityCreateButton() rect        { return rect{x: 128, y: 410, w: 284, h: 46} }
func communityMyArtButton() rect         { return rect{x: 108, y: 388, w: 324, h: 50} }
func communityCreatorsButton() rect      { return rect{x: 108, y: 462, w: 324, h: 50} }
func communityNewButton() rect           { return rect{x: 104, y: 270, w: 332, h: 48} }
func communityImportButton() rect        { return rect{x: 104, y: 338, w: 278, h: 48} }
func communityImportHelpButton() rect    { return rect{x: 392, y: 338, w: 44, h: 48} }
func communityLibraryArtTab() rect       { return rect{x: 54, y: 226, w: 136, h: 36} }
func communityLibraryPacksTab() rect     { return rect{x: 194, y: 226, w: 136, h: 36} }
func communityLibraryPublishedTab() rect { return rect{x: 334, y: 226, w: 152, h: 36} }
func communityPublishedRect(slot int) rect {
	return rect{x: 48, y: 278 + float64(slot)*78, w: 444, h: 70}
}
func communityPublishedPinButton(slot int) rect {
	r := communityPublishedRect(slot)
	return rect{x: r.x + 284, y: r.y + 17, w: 54, h: 36}
}
func communityPublishedRemoveButton(slot int) rect {
	r := communityPublishedRect(slot)
	return rect{x: r.x + 344, y: r.y + 17, w: 88, h: 36}
}
func communityPublishedEditLevelButton(slot int) rect {
	column := slot % 2
	row := slot / 2
	return rect{x: 64 + float64(column)*230, y: 278 + float64(row)*54, w: 212, h: 42}
}
func communityPublishedEditLevelRemoveButton(slot int) rect {
	r := communityPublishedEditLevelButton(slot)
	return rect{x: r.x + r.w - 32, y: r.y + 6, w: 26, h: 26}
}
func communityPublishedEditAddLevelButton() rect {
	return rect{x: 172, y: 390, w: 196, h: 30}
}
func communityPublishedEditTitleField() rect       { return rect{x: 84, y: 442, w: 372, h: 44} }
func communityPublishedEditDescriptionField() rect { return rect{x: 84, y: 514, w: 372, h: 44} }
func communityPublishedApplyButton() rect          { return rect{x: 84, y: 592, w: 180, h: 44} }
func communityPublishedUnpublishButton() rect      { return rect{x: 276, y: 592, w: 180, h: 44} }
func communityPublishTitleField() rect             { return rect{x: 270, y: 238, w: 220, h: 40} }
func communityPublishDescriptionField() rect       { return rect{x: 270, y: 316, w: 220, h: 40} }
func communityPublishTagsField() rect              { return rect{x: 270, y: 386, w: 220, h: 40} }
func communityPublishOfficialButton() rect         { return rect{x: 88, y: 544, w: 364, h: 40} }
func communityPublishRightsButton() rect           { return rect{x: 88, y: 592, w: 364, h: 40} }
func communityPublishConfirmButton() rect          { return rect{x: 170, y: 644, w: 200, h: 44} }
func communityPublishCoverButton() rect            { return rect{x: 166, y: 434, w: 82, h: 34} }
func communityPublishFinalCoverButton() rect       { return rect{x: 256, y: 434, w: 98, h: 34} }
func communityPublishQuestionCoverButton() rect    { return rect{x: 362, y: 434, w: 82, h: 34} }
func communityImportPreviewRect(slot int) rect {
	column := slot % 3
	row := slot / 3
	return rect{x: 68 + float64(column)*140, y: 246 + float64(row)*120, w: 104, h: 104}
}
func communityImportBeforePreviewRect(slot int) rect {
	return rect{x: 64, y: 250 + float64(slot)*78, w: 64, h: 64}
}
func communityImportAfterPreviewRect(slot int) rect {
	return rect{x: 178, y: 250 + float64(slot)*78, w: 64, h: 64}
}
func communityImportConfirmButton() rect { return rect{x: 142, y: 540, w: 256, h: 44} }
func communityImportSizeButton(slot int) rect {
	return rect{x: 86 + float64(slot)*76, y: 294, w: 58, h: 38}
}
func communityImportHorizontalButton() rect { return rect{x: 86, y: 392, w: 168, h: 42} }
func communityImportVerticalButton() rect   { return rect{x: 286, y: 392, w: 168, h: 42} }
func communityImportChooseButton() rect     { return rect{x: 150, y: 500, w: 240, h: 46} }
func communityArtSearchField() rect         { return rect{x: 48, y: 290, w: 300, h: 32} }
func communityArtCreateButton() rect        { return rect{x: 358, y: 290, w: 134, h: 32} }
func communityNewArtTitleField() rect       { return rect{x: 96, y: 290, w: 348, h: 44} }
func communityNewArtStartButton() rect      { return rect{x: 154, y: 370, w: 232, h: 44} }
func communityPackSetupPreview(art int) rect {
	column := art % 4
	row := art / 4
	return rect{x: 74 + float64(column)*58, y: 270 + float64(row)*54, w: 46, h: 46}
}
func communityPackSetupCoverPreview() rect   { return rect{x: 374, y: 274, w: 56, h: 56} }
func communityPackUploadCoverButton() rect   { return rect{x: 90, y: 356, w: 172, h: 32} }
func communityPackQuestionCoverButton() rect { return rect{x: 278, y: 356, w: 172, h: 32} }
func communityPackTitleField() rect          { return rect{x: 76, y: 424, w: 388, h: 40} }
func communityPackDescriptionField() rect    { return rect{x: 76, y: 492, w: 388, h: 40} }
func communityPackSaveDraftButton() rect     { return rect{x: 76, y: 568, w: 184, h: 44} }
func communityPackSetupPublishButton() rect  { return rect{x: 280, y: 568, w: 184, h: 44} }
func communityDraftRect(slot int) rect {
	return rect{x: 54, y: 234 + float64(slot)*88, w: 432, h: 74}
}
func communityMyArtRect(slot int) rect {
	return rect{x: 48, y: 332 + float64(slot)*92, w: 444, h: 82}
}
func communityMyArtBeforePreviewRect(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 10, y: r.y + 10, w: 62, h: 62}
}
func communityMyArtAfterPreviewRect(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 80, y: r.y + 10, w: 62, h: 62}
}
func communityDraftEditButton(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 286, y: r.y + 23, w: 44, h: 38}
}
func communityDraftPublishButton(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 336, y: r.y + 23, w: 62, h: 38}
}
func communityDraftDeleteButton(slot int) rect {
	r := communityMyArtRect(slot)
	return rect{x: r.x + 406, y: r.y + 23, w: 28, h: 38}
}
func communityPrevButton() rect { return rect{x: 62, y: 604, w: 92, h: 38} }
func communityNextButton() rect { return rect{x: 386, y: 604, w: 92, h: 38} }
func communityCatalogPlayButton(slot int) rect {
	r := communityDraftRect(slot)
	return rect{x: r.x + r.w - 90, y: r.y + 17, w: 72, h: 40}
}
func communityGalleryAllButton() rect    { return rect{x: 48, y: 226, w: 62, h: 34} }
func communityGalleryArtButton() rect    { return rect{x: 114, y: 226, w: 62, h: 34} }
func communityGalleryPacksButton() rect  { return rect{x: 180, y: 226, w: 78, h: 34} }
func communityGallerySortButton() rect   { return rect{x: 286, y: 226, w: 166, h: 34} }
func communityGallerySortMenu() rect     { return rect{x: 286, y: 264, w: 166, h: 96} }
func communityGalleryNewButton() rect    { return rect{x: 286, y: 264, w: 166, h: 32} }
func communityGalleryPlayedButton() rect { return rect{x: 286, y: 296, w: 166, h: 32} }
func communityGalleryTopButton() rect    { return rect{x: 286, y: 328, w: 166, h: 32} }
func communityGalleryCard(slot int) rect {
	return rect{x: 44, y: 314 + float64(slot)*92, w: 452, h: 86}
}
func communityGalleryOpenButton(slot int) rect {
	r := communityGalleryCard(slot)
	return rect{x: r.x + 318, y: r.y + 46, w: 48, h: 28}
}
func communityGalleryChatButton(slot int) rect {
	r := communityGalleryCard(slot)
	return rect{x: r.x + 370, y: r.y + 46, w: 30, h: 28}
}
func communityGalleryLikeButton(slot int) rect {
	r := communityGalleryCard(slot)
	return rect{x: r.x + 404, y: r.y + 46, w: 40, h: 28}
}
func communityGalleryPromoteButton(slot int) rect {
	r := communityGalleryCard(slot)
	return rect{x: r.x + 396, y: r.y + 46, w: 44, h: 28}
}
func communityGalleryPackLevelButton(slot int) rect {
	column := slot % 2
	row := slot / 2
	return rect{x: 46 + float64(column)*232, y: 270 + float64(row)*92, w: 216, h: 78}
}
func communityGalleryPackChatButton() rect { return rect{x: 410, y: 222, w: 42, h: 28} }
func communityChatMessageButton(slot int) rect {
	return rect{x: 56, y: 272 + float64(slot)*50, w: 420, h: 48}
}
func communityChatInputField() rect { return rect{x: 56, y: 564, w: 326, h: 42} }
func communityChatSendButton() rect { return rect{x: 392, y: 564, w: 92, h: 42} }
func communityCreatorSearchField() rect {
	return rect{x: 80, y: 232, w: 380, h: 36}
}
func communityCreatorProfileBaseContentY(social, bio, palette, favoriteColor string) float64 {
	nextY := 292
	if bio != "" {
		nextY += wrappedTextLineCount(bio, 38, 3)*20 + 14
	}
	if socialCount := countProfileSocials(social); socialCount > 0 {
		nextY += socialCount*24 + 8
	}
	if palette != "" || favoriteColor != "" {
		nextY += 26
	}
	if nextY+34 < 366 {
		return 366
	}
	return float64(nextY + 34)
}
func communityCreatorProfileLevelsY(social, bio, palette, favoriteColor string, featured bool) float64 {
	contentY := communityCreatorProfileBaseContentY(social, bio, palette, favoriteColor)
	if !featured {
		return contentY
	}
	r := communityCreatorFeaturedButtonAt(contentY + 30)
	return r.y + r.h + 12
}
func communityCreatorFeaturedButtonAt(y float64) rect { return rect{x: 46, y: y, w: 448, h: 70} }
func communityCreatorLevelButtonAt(slot int, y float64) rect {
	column := slot % 2
	row := slot / 2
	return rect{x: 46 + float64(column)*232, y: y + float64(row)*92, w: 216, h: 78}
}
func communityPackCreateButton() rect { return rect{x: 154, y: 270, w: 232, h: 38} }
func communityPackRect(slot int) rect { return rect{x: 44, y: 318 + float64(slot)*82, w: 452, h: 76} }
func communityPackArtPreview(slot, art, count int) rect {
	r := communityPackRect(slot)
	if count <= 5 {
		return rect{x: r.x + 274 + float64(art)*24, y: r.y + 8, w: 22, h: 22}
	}
	column := art % 10
	row := art / 10
	return rect{x: r.x + 274 + float64(column)*12, y: r.y + 6 + float64(row)*16, w: 11, h: 11}
}
func communityPackPlayButton(slot int) rect {
	r := communityPackRect(slot)
	return rect{x: r.x + 258, y: r.y + 40, w: 52, h: 28}
}
func communityPackPublishButton(slot int) rect {
	r := communityPackRect(slot)
	return rect{x: r.x + 316, y: r.y + 40, w: 80, h: 28}
}
func communityPackDeleteButton(slot int) rect {
	r := communityPackRect(slot)
	return rect{x: r.x + 404, y: r.y + 40, w: 28, h: 28}
}
func communityGoogleButton() rect     { return rect{x: 122, y: 370, w: 296, h: 42} }
func communityEmailInput() rect       { return rect{x: 102, y: 434, w: 336, h: 44} }
func communitySendLinkButton() rect   { return rect{x: 142, y: 500, w: 256, h: 42} }
func communityAccountNameField() rect { return rect{x: 210, y: 270, w: 228, h: 34} }
func communityAccountBioField() rect  { return rect{x: 210, y: 340, w: 228, h: 44} }
func communityAccountPaletteOptionButton(slot int) rect {
	return rect{x: 102 + float64(slot)*48, y: 420, w: 40, h: 30}
}
func communityAccountColorButton() rect {
	return rect{x: 344, y: 420, w: 40, h: 30}
}
func communityAccountSocialField(slot int) rect {
	return rect{x: 102, y: 468 + float64(slot)*36, w: 336, h: 32}
}
func communityAccountBioSaveButton() rect { return rect{x: 102, y: 674, w: 158, h: 30} }
func communitySignOutButton() rect        { return rect{x: 280, y: 674, w: 158, h: 30} }
func communityPackDraftButton(slot int) rect {
	return rect{x: 66, y: 246 + float64(slot)*72, w: 408, h: 64}
}
func communityPackDoneButton() rect { return rect{x: 160, y: 552, w: 220, h: 42} }
func communityCreatorButton(slot int) rect {
	return rect{x: 48, y: 286 + float64(slot)*88, w: 444, h: 76}
}
func communityCreatorPreviewRect(slot, art int) rect {
	r := communityCreatorButton(slot)
	return rect{x: r.x + 292 + float64(art)*48, y: r.y + 16, w: 42, h: 42}
}
func communityCreatorLevelButton(slot int) rect {
	column := slot % 2
	row := slot / 2
	return rect{x: 46 + float64(column)*232, y: 310 + float64(row)*92, w: 216, h: 78}
}

var defaultCommunityProfilePixels = func() [][]string {
	rows := []string{
		"0000110000",
		"0001111000",
		"0001111000",
		"0000110000",
		"0011111100",
		"0111111110",
		"0111111110",
		"0100000010",
		"0100000010",
		"0000000000",
	}
	pixels := make([][]string, len(rows))
	for y, row := range rows {
		pixels[y] = make([]string, len(row))
		for x, value := range row {
			if value == '1' {
				pixels[y][x] = "#000000FF"
			}
		}
	}
	return pixels
}()

var importExampleBefore = pixelExample([]string{
	"00011000", "00111100", "01111110", "11111111",
	"11111111", "01111110", "00111100", "00011000",
}, []string{"#000000FF"})

var importExampleAfter = pixelExample([]string{
	"00011000", "00122100", "01222210", "12233221",
	"12233221", "01222210", "00122100", "00011000",
}, []string{"#EB6B56FF", "#F4C95DFF", "#4B8F8CFF"})

func pixelExample(rows []string, colors []string) [][]string {
	pixels := make([][]string, len(rows))
	for y, row := range rows {
		pixels[y] = make([]string, len(row))
		for x, value := range row {
			if value == '0' {
				continue
			}
			index := int(value - '1')
			if index >= 0 && index < len(colors) {
				pixels[y][x] = colors[index]
			}
		}
	}
	return pixels
}

func truncateText(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func drawWrappedText(screen *ebiten.Image, text string, x, y, maxChars, maxLines int, c color.Color) {
	lines := wrapTextLines(text, maxChars, maxLines)
	for index, line := range lines {
		drawText(screen, line, x, y+index*20, c)
	}
}

func wrapTextLines(text string, maxChars, maxLines int) []string {
	words := strings.Fields(text)
	if len(words) == 0 || maxChars <= 0 || maxLines <= 0 {
		return nil
	}
	line := ""
	lines := make([]string, 0, maxLines)
	for _, word := range words {
		next := word
		if line != "" {
			next = line + " " + word
		}
		if len(next) <= maxChars {
			line = next
			continue
		}
		lines = append(lines, truncateText(line, maxChars))
		if len(lines) >= maxLines {
			return lines
		}
		line = word
	}
	if line != "" && len(lines) < maxLines {
		lines = append(lines, truncateText(line, maxChars))
	}
	return lines
}

func wrappedTextLineCount(text string, maxChars, maxLines int) int {
	return len(wrapTextLines(text, maxChars, maxLines))
}

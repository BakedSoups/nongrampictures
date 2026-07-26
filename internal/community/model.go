package community

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BakedSoups/community_nongrams/internal/nonogram"
)

const MaxPackItems = 32

type Visibility string

const (
	VisibilityDraft    Visibility = "draft"
	VisibilityPublic   Visibility = "public"
	VisibilityPackOnly Visibility = "pack_only"
	VisibilityUnlisted Visibility = "unlisted"
)

type LevelStatus string

const (
	LevelDraftStatus     LevelStatus = "draft"
	LevelPublishedStatus LevelStatus = "published"
	LevelHiddenStatus    LevelStatus = "hidden"
	LevelRemovedStatus   LevelStatus = "removed"
)

type SubmissionStatus string

const (
	SubmissionSubmitted       SubmissionStatus = "submitted"
	SubmissionInReview        SubmissionStatus = "in_review"
	SubmissionChangesRequired SubmissionStatus = "changes_requested"
	SubmissionApproved        SubmissionStatus = "approved"
	SubmissionDeclined        SubmissionStatus = "declined"
)

type LevelDraft struct {
	ID          string              `json:"id"`
	OwnerID     string              `json:"ownerId,omitempty"`
	Title       string              `json:"title"`
	Description string              `json:"description,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Visibility  Visibility          `json:"visibility"`
	Status      LevelStatus         `json:"status"`
	Playtested  bool                `json:"playtested"`
	Version     int                 `json:"version"`
	UpdatedAt   string              `json:"updatedAt"`
	Puzzle      *nonogram.Puzzle    `json:"puzzle"`
	Stats       CommunityLevelStats `json:"stats,omitempty"`
}

type CommunityLevelStats struct {
	Plays int `json:"plays"`
	Likes int `json:"likes"`
}

type CreatorProfile struct {
	ID            string           `json:"id"`
	DisplayName   string           `json:"displayName"`
	Bio           string           `json:"bio,omitempty"`
	Social        string           `json:"social,omitempty"`
	Palette       string           `json:"palette,omitempty"`
	FavoriteColor string           `json:"favoriteColor,omitempty"`
	AvatarPuzzle  *nonogram.Puzzle `json:"avatarPuzzle,omitempty"`
	Featured      []GalleryItem    `json:"featured,omitempty"`
	Levels        []LevelVersion   `json:"levels"`
}

type GalleryItem struct {
	Kind         string           `json:"kind"`
	ID           string           `json:"id"`
	LocalID      string           `json:"localId,omitempty"`
	OwnerID      string           `json:"ownerId"`
	CreatorName  string           `json:"creatorName"`
	CreatorBio   string           `json:"creatorBio,omitempty"`
	AvatarPuzzle *nonogram.Puzzle `json:"avatarPuzzle,omitempty"`
	Title        string           `json:"title"`
	Description  string           `json:"description,omitempty"`
	Plays        int              `json:"plays"`
	Likes        int              `json:"likes"`
	Liked        bool             `json:"liked"`
	Owned        bool             `json:"owned"`
	Promoted     bool             `json:"promoted"`
	Completed    bool             `json:"completed,omitempty"`
	PreviewRaw   [][]string       `json:"previewPixels,omitempty"`
	Puzzle       *nonogram.Puzzle `json:"puzzle,omitempty"`
	Levels       []LevelVersion   `json:"levels,omitempty"`
	PublishedAt  string           `json:"publishedAt"`
}

type ChatMessage struct {
	ID           string           `json:"id"`
	AuthorID     string           `json:"authorId"`
	AuthorName   string           `json:"authorName"`
	AvatarPuzzle *nonogram.Puzzle `json:"avatarPuzzle,omitempty"`
	Body         string           `json:"body"`
	CreatedAt    string           `json:"createdAt"`
}

type LevelVersion struct {
	ID          string           `json:"id"`
	LevelID     string           `json:"levelId"`
	LocalID     string           `json:"localId,omitempty"`
	Version     int              `json:"version"`
	Title       string           `json:"title"`
	Description string           `json:"description,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
	Puzzle      *nonogram.Puzzle `json:"puzzle"`
	Completed   bool             `json:"completed,omitempty"`
	PublishedAt string           `json:"publishedAt"`
}

type Pack struct {
	ID          string       `json:"id"`
	OwnerID     string       `json:"ownerId,omitempty"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
	Visibility  Visibility   `json:"visibility"`
	Status      LevelStatus  `json:"status"`
	Version     int          `json:"version"`
	Items       []PackItem   `json:"items"`
	UpdatedAt   string       `json:"updatedAt"`
	Progress    PackProgress `json:"progress,omitempty"`
}

type PackItem struct {
	LevelID        string `json:"levelId"`
	LevelVersionID string `json:"levelVersionId,omitempty"`
	Position       int    `json:"position"`
}

type PackProgress struct {
	CompletedLevelIDs []string `json:"completedLevelIds,omitempty"`
}

type OfficialSubmission struct {
	ID              string           `json:"id"`
	LevelID         string           `json:"levelId"`
	LevelVersionID  string           `json:"levelVersionId"`
	OwnerID         string           `json:"ownerId,omitempty"`
	Status          SubmissionStatus `json:"status"`
	CreatorNote     string           `json:"creatorNote,omitempty"`
	ReviewerNote    string           `json:"reviewerNote,omitempty"`
	RightsConfirmed bool             `json:"rightsConfirmed"`
	CreatedAt       string           `json:"createdAt"`
	UpdatedAt       string           `json:"updatedAt"`
}

type Library struct {
	Drafts      []LevelDraft         `json:"drafts"`
	Packs       []Pack               `json:"packs"`
	Submissions []OfficialSubmission `json:"submissions"`
}

func (l *Library) UpsertDraft(draft LevelDraft) {
	for i := range l.Drafts {
		if l.Drafts[i].ID == draft.ID {
			l.Drafts[i] = draft
			return
		}
	}
	l.Drafts = append([]LevelDraft{draft}, l.Drafts...)
}

func (l *Library) Draft(id string) (*LevelDraft, bool) {
	for i := range l.Drafts {
		if l.Drafts[i].ID == id {
			return &l.Drafts[i], true
		}
	}
	return nil, false
}

func (l *Library) Pack(id string) (*Pack, bool) {
	for i := range l.Packs {
		if l.Packs[i].ID == id {
			return &l.Packs[i], true
		}
	}
	return nil, false
}

func NewDraft(id string, puzzle *nonogram.Puzzle) LevelDraft {
	now := time.Now().UTC().Format(time.RFC3339)
	return LevelDraft{
		ID:         id,
		Title:      puzzle.Title,
		Visibility: VisibilityDraft,
		Status:     LevelDraftStatus,
		Version:    1,
		UpdatedAt:  now,
		Puzzle:     puzzle,
	}
}

func (d *LevelDraft) ValidateForSave() error {
	if d == nil || d.Puzzle == nil {
		return errors.New("level has no puzzle")
	}
	if d.ID == "" {
		return errors.New("level has no id")
	}
	if !supportedSize(d.Puzzle.Width) || !supportedSize(d.Puzzle.Height) {
		return fmt.Errorf("unsupported puzzle size %dx%d", d.Puzzle.Width, d.Puzzle.Height)
	}
	if err := d.Puzzle.ParseSolution(); err != nil {
		return err
	}
	if err := validatePixelMatrix(d.Puzzle.RevealRaw, d.Puzzle.Width, d.Puzzle.Height, false); err != nil {
		return fmt.Errorf("after layer: %w", err)
	}
	if err := validatePixelMatrix(d.Puzzle.SkeletonRaw, d.Puzzle.Width, d.Puzzle.Height, true); err != nil {
		return fmt.Errorf("before layer: %w", err)
	}
	return nil
}

func (d *LevelDraft) ValidateForPublish() error {
	if err := d.ValidateForSave(); err != nil {
		return err
	}
	if strings.TrimSpace(d.Title) == "" {
		return errors.New("title is required")
	}
	if len(d.Title) > 80 {
		return errors.New("title is too long")
	}
	if len(d.Description) > 500 {
		return errors.New("description is too long")
	}
	filled := 0
	for _, row := range d.Puzzle.Solution {
		for _, cell := range row {
			if cell {
				filled++
			}
		}
	}
	if filled == 0 {
		return errors.New("after artwork is empty")
	}
	return nil
}

func (d LevelDraft) PublishedVersion(versionID string) (LevelVersion, error) {
	if err := d.ValidateForPublish(); err != nil {
		return LevelVersion{}, err
	}
	return LevelVersion{
		ID:          versionID,
		LevelID:     d.ID,
		Version:     d.Version,
		Title:       d.Title,
		Description: d.Description,
		Tags:        append([]string(nil), d.Tags...),
		Puzzle:      d.Puzzle,
		PublishedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (p Pack) Validate() error {
	if strings.TrimSpace(p.Title) == "" {
		return errors.New("pack title is required")
	}
	if len(p.Items) < 1 || len(p.Items) > MaxPackItems {
		return fmt.Errorf("packs must contain 1 to %d levels", MaxPackItems)
	}
	seen := make(map[string]bool, len(p.Items))
	for index, item := range p.Items {
		if item.LevelID == "" {
			return fmt.Errorf("pack item %d has no level", index+1)
		}
		if seen[item.LevelID] {
			return fmt.Errorf("level %s appears more than once", item.LevelID)
		}
		seen[item.LevelID] = true
	}
	return nil
}

func supportedSize(size int) bool {
	switch size {
	case 8, 10, 15, 20, 32:
		return true
	default:
		return false
	}
}

func validatePixelMatrix(matrix [][]string, width, height int, blackOnly bool) error {
	if len(matrix) != height {
		return fmt.Errorf("has %d rows, expected %d", len(matrix), height)
	}
	for y, row := range matrix {
		if len(row) != width {
			return fmt.Errorf("row %d has %d cells, expected %d", y+1, len(row), width)
		}
		for _, value := range row {
			if value == "" || value == "transparent" {
				continue
			}
			if len(value) != 9 || value[0] != '#' {
				return fmt.Errorf("contains invalid color %q", value)
			}
			if blackOnly && !strings.EqualFold(value, "#000000FF") {
				return errors.New("must contain only black or transparent pixels")
			}
		}
	}
	return nil
}

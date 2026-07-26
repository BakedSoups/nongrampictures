package community

import (
	"fmt"
	"testing"

	"github.com/BakedSoups/community_nongrams/internal/nonogram"
)

func TestDraftPublishValidation(t *testing.T) {
	puzzle := &nonogram.Puzzle{
		ID:          "draft",
		Title:       "Flower",
		Width:       8,
		Height:      8,
		SolutionRaw: rows(8, "10000000"),
		SkeletonRaw: pixels(8, "#000000FF"),
		RevealRaw:   pixels(8, "#975347FF"),
	}
	draft := NewDraft("level-1", puzzle)
	if err := draft.ValidateForPublish(); err != nil {
		t.Fatal(err)
	}
}

func TestBeforeLayerMustBeBlack(t *testing.T) {
	puzzle := &nonogram.Puzzle{
		ID:          "draft",
		Title:       "Flower",
		Width:       8,
		Height:      8,
		SolutionRaw: rows(8, "10000000"),
		SkeletonRaw: pixels(8, "#FF0000FF"),
		RevealRaw:   pixels(8, "#975347FF"),
	}
	draft := NewDraft("level-1", puzzle)
	if err := draft.ValidateForSave(); err == nil {
		t.Fatal("colored before layer passed validation")
	}
}

func TestFullyFilledArtworkCanPublish(t *testing.T) {
	puzzle := &nonogram.Puzzle{
		ID:          "draft",
		Title:       "Block",
		Width:       8,
		Height:      8,
		SolutionRaw: filledRows(8),
		SkeletonRaw: pixels(8, "#000000FF"),
		RevealRaw:   filledPixels(8, "#975347FF"),
	}
	draft := NewDraft("level-1", puzzle)
	if err := draft.ValidateForPublish(); err != nil {
		t.Fatal(err)
	}
}

func TestPackValidationAllowsMaxPackItems(t *testing.T) {
	pack := Pack{Title: "Big Pack", Items: make([]PackItem, MaxPackItems)}
	for i := range pack.Items {
		pack.Items[i] = PackItem{LevelID: fmt.Sprintf("level-%d", i)}
	}
	if err := pack.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPackValidationRejectsTooManyItems(t *testing.T) {
	pack := Pack{Title: "Too Big", Items: make([]PackItem, MaxPackItems+1)}
	for i := range pack.Items {
		pack.Items[i] = PackItem{LevelID: fmt.Sprintf("level-%d", i)}
	}
	if err := pack.Validate(); err == nil {
		t.Fatal("oversized pack passed validation")
	}
}

func rows(count int, first string) []string {
	result := make([]string, count)
	result[0] = first
	for i := 1; i < count; i++ {
		result[i] = "00000000"
	}
	return result
}

func filledRows(size int) []string {
	result := make([]string, size)
	for i := range result {
		result[i] = "11111111"
	}
	return result
}

func pixels(size int, first string) [][]string {
	result := make([][]string, size)
	for y := range result {
		result[y] = make([]string, size)
	}
	result[0][0] = first
	return result
}

func filledPixels(size int, color string) [][]string {
	result := make([][]string, size)
	for y := range result {
		result[y] = make([]string, size)
		for x := range result[y] {
			result[y][x] = color
		}
	}
	return result
}

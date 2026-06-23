package pixelpuzzle

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
)

type PuzzleJSON struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Width          int        `json:"width"`
	Height         int        `json:"height"`
	Solution       []string   `json:"solution"`
	SkeletonPixels [][]string `json:"skeletonPixels,omitempty"`
	RevealPixels   [][]string `json:"revealPixels,omitempty"`
}

type SpriteSheetOptions struct {
	ID             string
	Title          string
	Source         string
	Out            string
	TileSize       int
	AlphaThreshold uint32
	UseBackground  bool
}

func GenerateSpriteSheet(opts SpriteSheetOptions) (PuzzleJSON, error) {
	if opts.ID == "" || opts.Title == "" || opts.Source == "" || opts.Out == "" || opts.TileSize <= 0 {
		return PuzzleJSON{}, fmt.Errorf("id, title, source, out, and tile size are required")
	}

	img, err := decodeImage(opts.Source)
	if err != nil {
		return PuzzleJSON{}, err
	}
	bounds := img.Bounds()
	if bounds.Dx() < opts.TileSize*2 || bounds.Dy() < opts.TileSize {
		return PuzzleJSON{}, fmt.Errorf("%s is %dx%d, expected at least %dx%d", opts.Source, bounds.Dx(), bounds.Dy(), opts.TileSize*2, opts.TileSize)
	}

	beforeRect := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+opts.TileSize, bounds.Min.Y+opts.TileSize)
	afterRect := image.Rect(bounds.Min.X+opts.TileSize, bounds.Min.Y, bounds.Min.X+opts.TileSize*2, bounds.Min.Y+opts.TileSize)
	solution := solutionFromImageRect(img, beforeRect, opts.AlphaThreshold, opts.UseBackground)

	puzzle := PuzzleJSON{
		ID:             opts.ID,
		Title:          opts.Title,
		Width:          opts.TileSize,
		Height:         opts.TileSize,
		Solution:       solution,
		SkeletonPixels: pixelRows(img, beforeRect),
		RevealPixels:   pixelRows(img, afterRect),
	}
	if err := writePuzzleJSON(opts.Out, puzzle); err != nil {
		return PuzzleJSON{}, err
	}
	return puzzle, nil
}

func decodeImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func solutionFromImageRect(img image.Image, r image.Rectangle, alphaThreshold uint32, useBackground bool) []string {
	bgR, bgG, bgB, _ := img.At(r.Min.X, r.Min.Y).RGBA()
	hasTransparency := imageRectHasTransparency(img, r, alphaThreshold)

	rows := make([]string, 0, r.Dy())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		row := make([]byte, 0, r.Dx())
		for x := r.Min.X; x < r.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			filled := a > alphaThreshold*257
			if !hasTransparency && useBackground && closeColor(r, g, b, bgR, bgG, bgB) {
				filled = false
			}
			if filled {
				row = append(row, '1')
			} else {
				row = append(row, '0')
			}
		}
		rows = append(rows, string(row))
	}
	return rows
}

func pixelRows(img image.Image, r image.Rectangle) [][]string {
	rows := make([][]string, 0, r.Dy())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		row := make([]string, 0, r.Dx())
		for x := r.Min.X; x < r.Max.X; x++ {
			red, green, blue, alpha := img.At(x, y).RGBA()
			if alpha == 0 {
				row = append(row, "")
				continue
			}
			row = append(row, fmt.Sprintf("#%02X%02X%02X%02X", uint8(red>>8), uint8(green>>8), uint8(blue>>8), uint8(alpha>>8)))
		}
		rows = append(rows, row)
	}
	return rows
}

func writePuzzleJSON(out string, puzzle PuzzleJSON) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(puzzle, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(out, "puzzle.json"), data, 0o644)
}

func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func imageRectHasTransparency(img image.Image, r image.Rectangle, alphaThreshold uint32) bool {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a <= alphaThreshold*257 {
				return true
			}
		}
	}
	return false
}

func closeColor(r, g, b, wantR, wantG, wantB uint32) bool {
	const tolerance = 2 * 257
	return diff(r, wantR) <= tolerance && diff(g, wantG) <= tolerance && diff(b, wantB) <= tolerance
}

func diff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

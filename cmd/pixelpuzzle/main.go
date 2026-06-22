package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
)

type puzzleJSON struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Width       int      `json:"width"`
	Height      int      `json:"height"`
	Solution    []string `json:"solution"`
	SkeletonArt string   `json:"skeletonArt"`
	RevealArt   string   `json:"revealArt"`
}

func main() {
	id := flag.String("id", "", "puzzle id, e.g. l1")
	title := flag.String("title", "", "puzzle title")
	source := flag.String("source", "", "source PNG used to generate the solution and skeleton art")
	reveal := flag.String("reveal", "", "reveal PNG; defaults to source")
	out := flag.String("out", "", "output puzzle directory, e.g. assets/puzzles/l1")
	alphaThreshold := flag.Uint("alpha-threshold", 128, "alpha threshold for filled pixels, 0-255")
	useBackground := flag.Bool("background-empty", true, "when image is opaque, treat the top-left color as empty")
	flag.Parse()

	if *id == "" || *title == "" || *source == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "required flags: -id -title -source -out")
		os.Exit(2)
	}
	if *reveal == "" {
		*reveal = *source
	}

	solution, width, height, err := solutionFromPNG(*source, uint32(*alphaThreshold), *useBackground)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	skeletonPath := filepath.Join(*out, "skeleton.png")
	revealPath := filepath.Join(*out, "full_art.png")
	if err := copyFile(*source, skeletonPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := copyFile(*reveal, revealPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	puzzle := puzzleJSON{
		ID:          *id,
		Title:       *title,
		Width:       width,
		Height:      height,
		Solution:    solution,
		SkeletonArt: filepath.ToSlash(filepath.Join(*out, "skeleton.png")),
		RevealArt:   filepath.ToSlash(filepath.Join(*out, "full_art.png")),
	}
	data, err := json.MarshalIndent(puzzle, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(*out, "puzzle.json"), data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("wrote %s (%dx%d)\n", filepath.Join(*out, "puzzle.json"), width, height)
}

func solutionFromPNG(path string, alphaThreshold uint32, useBackground bool) ([]string, int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, 0, 0, err
	}
	bounds := img.Bounds()
	bgR, bgG, bgB, _ := img.At(bounds.Min.X, bounds.Min.Y).RGBA()
	hasTransparency := imageHasTransparency(img, alphaThreshold)

	rows := make([]string, 0, bounds.Dy())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		row := make([]byte, 0, bounds.Dx())
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
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
	return rows, bounds.Dx(), bounds.Dy(), nil
}

func imageHasTransparency(img image.Image, alphaThreshold uint32) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
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

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

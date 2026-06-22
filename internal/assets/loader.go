package assets

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/alex/nongrampictures/internal/nonogram"
	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed embedded/assets/puzzles/test_001/* embedded/assets/puzzles/l1/* embedded/assets/ui/*
var embeddedFiles embed.FS

type PuzzleAssets struct {
	Puzzle   *nonogram.Puzzle
	Skeleton *ebiten.Image
	Reveal   *ebiten.Image
}

type UIIcons struct {
	Home   *ebiten.Image
	Gear   *ebiten.Image
	Pencil *ebiten.Image
	Eraser *ebiten.Image
}

func LoadPuzzleAssets(puzzlePath string) (*PuzzleAssets, error) {
	puzzle, err := nonogram.LoadPuzzle(puzzlePath)
	if err != nil {
		puzzle, err = nonogram.LoadPuzzleFS(embeddedFiles, embeddedPath(puzzlePath))
		if err != nil {
			return nil, err
		}
	}

	base := filepath.Dir(puzzlePath)
	skeleton, err := loadImage(resolveAsset(base, puzzle.SkeletonArt))
	if err != nil {
		skeleton, err = loadEmbeddedImage(puzzle.SkeletonArt)
		if err != nil {
			return nil, err
		}
	}
	reveal, err := loadImage(resolveAsset(base, puzzle.RevealArt))
	if err != nil {
		reveal, err = loadEmbeddedImage(puzzle.RevealArt)
		if err != nil {
			return nil, err
		}
	}

	return &PuzzleAssets{Puzzle: puzzle, Skeleton: skeleton, Reveal: reveal}, nil
}

func LoadUIIcons() (*UIIcons, error) {
	home, err := loadImageWithFallback("assets/ui/home.png")
	if err != nil {
		return nil, err
	}
	gear, err := loadImageWithFallback("assets/ui/gear.png")
	if err != nil {
		return nil, err
	}
	pencil, err := loadImageWithFallback("assets/ui/pencil.png")
	if err != nil {
		return nil, err
	}
	eraser, err := loadImageWithFallback("assets/ui/eraser.png")
	if err != nil {
		return nil, err
	}
	return &UIIcons{Home: home, Gear: gear, Pencil: pencil, Eraser: eraser}, nil
}

func resolveAsset(base, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return filepath.Join(base, filepath.Base(path))
}

func loadImage(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

func loadEmbeddedImage(path string) (*ebiten.Image, error) {
	data, err := fs.ReadFile(embeddedFiles, embeddedPath(path))
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

func loadImageWithFallback(path string) (*ebiten.Image, error) {
	img, err := loadImage(path)
	if err == nil {
		return img, nil
	}
	return loadEmbeddedImage(path)
}

func embeddedPath(path string) string {
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")
	return filepath.ToSlash(filepath.Join("embedded", path))
}

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/alex/nongrampictures/internal/pixelpuzzle"
)

var levelFilePattern = regexp.MustCompile(`^(L[0-9]+)[-_](.+)_([0-9]+)(?:\.png)?$`)

func main() {
	levelsDir := flag.String("levels", "levels", "folder containing L1-name_16 spritesheet files")
	outRoot := flag.String("out", "assets/puzzles", "output puzzle root")
	embedRoot := flag.String("embed", "internal/assets/embedded/assets/puzzles", "embedded puzzle root to refresh")
	alphaThreshold := flag.Uint("alpha-threshold", 128, "alpha threshold for filled pixels, 0-255")
	useBackground := flag.Bool("background-empty", true, "when image is opaque, treat the top-left color as empty")
	flag.Parse()

	entries, err := os.ReadDir(*levelsDir)
	if err != nil {
		fatal(err)
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := levelFilePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		levelName := matches[1]
		artName := matches[2]
		tileSize, err := strconv.Atoi(matches[3])
		if err != nil {
			fatal(err)
		}
		id := strings.ToLower(levelName)
		source := filepath.Join(*levelsDir, entry.Name())
		out := filepath.Join(*outRoot, id)

		puzzle, err := pixelpuzzle.GenerateSpriteSheet(pixelpuzzle.SpriteSheetOptions{
			ID:             id,
			Title:          levelTitle(levelName, artName),
			Source:         source,
			Out:            out,
			TileSize:       tileSize,
			AlphaThreshold: uint32(*alphaThreshold),
			UseBackground:  *useBackground,
		})
		if err != nil {
			fatal(err)
		}

		embedOut := filepath.Join(*embedRoot, id)
		if err := os.MkdirAll(embedOut, 0o755); err != nil {
			fatal(err)
		}
		if err := pixelpuzzle.CopyFile(filepath.Join(out, "puzzle.json"), filepath.Join(embedOut, "puzzle.json")); err != nil {
			fatal(err)
		}

		count++
		fmt.Printf("generated %s from %s (%dx%d, json only)\n", filepath.Join(out, "puzzle.json"), entry.Name(), puzzle.Width, puzzle.Height)
	}

	if count == 0 {
		fatal(fmt.Errorf("no level spritesheets found in %s", *levelsDir))
	}
}

func levelTitle(levelName, artName string) string {
	words := strings.FieldsFunc(artName, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})
	for i, word := range words {
		if word == "" {
			continue
		}
		words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
	}
	return levelName + " " + strings.Join(words, " ")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

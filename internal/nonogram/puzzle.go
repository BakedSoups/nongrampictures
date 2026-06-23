package nonogram

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
)

type Puzzle struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Width       int        `json:"width"`
	Height      int        `json:"height"`
	SolutionRaw []string   `json:"solution"`
	SkeletonArt string     `json:"skeletonArt,omitempty"`
	RevealArt   string     `json:"revealArt,omitempty"`
	SkeletonRaw [][]string `json:"skeletonPixels,omitempty"`
	RevealRaw   [][]string `json:"revealPixels,omitempty"`

	Solution [][]bool `json:"-"`
}

func LoadPuzzle(path string) (*Puzzle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var p Puzzle
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if err := p.ParseSolution(); err != nil {
		return nil, err
	}
	return &p, nil
}

func LoadPuzzleFS(files fs.FS, path string) (*Puzzle, error) {
	data, err := fs.ReadFile(files, path)
	if err != nil {
		return nil, err
	}

	var p Puzzle
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	if err := p.ParseSolution(); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *Puzzle) ParseSolution() error {
	if p.Width <= 0 || p.Height <= 0 {
		return fmt.Errorf("puzzle dimensions must be positive")
	}
	if len(p.SolutionRaw) != p.Height {
		return fmt.Errorf("solution has %d rows, expected %d", len(p.SolutionRaw), p.Height)
	}

	p.Solution = make([][]bool, p.Height)
	for y, row := range p.SolutionRaw {
		if len(row) != p.Width {
			return fmt.Errorf("solution row %d has width %d, expected %d", y, len(row), p.Width)
		}
		p.Solution[y] = make([]bool, p.Width)
		for x, ch := range row {
			switch ch {
			case '0':
				p.Solution[y][x] = false
			case '1':
				p.Solution[y][x] = true
			default:
				return fmt.Errorf("solution row %d col %d has invalid value %q", y, x, ch)
			}
		}
	}
	return nil
}

# Nonogram Pictures

A small Go + Ebitengine vertical-slice prototype for a cozy, DS-era-inspired nonogram puzzle game. It loads one manual puzzle from `assets/`, generates clues from the solution, lets you draw with a right-side tool trigger, and reveals pixel art when solved.

## Run

```sh
go run ./cmd/game
```

Controls:

- Click or touch the right-side trigger to switch between fill and X-mark tools.
- Drag across cells to keep applying the selected tool.
- `F` selects fill, `X` or `M` selects X-mark, `Z` undoes, and `R` resets.

## Add A Puzzle

Create a folder under `assets/puzzles/`, then add a `puzzle.json`:

```json
{
  "id": "test_002",
  "title": "New Puzzle",
  "width": 10,
  "height": 10,
  "solution": [
    "0011110000",
    "0111111000",
    "1111111100",
    "0011110000",
    "0011110000",
    "0011110000",
    "0011110000",
    "0001100000",
    "0001100000",
    "0001100000"
  ],
  "skeletonArt": "assets/puzzles/test_002/skeleton.png",
  "revealArt": "assets/puzzles/test_002/full_art.png"
}
```

The solution strings define filled cells: `1` is filled and `0` is empty. Row and column clues are generated automatically.

For the MVP, `cmd/game/main.go` loads `assets/puzzles/test_001/puzzle.json`. Change that path to test another puzzle.

You can also generate a puzzle from pixel art:

```sh
go run ./cmd/pixelpuzzle \
  -id l1 \
  -title "Level 1" \
  -source levels/L1-cactus.png \
  -reveal levels/L1-cactus-reveal.png \
  -out assets/puzzles/l1
```

Transparent pixels become empty cells. If the PNG is fully opaque, the helper treats pixels matching the top-left color as empty.

## Replace Art

Replace these files with same-name PNGs:

- `assets/puzzles/test_001/skeleton.png`: faint preview shown during play.
- `assets/puzzles/test_001/full_art.png`: reward art shown after completion.

The game scales the images, so square pixel art works best but is not required.

## Checks

Local checks:

```sh
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
```

GitHub Actions runs formatting, tests, vet, build, and `gocritic` duplicate-code checks on every push and pull request.

## Web And Mobile

Ebitengine can target WebAssembly, Android, and iOS later. A future web build can use:

```sh
GOOS=js GOARCH=wasm go build -o static/game.wasm ./cmd/game
```

Android and iOS should be added once the MVP input, layout, and asset loading choices are stable.

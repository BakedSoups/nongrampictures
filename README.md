# Nonogram Pictures

A small Go + Ebitengine vertical-slice prototype for a cozy, DS-era-inspired nonogram puzzle game. It loads puzzle JSON from `assets/`, generates clues from the solution, lets you draw with a right-side tool trigger, and reveals pixel art when solved.

## Run

```sh
go run ./cmd/game
```

Controls:

- Click or touch the right-side trigger to switch between fill and X-mark tools.
- Drag across cells to keep applying the selected tool.
- `F` selects fill, `X` or `M` selects X-mark, `Z` undoes, and `R` resets.

## Add A Puzzle

Create one spritesheet in `levels/` named like:

```text
L3-Flower_16.png
```

The suffix is the tile size. A `_16` file should be 32x16: the left 16x16 tile is the before/line art and puzzle solution, and the right 16x16 tile is the colored reveal.

Generate puzzle JSON from every sheet in `levels/`:

```sh
go run ./cmd/genlevels
```

This writes self-contained `puzzle.json` files under `assets/puzzles/` and `internal/assets/embedded/assets/puzzles/`. No split skeleton/reveal images are generated.

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

For a local web dev loop, install `watchexec`, then run:

```sh
PORT=8000 ./scripts/dev-web.sh
```

That serves `static/`, rebuilds `static/game.wasm` when Go/assets change, and reloads the browser on localhost after the WASM file changes.

Android and iOS should be added once the MVP input, layout, and asset loading choices are stable.

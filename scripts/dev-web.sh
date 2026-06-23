#!/usr/bin/env bash
set -euo pipefail

PORT="${PORT:-8000}"
BUILD_CMD='go run ./cmd/genlevels && GOOS=js GOARCH=wasm go build -o static/game.wasm ./cmd/game'
WATCH_PATHS=(cmd internal/game internal/nonogram internal/pixelpuzzle internal/assets/loader.go assets/ui levels static/index.html)

build_game() {
  echo "generating levels and building static/game.wasm"
  sh -c "$BUILD_CMD"
}

latest_stamp() {
  find "${WATCH_PATHS[@]}" -type f \( -name '*.go' -o -name '*.png' -o -name '*.json' -o -name '*.html' -o -name '*.js' \) -printf '%T@\n' |
    sort -n |
    tail -n 1
}

build_game

python3 -m http.server "$PORT" --directory static &
server_pid=$!

cleanup() {
  kill "$server_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

echo "serving http://localhost:${PORT}"

if ! command -v watchexec >/dev/null 2>&1; then
  echo "watchexec is not installed; using built-in polling rebuilds."
  last_stamp="$(latest_stamp)"
  while kill -0 "$server_pid" 2>/dev/null; do
    sleep 1
    next_stamp="$(latest_stamp)"
    if [[ "$next_stamp" != "$last_stamp" ]]; then
      last_stamp="$next_stamp"
      if ! build_game; then
        echo "build failed; keeping the previous static/game.wasm"
      fi
    fi
  done
  exit 0
fi

watch_args=()
for path in "${WATCH_PATHS[@]}"; do
  watch_args+=(--watch "$path")
done

watchexec \
  "${watch_args[@]}" \
  --exts go,png,json,html,js \
  --debounce 250ms \
  -- sh -c "$BUILD_CMD"

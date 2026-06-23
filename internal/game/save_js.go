//go:build js

package game

import (
	"strconv"
	"syscall/js"
	"time"
)

const saveKeyPrefix = "pixaross.best."

func loadSavedBest(levelID string) time.Duration {
	storage := js.Global().Get("localStorage")
	if storage.IsUndefined() || storage.IsNull() {
		return 0
	}
	raw := storage.Call("getItem", saveKeyPrefix+levelID)
	if raw.IsNull() || raw.IsUndefined() {
		return 0
	}
	ms, err := strconv.Atoi(raw.String())
	if err != nil || ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}

func saveBest(levelID string, best time.Duration) {
	storage := js.Global().Get("localStorage")
	if storage.IsUndefined() || storage.IsNull() || best <= 0 {
		return
	}
	storage.Call("setItem", saveKeyPrefix+levelID, strconv.Itoa(int(best/time.Millisecond)))
}

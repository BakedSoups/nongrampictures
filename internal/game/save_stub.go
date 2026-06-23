//go:build !js

package game

import "time"

func loadSavedBest(string) time.Duration {
	return 0
}

func saveBest(string, time.Duration) {}

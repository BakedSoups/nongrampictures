//go:build js

package game

import "syscall/js"

func setWebMusicMuted(muted bool) {
	fn := js.Global().Get("setMusicMuted")
	if fn.Type() == js.TypeFunction {
		fn.Invoke(muted)
	}
}

func playWebSFX(name string) {
	fn := js.Global().Get("playSFX")
	if fn.Type() == js.TypeFunction {
		fn.Invoke(name)
	}
}

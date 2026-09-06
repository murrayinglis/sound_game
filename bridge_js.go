//go:build js

package main

import (
	"encoding/json"
	"syscall/js"
)

// bridge publishes the controls to the page, which draws the real UI in HTML.
// Callbacks arrive on the JS event loop rather than Ebiten's goroutine, so
// everything they touch goes through mu.
//
// The page waits for window.soundgame to appear rather than being called back,
// because this runs before RunGame and so before Ebiten's canvas exists.
func bridge() {
	js.Global().Set("soundgame", map[string]any{
		"instruments": fn(func([]js.Value) any {
			type entry struct {
				Name  string `json:"name"`
				Color string `json:"color"`
			}
			list := make([]entry, len(cfg.Instruments))
			for i, in := range cfg.Instruments {
				list[i] = entry{shortName(in.File), in.Color}
			}
			return marshal(list)
		}),
		"state": fn(func([]js.Value) any {
			mu.Lock()
			defer mu.Unlock()
			return marshal(map[string]any{
				"cur":  curInst,
				"grid": showGrid,
				// Inlined rather than calling bpm(), which would deadlock on mu.
				"bpm": float64(cfg.Steps) * 60 / sweepSec,
			})
		}),
		"setInstrument": fn(func(a []js.Value) any { setInstrument(a[0].Int()); return nil }),
		"setBPM":        fn(func(a []js.Value) any { setBPM(a[0].Float()); return nil }),
		"setGrid":       fn(func(a []js.Value) any { setGrid(a[0].Bool()); return nil }),
		"clear": fn(func([]js.Value) any {
			mu.Lock()
			clearCanvas()
			mu.Unlock()
			return nil
		}),
	})
}

func fn(f func([]js.Value) any) js.Func {
	return js.FuncOf(func(_ js.Value, args []js.Value) any { return f(args) })
}

func marshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(b)
}

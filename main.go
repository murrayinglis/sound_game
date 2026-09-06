// Draw on the canvas; a playhead sweeps left to right turning what it crosses
// into sound. Y is pitch, X is time.
package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

func main() {
	loadConfig("config.json")
	loadBackground() // sets W and H, so it has to run before the canvas exists
	initCanvas()
	loadPalette()
	loadTables()

	p, err := audio.NewContext(sampleRate).NewPlayer(&synth{sweep: sweepSec})
	if err != nil {
		log.Fatal(err)
	}
	p.Play() // on wasm this stays silent until the first click; that's the browser

	ebiten.SetWindowSize(W, H+pickerH)
	ebiten.SetWindowTitle("sound game: draw, G grid, -/= tempo, C clear")
	g := &Game{canvas: ebiten.NewImage(W, H), player: p, grid: gridDefault, sweep: sweepSec, start: time.Now()}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

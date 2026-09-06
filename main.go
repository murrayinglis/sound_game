// Draw on the canvas; a playhead sweeps left to right turning what it crosses
// into sound. Y is pitch, X is time.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

func main() {
	clearCanvas()

	p, err := audio.NewContext(sampleRate).NewPlayer(&synth{})
	if err != nil {
		log.Fatal(err)
	}
	p.Play() // on wasm this stays silent until the first click; that's the browser

	ebiten.SetWindowSize(W, H)
	ebiten.SetWindowTitle("sound game: draw, C to clear")
	if err := ebiten.RunGame(&Game{canvas: ebiten.NewImage(W, H), player: p}); err != nil {
		log.Fatal(err)
	}
}

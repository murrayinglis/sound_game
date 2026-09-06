package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	canvas       *ebiten.Image
	player       *audio.Player
	lastX, lastY int
	drawing      bool
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		mu.Lock()
		clearCanvas()
		mu.Unlock()
	}
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.drawing = false
		return nil
	}
	x, y := ebiten.CursorPosition()
	mu.Lock()
	if g.drawing {
		paintLine(g.lastX, g.lastY, x, y)
	} else {
		paint(x, y)
	}
	mu.Unlock()
	g.lastX, g.lastY, g.drawing = x, y, true
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	mu.Lock()
	g.canvas.WritePixels(pix)
	mu.Unlock()
	screen.DrawImage(g.canvas, nil)

	// Position() is what is audible right now, not what the synth has already
	// written into the buffer, so the line sits on the column you can hear.
	ph := float32(math.Mod(g.player.Position().Seconds()/sweepSec, 1) * W)
	vector.StrokeLine(screen, ph, 0, ph, H, 1, color.RGBA{255, 90, 90, 255}, false)
}

func (g *Game) Layout(int, int) (int, int) { return W, H }

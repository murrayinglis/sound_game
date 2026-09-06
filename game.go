package main

import (
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// The picker lives in its own strip below the canvas rather than floating over
// it, so there is no region where a click might mean either "draw" or "choose".
const pickerH = 30

// Listed rather than Key1+i: the Key constants are not promised to be contiguous.
var pickKeys = []ebiten.Key{
	ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4,
	ebiten.Key5, ebiten.Key6, ebiten.Key7, ebiten.Key8,
}

type Game struct {
	canvas       *ebiten.Image
	player       *audio.Player
	cur          int // selected instrument
	lastX, lastY int
	drawing      bool
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		mu.Lock()
		clearCanvas()
		mu.Unlock()
	}
	for i := range min(len(cfg.Instruments), len(pickKeys)) {
		if inpututil.IsKeyJustPressed(pickKeys[i]) {
			g.cur = i
		}
	}

	x, y := ebiten.CursorPosition()
	if y >= H { // in the picker strip
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if i := x * len(cfg.Instruments) / W; i >= 0 && i < len(cfg.Instruments) {
				g.cur = i
			}
		}
		g.drawing = false
		return nil
	}
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.drawing = false
		return nil
	}

	mu.Lock()
	if g.drawing {
		paintLine(g.lastX, g.lastY, x, y, palette[g.cur])
	} else {
		paint(x, y, palette[g.cur])
	}
	mu.Unlock()
	g.lastX, g.lastY, g.drawing = x, y, true
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Position() is what is audible right now, not what the synth has already
	// written into the buffer, so the line sits on the column you can hear.
	ph := math.Mod(g.player.Position().Seconds()/sweepSec, 1) * W

	var runs [maxVoices]hit
	mu.Lock()
	g.canvas.WritePixels(pix)
	// The same detection the synth does, but at the audible column rather than
	// the write cursor, so a dot marks exactly what you can hear right now.
	n := findRuns(min(int(ph), W-1), &runs)
	mu.Unlock()

	screen.DrawImage(g.canvas, nil)
	x := float32(ph)
	vector.StrokeLine(screen, x, 0, x, H, 1, color.RGBA{255, 90, 90, 255}, false)
	for i := range n {
		y := float32(runs[i].y)
		vector.FillCircle(screen, x, y, 5, color.RGBA{255, 255, 255, 255}, true)
		vector.FillCircle(screen, x, y, 3.5, rgb(palette[runs[i].inst]), true)
	}
	g.drawPicker(screen)
}

func (g *Game) drawPicker(screen *ebiten.Image) {
	w := float32(W) / float32(len(cfg.Instruments))
	for i, in := range cfg.Instruments {
		x := float32(i) * w
		vector.FillRect(screen, x, H, w, pickerH, rgb(palette[i]), false)
		if i == g.cur {
			// Selected: a white bar along the top edge of the swatch.
			vector.FillRect(screen, x, H, w, 4, color.RGBA{255, 255, 255, 255}, false)
		}
		ebitenutil.DebugPrintAt(screen, shortName(in.File), int(x)+8, H+12)
	}
}

func rgb(c [3]byte) color.RGBA { return color.RGBA{c[0], c[1], c[2], 255} }

// shortName turns "AKWF_clarinett_0001.wav" into "clarinett".
func shortName(f string) string {
	s := strings.TrimSuffix(strings.TrimPrefix(f, "AKWF_"), ".wav")
	if i := strings.LastIndex(s, "_"); i > 0 {
		s = s[:i]
	}
	return s
}

func (g *Game) Layout(int, int) (int, int) { return W, H + pickerH }

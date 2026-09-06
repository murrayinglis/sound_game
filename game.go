package main

import (
	"fmt"
	"image/color"
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
	grid         bool

	// This pass's tempo and when it started, in playback seconds. The renderer
	// walks its own pass boundary forward exactly as the synth does, so both
	// adopt a new tempo at the same point in the music despite the buffer lag.
	sweep, startSec float64
}

// held reports a press and then repeats while the key stays down, so the tempo
// can be swept by holding rather than tapping.
func held(k ebiten.Key) bool {
	d := inpututil.KeyPressDuration(k)
	return d == 1 || (d > 30 && d%4 == 0)
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		mu.Lock()
		clearCanvas()
		mu.Unlock()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		g.grid = !g.grid
	}
	if held(ebiten.KeyMinus) {
		setBPM(bpm() - 5)
	}
	if held(ebiten.KeyEqual) {
		setBPM(bpm() + 5)
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
	t := g.player.Position().Seconds()
	for t-g.startSec >= g.sweep {
		g.startSec += g.sweep
		g.sweep = tempo()
	}
	ph := min((t-g.startSec)/g.sweep*W, W-1)

	step := min(int(ph*float64(cfg.Steps)/W), cfg.Steps-1)

	var runs [maxVoices]hit
	mu.Lock()
	g.canvas.WritePixels(pix)
	// The same column the synth sampled for this block, so the dots sit on the
	// pixels you can actually hear rather than wherever the playhead has got to.
	n := findRuns(stepCol(step), &runs)
	mu.Unlock()

	screen.DrawImage(g.canvas, nil)
	if g.grid {
		g.drawGrid(screen, step)
	}

	vector.StrokeLine(screen, float32(ph), 0, float32(ph), H, 1, color.RGBA{255, 90, 90, 255}, false)
	x := float32(stepCol(step))
	for i := range n {
		y := float32(runs[i].y)
		vector.FillCircle(screen, x, y, 5, color.RGBA{255, 255, 255, 255}, true)
		vector.FillCircle(screen, x, y, 3.5, rgb(palette[runs[i].inst]), true)
	}
	g.drawPicker(screen)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.0f bpm", bpm()), 4, 4)
}

// drawGrid marks the block boundaries — without them you can't tell where a
// note will start — and shades the block currently sounding.
func (g *Game) drawGrid(screen *ebiten.Image, step int) {
	w := float32(W) / float32(cfg.Steps)
	vector.FillRect(screen, float32(step)*w, 0, w, H, color.RGBA{255, 255, 255, 14}, false)
	for i := 1; i < cfg.Steps; i++ {
		x := float32(i) * w
		vector.StrokeLine(screen, x, 0, x, H, 1, color.RGBA{255, 255, 255, 26}, false)
	}
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

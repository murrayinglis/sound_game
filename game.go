package main

import (
	"image/color"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// How far the background is dimmed so strokes stay readable over it. 0 shows the
// gif untouched and makes the paler instruments hard to see; 255 hides it.
const scrim = 40

// UI state, kept out of Game because the page's controls reach it from the JS
// event loop rather than from Ebiten's goroutine. Guarded by mu.
var (
	curInst  int
	showGrid bool // set from the config at startup
)

func setInstrument(i int) {
	if i < 0 || i >= len(cfg.Instruments) {
		return
	}
	mu.Lock()
	curInst = i
	mu.Unlock()
}

func setGrid(on bool) {
	mu.Lock()
	showGrid = on
	mu.Unlock()
}

// Listed rather than Key1+i: the Key constants are not promised to be contiguous.
var pickKeys = []ebiten.Key{
	ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4,
	ebiten.Key5, ebiten.Key6, ebiten.Key7, ebiten.Key8,
}

type Game struct {
	canvas       *ebiten.Image
	player       *audio.Player
	lastX, lastY int
	drawing      bool

	// This pass's tempo and when it started, in playback seconds. The renderer
	// walks its own pass boundary forward exactly as the synth does, so both
	// adopt a new tempo at the same point in the music despite the buffer lag.
	sweep, startSec float64
	start           time.Time // wall clock, for the background animation
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
		mu.Lock()
		showGrid = !showGrid
		mu.Unlock()
	}
	if held(ebiten.KeyMinus) {
		setBPM(bpm() - 5)
	}
	if held(ebiten.KeyEqual) {
		setBPM(bpm() + 5)
	}
	for i := range min(len(cfg.Instruments), len(pickKeys)) {
		if inpututil.IsKeyJustPressed(pickKeys[i]) {
			setInstrument(i)
		}
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.drawing = false
		return nil
	}
	x, y := ebiten.CursorPosition()
	if x < 0 || x >= W || y < 0 || y >= H {
		g.drawing = false
		return nil
	}

	mu.Lock()
	c := palette[curInst]
	if g.drawing {
		paintLine(g.lastX, g.lastY, x, y, c)
	} else {
		paint(x, y, c)
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
	ph := min((t-g.startSec)/g.sweep*float64(W), float64(W-1))

	step := min(int(ph*float64(cfg.Steps)/float64(W)), cfg.Steps-1)

	var runs [maxVoices]hit
	mu.Lock()
	g.canvas.WritePixels(pix)
	// Sampled at the live column, not the block's, so the dots ride along the
	// stroke as it's drawn. They track where the playhead is touching the line;
	// the pitch you hear still comes from the block. On a sloped stroke the two
	// deliberately disagree, and the dot is showing the drawing, not the note.
	n := findRuns(int(ph), &runs)
	grid := showGrid
	mu.Unlock()

	screen.DrawImage(bgFrame(time.Since(g.start)), nil)
	vector.FillRect(screen, 0, 0, float32(W), float32(H), color.RGBA{0, 0, 0, scrim}, false)
	screen.DrawImage(g.canvas, nil)
	if grid {
		g.drawGrid(screen, step)
	}

	x := float32(ph)
	vector.StrokeLine(screen, x, 0, x, float32(H), 1, color.RGBA{255, 90, 90, 255}, false)
	for i := range n {
		y := float32(runs[i].y)
		vector.FillCircle(screen, x, y, 5, color.RGBA{255, 255, 255, 255}, true)
		vector.FillCircle(screen, x, y, 3.5, rgb(palette[runs[i].inst]), true)
	}
}

// drawGrid marks the block boundaries, without which you can't tell where a
// note will start, and shades the block currently sounding.
func (g *Game) drawGrid(screen *ebiten.Image, step int) {
	w := float32(W) / float32(cfg.Steps)
	vector.FillRect(screen, float32(step)*w, 0, w, float32(H), color.RGBA{255, 255, 255, 14}, false)
	for i := 1; i < cfg.Steps; i++ {
		x := float32(i) * w
		vector.StrokeLine(screen, x, 0, x, float32(H), 1, color.RGBA{255, 255, 255, 40}, false)
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

func (g *Game) Layout(int, int) (int, int) { return W, H }

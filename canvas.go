package main

import (
	"fmt"
	"log"
	"math"
	"sync"
)

const (
	W, H  = 800, 400
	brush = 1 // pencil radius in pixels
)

var (
	mu  sync.Mutex
	pix = make([]byte, W*H*4) // RGBA; this is the canvas AND the score
	bg  = [3]byte{18, 18, 26}

	// palette[i] is the colour instrument i draws in, parsed from the config.
	// The canvas stores no instrument index of its own — the colour of a pixel
	// is the index, which is why the config forbids two sharing one.
	palette [][3]byte
)

func loadPalette() {
	for _, in := range cfg.Instruments {
		var c [3]byte
		if _, err := fmt.Sscanf(in.Color, "#%02x%02x%02x", &c[0], &c[1], &c[2]); err != nil {
			log.Fatalf("config: colour %q for %s: %v", in.Color, in.File, err)
		}
		if c == bg {
			log.Fatalf("config: colour %q for %s is the background colour", in.Color, in.File)
		}
		palette = append(palette, c)
	}
}

func clearCanvas() {
	for i := 0; i < len(pix); i += 4 {
		pix[i], pix[i+1], pix[i+2], pix[i+3] = bg[0], bg[1], bg[2], 255
	}
}

// instAt returns the instrument drawn at a pixel, or -1 for bare canvas.
func instAt(x, y int) int {
	i := (y*W + x) * 4
	c := [3]byte{pix[i], pix[i+1], pix[i+2]}
	if c == bg {
		return -1
	}
	for k, p := range palette {
		if c == p {
			return k
		}
	}
	return 0 // a colour no longer in the config: play it with the first instrument
}

func paint(x, y int, c [3]byte) {
	for dy := -brush; dy <= brush; dy++ {
		for dx := -brush; dx <= brush; dx++ {
			if dx*dx+dy*dy > brush*brush {
				continue
			}
			px, py := x+dx, y+dy
			if px < 0 || px >= W || py < 0 || py >= H {
				continue
			}
			i := (py*W + px) * 4
			pix[i], pix[i+1], pix[i+2] = c[0], c[1], c[2]
		}
	}
}

// paintLine interpolates between mouse samples; without it fast strokes are dotted.
func paintLine(x0, y0, x1, y1 int, c [3]byte) {
	n := max(abs(x1-x0), abs(y1-y0))
	if n == 0 {
		paint(x0, y0, c)
		return
	}
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		paint(int(math.Round(float64(x0)+t*float64(x1-x0))),
			int(math.Round(float64(y0)+t*float64(y1-y0))), c)
	}
}

func abs(a int) int { return max(a, -a) }

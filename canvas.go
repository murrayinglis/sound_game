package main

import (
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
	bg  = [4]byte{18, 18, 26, 255}
)

func clearCanvas() {
	for i := 0; i < len(pix); i += 4 {
		copy(pix[i:], bg[:])
	}
}

func paint(x, y int) {
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
			pix[i], pix[i+1], pix[i+2] = 255, 255, 255
		}
	}
}

// paintLine interpolates between mouse samples; without it fast strokes are dotted.
func paintLine(x0, y0, x1, y1 int) {
	n := max(abs(x1-x0), abs(y1-y0))
	if n == 0 {
		paint(x0, y0)
		return
	}
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		paint(int(math.Round(float64(x0)+t*float64(x1-x0))),
			int(math.Round(float64(y0)+t*float64(y1-y0))))
	}
}

func abs(a int) int { return max(a, -a) }

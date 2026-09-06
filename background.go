package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/draw"
	"image/gif"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed frutiger.gif
var gifData []byte

var (
	bgFrames []*ebiten.Image
	bgUntil  []time.Duration // cumulative end time of each frame
	bgLoop   time.Duration
)

// loadBackground decodes the gif and sizes the canvas to it. Frames are
// composited up front rather than per draw: a gif frame is often only the patch
// that changed, so each one has to be painted over the ones before it.
//
// ponytail: handles the two disposal methods that appear in practice, leaving
// the frame in place or clearing its rect afterwards. DisposalPrevious would
// need the canvas snapshotted before each frame; add it if one ever needs it.
func loadBackground() {
	g, err := gif.DecodeAll(bytes.NewReader(gifData))
	if err != nil {
		log.Fatalf("background: %v", err)
	}

	W, H = g.Config.Width, g.Config.Height
	if W == 0 || H == 0 { // some encoders omit the logical screen size
		b := g.Image[0].Bounds()
		W, H = b.Dx(), b.Dy()
	}

	acc := image.NewRGBA(image.Rect(0, 0, W, H))
	for i, src := range g.Image {
		draw.Draw(acc, src.Bounds(), src, src.Bounds().Min, draw.Over)

		snap := image.NewRGBA(acc.Rect)
		copy(snap.Pix, acc.Pix)
		bgFrames = append(bgFrames, ebiten.NewImageFromImage(snap))

		d := time.Duration(g.Delay[i]) * 10 * time.Millisecond
		if d <= 0 {
			d = 100 * time.Millisecond // what browsers do with a zero delay
		}
		bgLoop += d
		bgUntil = append(bgUntil, bgLoop)

		if i < len(g.Disposal) && g.Disposal[i] == gif.DisposalBackground {
			draw.Draw(acc, src.Bounds(), image.Transparent, image.Point{}, draw.Src)
		}
	}
	log.Printf("background: %dx%d, %d frames, %v loop", W, H, len(bgFrames), bgLoop)
}

func bgFrame(elapsed time.Duration) *ebiten.Image {
	if bgLoop <= 0 {
		return bgFrames[0]
	}
	t := elapsed % bgLoop
	for i, until := range bgUntil {
		if t < until {
			return bgFrames[i]
		}
	}
	return bgFrames[len(bgFrames)-1]
}

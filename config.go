package main

import (
	"encoding/json"
	"log"
	"math"
	"os"

	"github.com/go-playground/validator/v10"
)

// Instrument pairs a waveform with the colour its strokes are drawn in. The
// colour is not decoration: it is how the synth knows which waveform to play a
// stroke with, so two instruments must never share one.
type Instrument struct {
	File  string `json:"file" validate:"required,endswith=.wav"`
	Color string `json:"color" validate:"required,hexcolor,len=7"`
}

// Config is read from config.json at startup. Fields absent from the file keep
// the defaults below. A missing file is not an error: that's the normal case
// for the wasm build, which has no filesystem to read.
type Config struct {
	// Beats per minute. One beat is one block of the canvas.
	BPM float64 `json:"bpm" validate:"gt=0,lte=300"`
	// How many blocks the canvas is divided into left to right. Everything drawn
	// inside a block sounds as one note, held for the whole beat.
	Steps int `json:"steps" validate:"gte=1,lte=64"`
	// Hz at the bottom of the canvas.
	Root float64 `json:"root" validate:"gt=0"`
	// How far above the root the top of the canvas reaches.
	Octaves int `json:"octaves" validate:"gte=1,lte=8"`
	// One of the names in scales; see scale.go.
	Scale string `json:"scale" validate:"required,scalename"`
	// What the picker offers, left to right. Capped at 8 so the strip stays legible.
	// unique=Color matters: a repeated colour would silently play one
	// instrument's strokes with another's waveform.
	Instruments []Instrument `json:"instruments" validate:"required,min=1,max=8,unique=Color,dive"`
}

var cfg = Config{
	BPM:     100,
	Steps:   16,
	Root:    110, // A2
	Octaves: 3,
	Scale:   "minor_pentatonic",
	Instruments: []Instrument{
		{"AKWF_clarinett_0001.wav", "#7ec8ff"},
		{"AKWF_epiano_0001.wav", "#ffd166"},
		{"AKWF_flute_0004.wav", "#8ce99a"},
		{"AKWF_hvoice_0001.wav", "#ff8fab"},
	},
}

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	// Registered rather than a literal oneof= list, so the tag can't drift out of
	// sync with the scales map.
	v.RegisterValidation("scalename", func(fl validator.FieldLevel) bool {
		_, ok := scales[fl.Field().String()]
		return ok
	})
	// Root and Octaves are only wrong in combination, so this one is struct-level:
	// a range topping out above Nyquist folds back down the spectrum as aliasing.
	v.RegisterStructValidation(func(sl validator.StructLevel) {
		c := sl.Current().Interface().(Config)
		if c.Root*math.Pow(2, float64(c.Octaves)) > sampleRate/2 {
			sl.ReportError(c.Octaves, "Octaves", "Octaves", "nyquist", "")
		}
	}, Config{})
	return v
}

// sweepSec is how long the playhead takes to cross the canvas, derived from the
// tempo rather than set directly: one block is one beat.
//
// It is the *pending* tempo. Changing it mid-pass would stretch the pass already
// underway and slide the playhead out from under the sound, so the synth and the
// renderer each take a copy at their own wrap point instead. Guarded by mu
// because the audio goroutine reads it.
var sweepSec float64

func tempo() float64 {
	mu.Lock()
	defer mu.Unlock()
	return sweepSec
}

func bpm() float64 { return float64(cfg.Steps) * 60 / tempo() }

// setBPM takes effect when the playhead next wraps, not immediately.
func setBPM(b float64) {
	b = min(max(b, 20), 300)
	mu.Lock()
	sweepSec = float64(cfg.Steps) * 60 / b
	mu.Unlock()
}

func loadConfig(path string) {
	// Unmarshalling over the populated struct means the file only has to carry
	// what it wants to change. A missing file is fine: that's the wasm build,
	// which has no filesystem, running on the defaults above.
	if b, err := os.ReadFile(path); err != nil {
		log.Printf("config: using defaults (%v)", err)
	} else if err := json.Unmarshal(b, &cfg); err != nil {
		log.Fatalf("config: %s is not valid JSON: %v", path, err)
	} else if err := validate.Struct(cfg); err != nil {
		log.Fatalf("config: %v\nscale must be one of: %v", err, scaleNames())
	}

	notes = scales[cfg.Scale]
	sweepSec = float64(cfg.Steps) * 60 / cfg.BPM
	log.Printf("config: %.0f bpm, %d steps (%.1fs a pass), %s from %.0fHz over %d octaves",
		cfg.BPM, cfg.Steps, sweepSec, cfg.Scale, cfg.Root, cfg.Octaves)
}

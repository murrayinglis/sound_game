package main

import (
	"encoding/json"
	"log"
	"math"
	"os"

	"github.com/go-playground/validator/v10"
)

// Config is read from config.json at startup. Fields absent from the file keep
// the defaults below, and a missing file is not an error — that's the normal
// case for the wasm build, which has no filesystem to read.
type Config struct {
	// Hz at the bottom of the canvas.
	Root float64 `json:"root" validate:"gt=0"`
	// How far above the root the top of the canvas reaches.
	Octaves int `json:"octaves" validate:"gte=1,lte=8"`
	// Semitone offsets from the root, one octave; the pattern repeats upward.
	Scale []int `json:"scale" validate:"required,min=1,dive,gte=0,lte=48"`
	// A .wav filename from wave/.
	Instrument string `json:"instrument" validate:"required,endswith=.wav"`
}

var cfg = Config{
	Root:       110, // A2
	Octaves:    3,
	Scale:      []int{0, 3, 5, 7, 10}, // minor pentatonic
	Instrument: "AKWF_clarinett_0001.wav",
}

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
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

func loadConfig(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Printf("config: using defaults (%v)", err)
		return
	}
	// Unmarshalling over the populated struct means the file only has to carry
	// what it wants to change.
	if err := json.Unmarshal(b, &cfg); err != nil {
		log.Fatalf("config: %s is not valid JSON: %v", path, err)
	}
	if err := validate.Struct(cfg); err != nil {
		log.Fatalf("config: %v", err)
	}
	log.Printf("config: root %.0fHz, %d octaves, scale %v, %s",
		cfg.Root, cfg.Octaves, cfg.Scale, cfg.Instrument)
}

package main

import (
	"maps"
	"math"
	"slices"
)

// The canvas height is divided into scale degrees rather than raw Hz, so a note
// is always in tune. Each entry is semitone offsets from the root for one
// octave; the pattern repeats upward.
var scales = map[string][]int{
	"major":            {0, 2, 4, 5, 7, 9, 11},
	"minor":            {0, 2, 3, 5, 7, 8, 10},
	"harmonic_minor":   {0, 2, 3, 5, 7, 8, 11},
	"dorian":           {0, 2, 3, 5, 7, 9, 10},
	"major_pentatonic": {0, 2, 4, 7, 9},
	"minor_pentatonic": {0, 3, 5, 7, 10},
	"blues":            {0, 3, 5, 6, 7, 10},
	"whole_tone":       {0, 2, 4, 6, 8, 10},
	"chromatic":        {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
}

func scaleNames() []string { return slices.Sorted(maps.Keys(scales)) }

// notes is the chosen scale, resolved from its name once the config is read.
// Fewer notes means fatter, more forgiving bands to draw in, and less chance
// that two strokes sounding together clash.
var notes []int

// degreeAt converts a row to a fractional scale degree, 0 at the bottom.
func degreeAt(y float64) float64 {
	return (1 - y/H) * float64(len(notes)*cfg.Octaves)
}

// freqOfDegree takes fractional degrees so a slide between two notes is smooth.
// Interpolating in degree space rather than Hz means a slide tracks the scale's
// own spacing, easing through wide intervals and hurrying through narrow ones.
func freqOfDegree(d float64) float64 {
	i := math.Floor(d)
	f := d - i
	lo := semitone(int(i))
	return cfg.Root * math.Pow(2, (lo+(semitone(int(i)+1)-lo)*f)/12)
}

// semitone is the pitch of degree i above the root, carrying the octave as the
// index runs past the end of the scale.
func semitone(i int) float64 {
	n := len(notes)
	oct, k := i/n, i%n
	if k < 0 {
		k, oct = k+n, oct-1
	}
	return float64(notes[k] + 12*oct)
}

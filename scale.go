package main

import "math"

// The canvas height is divided into scale degrees rather than raw Hz, so a held
// line sits exactly in tune. See config.json for the scale, root and range.

// degreeAt converts a row to a fractional scale degree, 0 at the bottom.
func degreeAt(y float64) float64 {
	return (1 - y/H) * float64(len(cfg.Scale)*cfg.Octaves)
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
	n := len(cfg.Scale)
	oct, k := i/n, i%n
	if k < 0 {
		k, oct = k+n, oct-1
	}
	return float64(cfg.Scale[k] + 12*oct)
}

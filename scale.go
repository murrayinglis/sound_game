package main

import "math"

// The canvas height is divided into scale degrees rather than raw Hz, so a held
// line sits exactly in tune. Minor pentatonic by default: with up to 8 strokes
// sounding at once, a scale with no dissonant intervals means any scribble works.
//
// Swap for something else if you want more notes and more ways to clash:
//
//	{0, 2, 4, 7, 9}          major pentatonic
//	{0, 2, 3, 5, 7, 8, 10}   natural minor
//	{0, 2, 4, 5, 7, 9, 11}   major
var scale = []int{0, 3, 5, 7, 10} // minor pentatonic

const (
	root    = 110.0 // A2, the pitch at the bottom of the canvas
	octaves = 4     // ...so the top is 1760Hz
)

// degreeAt converts a row to a fractional scale degree, 0 at the bottom.
func degreeAt(y float64) float64 {
	return (1 - y/H) * float64(len(scale)*octaves)
}

// freqOfDegree takes fractional degrees so a slide between two notes is smooth.
// Interpolating in degree space rather than Hz means a slide tracks the scale's
// own spacing, easing through wide intervals and hurrying through narrow ones.
func freqOfDegree(d float64) float64 {
	i := math.Floor(d)
	f := d - i
	lo := semitone(int(i))
	return root * math.Pow(2, (lo+(semitone(int(i)+1)-lo)*f)/12)
}

// semitone is the pitch of degree i above the root, carrying the octave as the
// index runs past the end of the scale.
func semitone(i int) float64 {
	n := len(scale)
	oct, k := i/n, i%n
	if k < 0 {
		k, oct = k+n, oct-1
	}
	return float64(scale[k] + 12*oct)
}

package main

import "math"

const (
	sampleRate = 44100
	sweepSec   = 4.0 // time for the playhead to cross the canvas
	maxVoices  = 8
	fMin, fMax = 80.0, 2000.0

	// Tuning knobs. ctrlSamples is how often the synth re-reads the canvas
	// (~5ms); glide and ampRate are one-pole smoothing coefficients per sample.
	// Raising glide makes pitch track the drawing more literally at the cost of
	// zipper noise; lowering ampRate softens note onsets.
	ctrlSamples = 220
	glide       = 0.0015
	ampRate     = 0.0008
)

// freqAt maps a row to a pitch exponentially, so equal vertical distances are
// equal musical intervals. Deliberately unquantized: a curve should glissando.
func freqAt(y int) float64 {
	return fMin * math.Pow(fMax/fMin, 1-float64(y)/H)
}

// findRuns returns the centre row of each contiguous vertical stroke in a
// column, one voice per run. Caller holds mu.
// ponytail: voices are assigned to runs top-to-bottom, so crossing strokes swap
// voices mid-glide. Match by nearest previous centre if that ever sounds wrong.
func findRuns(col int, out *[maxVoices]int) int {
	n, start := 0, -1
	for y := range H {
		on := pix[(y*W+col)*4] > 128
		switch {
		case on && start < 0:
			start = y
		case !on && start >= 0:
			out[n] = (start + y - 1) / 2
			n++
			start = -1
			if n == maxVoices {
				return n
			}
		}
	}
	if start >= 0 {
		out[n] = (start + H - 1) / 2
		n++
	}
	return n
}

type voice struct{ phase, freq, amp, target, gain float64 }

type synth struct {
	voices [maxVoices]voice
	acc    int
	pos    float64 // local playhead, published to the renderer once per Read
}

// The synth is the clock: driving the playhead from sample count keeps picture
// and sound locked together, which a 60fps Update would slowly drift away from.
func (s *synth) Read(buf []byte) (int, error) {
	n := len(buf) / 4 * 4
	const step = W / (sweepSec * sampleRate)

	for i := 0; i < n; i += 4 {
		if s.acc <= 0 {
			s.control()
			s.acc = ctrlSamples
		}
		s.acc--

		sum := 0.0
		for v := range s.voices {
			p := &s.voices[v]
			p.freq += (p.target - p.freq) * glide
			p.amp += (p.gain - p.amp) * ampRate
			if p.amp < 1e-4 {
				continue
			}
			p.phase += 2 * math.Pi * p.freq / sampleRate
			if p.phase > 2*math.Pi {
				p.phase -= 2 * math.Pi
			}
			sum += math.Sin(p.phase) * p.amp
		}

		// tanh instead of a mixer: soft-clips 8 voices without going quiet at 1.
		v := int16(math.Tanh(sum*0.35) * 26000)
		buf[i], buf[i+1] = byte(v), byte(v>>8)
		buf[i+2], buf[i+3] = byte(v), byte(v>>8)

		s.pos += step
		if s.pos >= W {
			s.pos -= W
		}
	}
	return n, nil
}

func (s *synth) control() {
	var runs [maxVoices]int
	mu.Lock()
	n := findRuns(int(s.pos), &runs)
	mu.Unlock()

	for i := range s.voices {
		p := &s.voices[i]
		if i < n {
			p.target = freqAt(runs[i])
			p.gain = 1
			if p.amp < 1e-3 {
				p.freq = p.target // fresh voice: don't glide up from a stale pitch
			}
		} else {
			p.gain = 0
		}
	}
}

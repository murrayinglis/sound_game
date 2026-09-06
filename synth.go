package main

import "math"

const (
	sampleRate = 44100
	sweepSec   = 4.0 // time for the playhead to cross the canvas
	maxVoices  = 8

	// How much slope it takes to break free of the scale, in degrees per control
	// tick. Lower and only near-vertical lines bend; higher and nothing ever
	// lands in tune. This is the knob for how "fretted" the instrument feels.
	snapWidth = 0.02

	// Tuning knobs. ctrlSamples is how often the synth re-reads the canvas
	// (~5ms); glide and ampRate are one-pole smoothing coefficients per sample.
	// Raising glide makes pitch track the drawing more literally at the cost of
	// zipper noise; lowering ampRate softens note onsets.
	ctrlSamples = 220
	glide       = 0.0015
	ampRate     = 0.0008

	// Voice mix level, set so a full 8 voices soft-clip rather than blare.
	drive = 0.35

	// A damped feedback delay — cheap reverb, and the biggest single difference
	// between "a tone" and "an instrument in a room". feedback is how long it
	// rings, wet how much you hear it, damp how fast repeats lose their top end.
	delayLen = sampleRate * 28 / 100
	feedback = 0.42
	wet      = 0.35
	damp     = 0.35
)

// hit is one sounding stroke: the row to pitch it from, and which instrument
// drew it.
type hit struct{ y, inst int }

// findRuns returns the centre row of each contiguous vertical stroke in a
// column, one voice per run. Caller holds mu.
// ponytail: voices are assigned to runs top-to-bottom, so crossing strokes swap
// voices mid-glide. Match by nearest previous centre if that ever sounds wrong.
// ponytail: a run of two overlapping colours takes the instrument of its topmost
// pixel; splitting it in two would need a second voice per crossing.
func findRuns(col int, out *[maxVoices]hit) int {
	n, start, inst := 0, -1, 0
	for y := range H {
		k := instAt(col, y)
		switch {
		case k >= 0 && start < 0:
			start, inst = y, k
		case k < 0 && start >= 0:
			out[n] = hit{(start + y - 1) / 2, inst}
			n++
			start = -1
			if n == maxVoices {
				return n
			}
		}
	}
	if start >= 0 {
		out[n] = hit{(start + H - 1) / 2, inst}
		n++
	}
	return n
}

type voice struct {
	phase, freq, amp, target, gain float64
	raw                            float64 // last unsnapped degree, for measuring slope
	inst                           int
}

type synth struct {
	voices [maxVoices]voice
	acc    int
	pos    float64 // write cursor in columns; the renderer uses player.Position()

	delay [delayLen]float64
	dpos  int
	lp    float64 // one-pole state damping the delay's feedback path
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
			// ponytail: phase is in turns, not radians, so it indexes the table
			// directly. No band-limiting — safe for the waveforms shipped here,
			// but a brighter table will alias near the top of the range.
			p.phase += p.freq / sampleRate
			if p.phase >= 1 {
				p.phase--
			}
			sum += wave(p.inst, p.phase) * p.amp
		}

		// tanh instead of a mixer: soft-clips 8 voices without going quiet at 1.
		dry := math.Tanh(sum * drive)

		echo := s.delay[s.dpos]
		s.lp += (dry + echo*feedback - s.lp) * damp
		s.delay[s.dpos] = s.lp
		s.dpos++
		if s.dpos == delayLen {
			s.dpos = 0
		}

		// tanh again on the way out: the delay's feedback can push dry+echo well
		// past full scale, and an int16 conversion that overflows is undefined.
		v := int16(math.Tanh(dry+echo*wet) * 26000)
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
	var runs [maxVoices]hit
	mu.Lock()
	n := findRuns(int(s.pos), &runs)
	mu.Unlock()

	for i := range s.voices {
		p := &s.voices[i]
		if i >= n {
			p.gain = 0
			continue
		}

		raw := degreeAt(float64(runs[i].y))
		fresh := p.amp < 1e-3
		p.inst = runs[i].inst

		// Snap hard while the line is flat, let go as it steepens: a held note
		// lands in tune, a slope bends between notes like a pitch wheel. The
		// existing glide filter then smooths whatever steps are left.
		w := 1.0
		if !fresh {
			w = math.Exp(-math.Abs(raw-p.raw) / snapWidth)
		}
		p.raw = raw
		p.target = freqOfDegree(raw + (math.Round(raw)-raw)*w)

		p.gain = 1
		if fresh {
			p.freq = p.target // don't glide up from a stale pitch
		}
	}
}

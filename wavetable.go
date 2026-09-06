package main

import (
	"embed"
	"encoding/binary"
	"io/fs"
	"log"
	"math"
)

// Single-cycle waveforms from the AKWF collection (public domain, CC0).
// One period of a real instrument, so reading it at a variable rate gives
// continuous pitch for free.

//go:embed wave/*.wav
var waves embed.FS

// Loaded in main once the config names an instrument.
var table []float64

// wave reads the table at a fractional position. The interpolation is not
// optional: nearest-sample lookup on a 600-point table is audibly gritty.
func wave(phase float64) float64 {
	x := phase * float64(len(table))
	i := int(x)
	j := i + 1
	if j == len(table) {
		j = 0
	}
	f := x - float64(i)
	return table[i]*(1-f) + table[j]*f
}

// loadTable walks the RIFF chunks for the sample data, rather than assuming a
// 44-byte header — AKWF files carry extra chunks. Normalising to peak 1.0 means
// drive stays meaningful when you swap in a louder or quieter waveform.
func loadTable(name string) []float64 {
	b, err := waves.ReadFile(name)
	if err != nil {
		available, _ := fs.Glob(waves, "wave/*.wav")
		log.Fatalf("instrument %q not found; available: %v", name, available)
	}
	for i := 12; i+8 <= len(b); {
		size := int(binary.LittleEndian.Uint32(b[i+4 : i+8]))
		if string(b[i:i+4]) == "data" {
			t := make([]float64, size/2)
			peak := 0.0
			for j := range t {
				t[j] = float64(int16(binary.LittleEndian.Uint16(b[i+8+j*2:]))) / 32768
				peak = math.Max(peak, math.Abs(t[j]))
			}
			for j := range t {
				t[j] /= peak
			}
			return t
		}
		i += 8 + size + size%2
	}
	panic("no data chunk in " + name)
}

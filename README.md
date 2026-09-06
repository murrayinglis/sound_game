# sound game

Draw on the canvas; a playhead sweeps left to right and plays what it crosses.
Y is pitch (exponential, 80Hz–2kHz), X is time.

## Run

```bash
go run .
```

Left mouse draws. `C` clears.

## Build for the browser

```bash
GOOS=js GOARCH=wasm go build -o main.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
```

Serve the directory over http and load it with Ebitengine's standard wasm HTML
shell. Audio stays silent until the first click — browsers require a gesture.

## Files

- `main.go` — wiring
- `canvas.go` — the pixel buffer and pencil
- `synth.go` — reads the playhead column, turns strokes into voices
- `wavetable.go` — loads the single-cycle waveform the oscillators read
- `game.go` — input and rendering

Tuning knobs live at the top of `canvas.go` and `synth.go`.

## Sound

Each voice reads a single-cycle waveform at a variable rate, so pitch is
continuous — a drawn curve glissandos rather than stepping between notes.
Change the instrument with `waveFile` in `wavetable.go`:

| file | character |
| --- | --- |
| `AKWF_epiano_0001.wav` | warm, full — the default |
| `AKWF_flute_0004.wav` | soft, close to a sine |
| `AKWF_clarinett_0001.wav` | hollow, odd harmonics only |
| `AKWF_hvoice_0001.wav` | vocal, strong 2nd harmonic, thin at the bottom |

Waveforms are from [AKWF-FREE](https://github.com/KristofferKarlAxelEkstrand/AKWF-FREE)
by Kristoffer Ekstrand, released into the public domain (CC0).

There's no band-limiting, so a waveform with energy above the 8th harmonic will
alias near the top of the canvas. The four above were picked to stay clean.

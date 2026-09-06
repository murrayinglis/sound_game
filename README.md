# sound game

Draw on the canvas; a playhead sweeps left to right and plays what it crosses.
Y is pitch (exponential, 80Hz–2kHz), X is time.

## Run

```bash
go run .
```

Left mouse draws. `C` clears. Click the strip along the bottom, or press `1`–`8`,
to pick an instrument — each draws in its own colour, and strokes keep playing
with the instrument that drew them, so one canvas can hold several at once.

## Config

Copy `config.example.json` to `config.json` (gitignored) and edit:

| key | meaning |
| --- | --- |
| `root` | Hz at the bottom of the canvas |
| `octaves` | how far above the root the top reaches |
| `scale` | semitone offsets from the root, repeated each octave |
| `instruments` | what the picker offers, each a waveform plus a hex colour |

Scales worth trying: `[0,3,5,7,10]` minor pentatonic (the default — nothing
clashes, so any scribble works), `[0,2,4,7,9]` major pentatonic,
`[0,2,3,5,7,8,10]` natural minor, `[0,2,4,5,7,9,11]` major.

Two instruments must not share a colour: the colour of a pixel is how the synth
knows which waveform to play it with. The config is validated at startup, so a
duplicate colour or a range reaching past the Nyquist limit fails with a message
rather than sounding wrong.

There's no config in the browser build — wasm has no filesystem, so it runs on
the defaults compiled into `config.go`.

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

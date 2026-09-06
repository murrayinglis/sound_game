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
- `game.go` — input and rendering

Tuning knobs live at the top of `canvas.go` and `synth.go`.

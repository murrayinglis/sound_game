# sound game

Draw on the canvas; a playhead sweeps left to right and plays what it crosses.
Y is pitch, X is time.

## Run

```bash
go run .
```

- Left mouse draws
- `1`–`8` or the strip along the bottom picks an instrument
- `C` clears

## Config

Copy `config.example.json` to `config.json` and edit `root`, `octaves`,
`scale`, `instruments`. Two instruments can't share a colour; the colour of a
stroke is how the synth knows what to play it with.

## Browser

```bash
GOOS=js GOARCH=wasm go build -o main.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
```

Serve over http. No config file — wasm uses the defaults in `config.go`. Audio
starts on the first click.

---

Waveforms from [AKWF-FREE](https://github.com/KristofferKarlAxelEkstrand/AKWF-FREE), public domain.

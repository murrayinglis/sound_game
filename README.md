# sound game

Draw on the canvas; a playhead sweeps left to right and plays what it crosses.
Y is pitch, X is time.

## Run

```bash
make run
```

Then open http://localhost:8000. The controls live in the page, so the browser
is the way to run it. `make native` still works but has only the keyboard.

- Left mouse draws
- `1`-`8` or the panel picks an instrument
- `G` toggles the block grid
- `-` / `=` change the tempo, applied when the playhead next wraps
- `C` clears

## Config

`config.default.json` is committed and embedded in the binary. It is what the
deployed build runs on, so edit it to change what visitors get.

To change settings locally without touching the deploy, copy it to
`config.json`, which is gitignored and overrides it field by field. wasm has no
filesystem, so the browser only ever sees the embedded copy.

Keys: `bpm`, `steps`, `grid`, `root`, `octaves`, `scale`, `instruments`.
`scale` is a name in `major`, `minor`, `harmonic_minor`, `dorian`,
`major_pentatonic`, `minor_pentatonic`, `blues`, `whole_tone`, `chromatic`.


---

Waveforms from [AKWF-FREE](https://github.com/KristofferKarlAxelEkstrand/AKWF-FREE), public domain.

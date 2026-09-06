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

Copy `config.example.json` to `config.json` and edit `bpm`, `steps`, `root`,
`octaves`, `scale`, `instruments`.

`scale` is a name in `major`, `minor`, `harmonic_minor`, `dorian`,
`major_pentatonic`, `minor_pentatonic`, `blues`, `whole_tone`, `chromatic`.

There is no config file in the browser, since wasm has no filesystem to read
one from. The deployed build runs on the defaults in `config.go`, with the
starting tempo and grid set in `defaults_js.go`.


---

Waveforms from [AKWF-FREE](https://github.com/KristofferKarlAxelEkstrand/AKWF-FREE), public domain.

//go:build js

package main

// The deployed browser build has no config.json to read, so its starting point
// is set here instead. Runs before loadConfig, which on wasm only ever falls
// back to these defaults.
func init() {
	cfg.BPM = 150
	gridDefault = false
}

//go:build !js

package main

// The UI lives in the page, so the native build has only the keyboard.
func bridge() {}

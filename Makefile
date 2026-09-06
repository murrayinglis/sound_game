# $$ so make passes $(go env GOROOT) through to the shell instead of expanding
# it as a make function, which is what leaves the path starting at /lib.
GOROOT := $(shell go env GOROOT)

.PHONY: run build native clean

run: build
	python3 -m http.server

build:
	GOOS=js GOARCH=wasm go build -o main.wasm .
	# -f so an existing read-only copy gets replaced rather than refused
	cp -f "$(GOROOT)/lib/wasm/wasm_exec.js" .

native:
	go run .

clean:
	rm -f main.wasm wasm_exec.js

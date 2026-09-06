# $(shell ...) so the path is resolved by make rather than left for the shell,
# which is what left it starting at /lib when written as $(go env GOROOT).
GOROOT := $(shell go env GOROOT)

.PHONY: run build native clean

run: build
	cd dist && python3 -m http.server

build:
	mkdir -p dist
	GOOS=js GOARCH=wasm go build -o dist/main.wasm .
	# -f so an existing read-only copy gets replaced rather than refused
	cp -f "$(GOROOT)/lib/wasm/wasm_exec.js" dist/
	cp -f media/frutiger.gif media/worm.gif dist/
	# Stamp the wasm's checksum into its URL. Without this a browser will happily
	# pair a fresh index.html with a main.wasm cached from an earlier deploy, and
	# the page then calls bridge functions the old binary never exported.
	sed 's|fetch("main\.wasm")|fetch("main.wasm?v='"$$(cksum dist/main.wasm | cut -d' ' -f1)"'")|' index.html > dist/index.html

native:
	go run .

clean:
	rm -rf dist main.wasm wasm_exec.js

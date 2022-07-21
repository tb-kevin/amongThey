# !/bin/bash

GOOS=js GOARCH=wasm go build -o out/taubyteTestGame.wasm ./examples/isometric

cp ./index.html out/index.html
cp ./wasm_exec.js out/wasm_exec.js


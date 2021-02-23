const go = new Go();
WebAssembly.instantiateStreaming(fetch("./web/assets/cmd/lib.wasm"), go.importObject).then((result) => {
go.run(result.instance);
});

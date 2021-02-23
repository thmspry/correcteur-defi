package main

import (
	"syscall/js"
)

func getDefis(this js.Value, inputs []js.Value) interface{} {
	return 0
}

func main() {
	c := make(chan int)
	js.Global().Set("getDefis", js.FuncOf(getDefis))
	<-c
}

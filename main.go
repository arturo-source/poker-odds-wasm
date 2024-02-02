package main

import (
	"syscall/js"
)

func main() {
	js.Global().Set("getResultsInHTML", js.FuncOf(func(this js.Value, args []js.Value) any {
		handsStr := args[0].String()
		boardStr := args[1].String()
		return getResultsInHTML(handsStr, boardStr)
	}))

	// listen infinite
	<-make(chan bool)
}

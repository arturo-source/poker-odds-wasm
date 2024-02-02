package main

import (
	"syscall/js"
	"time"
)

func main() {
	js.Global().Set("getResultsInHTML", js.FuncOf(func(this js.Value, args []js.Value) any {
		handsStr := args[0].String()
		boardStr := args[1].String()

		hands, board, err := parseUserInputs(handsStr, boardStr)
		if err != nil {
			return getErrorInHTML(err)
		}

		var start = time.Now()
		equities, nCombinations := calculateEquities(hands, board)
		timeElapsed := time.Since(start)

		return getResultsInHTML(hands, board, equities, nCombinations, timeElapsed)
	}))

	// listen infinite
	<-make(chan bool)
}

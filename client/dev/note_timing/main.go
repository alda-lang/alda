package main

import (
	"fmt"
	"os"

	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"alda.io/client/transmitter"
)

func printUsage() {
	fmt.Printf("Usage: %s SCORE_FILE\n", os.Args[0])
}

func main() {
	log.SetGlobalLevel("warn")

	numArgs := len(os.Args[1:])

	if numArgs != 1 {
		printUsage()
		os.Exit(1)
	}

	scoreFilename := os.Args[1]

	ast, err := parser.ParseFile(scoreFilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	scoreUpdates, err := ast.Updates()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	score := model.NewScore()
	if err := score.Update(scoreUpdates...); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	transmitter.NoteTimingTransmitter{}.TransmitScore(score)
}

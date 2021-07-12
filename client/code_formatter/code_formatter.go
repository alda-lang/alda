package code_formatter

import (
	"alda.io/client/parser"
	"fmt"
	"io"
)

func Format(tokens []parser.Token, w io.Writer) {
	lineLength := 0
	for _, token := range tokens {
		lineLength += len(token.Text)
		if lineLength > 60 {
			lineLength = 0
			fmt.Fprint(w, "\n")
		}

		fmt.Fprint(w, token.Text)
	}
}

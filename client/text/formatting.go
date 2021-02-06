package text

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/acarl005/stripansi"
)

// Indent returns a modified version of the supplied string where each line is
// indented the desired amount.
//
// A single indent level is represented as two spaces.
func Indent(amount int, str string) string {
	var buffer bytes.Buffer

	for _, line := range strings.Split(str, "\n") {
		buffer.WriteString(strings.Repeat("  ", amount) + line + "\n")
	}

	return strings.TrimRight(buffer.String(), "\n")
}

func lineLength(line string) int {
	return utf8.RuneCountInString(stripansi.Strip(line))
}

// Boxed returns a multiline string that looks like the provided string in a
// box.
//
// The width is determined by the longest line in the string.
func Boxed(str string) string {
	lines := strings.Split(str, "\n")

	maxLineLength := 0
	for _, line := range lines {
		lineLength := lineLength(line)

		if lineLength > maxLineLength {
			maxLineLength = lineLength
		}
	}

	var buffer bytes.Buffer

	buffer.WriteString("┌" + strings.Repeat("─", maxLineLength+2) + "┐\n")
	buffer.WriteString("│" + strings.Repeat(" ", maxLineLength+2) + "│\n")

	for _, line := range lines {
		lineLength := lineLength(line)

		buffer.WriteString(
			"│ " +
				line +
				strings.Repeat(" ", (maxLineLength-lineLength)) +
				" │\n",
		)
	}

	buffer.WriteString("│" + strings.Repeat(" ", maxLineLength+2) + "│\n")
	buffer.WriteString("└" + strings.Repeat("─", maxLineLength+2) + "┘")

	return buffer.String()
}

package text

import (
	"bytes"
	"strings"
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

package code_generator

import (
	"fmt"
	"strings"
)

func formatFloat(quantity float64) string {
	return strings.TrimRight(
		strings.TrimRight(
			fmt.Sprintf("%f", quantity),
			"0",
		),
		".",
	)
}

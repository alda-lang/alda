package parser

import (
	"fmt"
	"io/ioutil"
)

// Parse a string of input.
func Parse(filepath string, input string) error {
	tokens, err := Scan(filepath, input)

	if err != nil {
		return err
	}

	fmt.Printf("tokens: %+v\n", len(tokens))
	return nil
}

// ParseFile reads a file and parses the input.
func ParseFile(filepath string) error {
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	return Parse(filepath, string(contents))
}

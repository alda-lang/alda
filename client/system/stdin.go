package system

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Returns true if input is being piped into stdin.
//
// Returns an error if something went wrong while trying to determine this.
func IsInputBeingPipedIn() (bool, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}

	return stat.Mode()&os.ModeCharDevice == 0, nil
}

var ErrNoInputSupplied = fmt.Errorf("no input supplied")

// Reads all bytes piped into stdin and returns them.
//
// Returns the error `ErrNoInputSupplied` if no input is being piped in, or a
// different error if something else went wrong.
func ReadStdin() ([]byte, error) {
	isInputSupplied, err := IsInputBeingPipedIn()
	if err != nil {
		return nil, err
	}

	if !isInputSupplied {
		return nil, ErrNoInputSupplied
	}

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

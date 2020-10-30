package errors

import "fmt"

// AldaSourceError is a custom error type that wraps an underlying error to
// provide context about the Alda source file from which the error originated.
type AldaSourceError struct {
	Filename string
	Line     int
	Column   int
	Err      error
}

func (e *AldaSourceError) Error() string {
	return fmt.Sprintf(
		"%s:%d:%d %s",
		e.Filename,
		e.Line,
		e.Column,
		e.Err.Error(),
	)
}

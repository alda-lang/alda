package model

import "fmt"

// AldaSourceContext provides some context about the origin of an error in an
// Alda source file.
type AldaSourceContext struct {
	Filename string
	Line     int
	Column   int
}

// AldaSourceError is a custom error type that wraps an underlying error to
// provide context about the Alda source file from which the error originated.
type AldaSourceError struct {
	Context AldaSourceContext
	Err     error
}

func (e *AldaSourceError) Error() string {
	return fmt.Sprintf(
		"%s:%d:%d %s",
		e.Context.Filename,
		e.Context.Line,
		e.Context.Column,
		e.Err.Error(),
	)
}

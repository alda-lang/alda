package model

import (
	"errors"
	"fmt"
)

// AldaSourceContext provides some context about the origin of an error in an
// Alda source file.
type AldaSourceContext struct {
	Filename string
	Line     int
	Column   int
}

// HasSourceContext is an interface implemented by types that can be linked to a
// particular place (e.g. line and column) in an Alda source file.
type HasSourceContext interface {
	// GetSourceContext returns context (line number, etc.) about the event within
	// the Alda source code.
	GetSourceContext() AldaSourceContext
}

// AldaSourceError is a custom error type that wraps an underlying error to
// provide context about the Alda source file from which the error originated.
type AldaSourceError struct {
	Context AldaSourceContext
	Err     error
}

// Unwrap implements error (un)wrapping as introduced in Go 1.13.
// See: https://blog.golang.org/go1.13-errors
func (ase *AldaSourceError) Unwrap() error {
	return ase.Err
}

// Error returns a string representation of an AldaSourceError.
//
// NOTE: An AldaSourceError can wrap another AldaSourceError, which is
// essentially a stacktrace. I'm punting on representing the entire stacktrace
// here and instead, unwrapping until we reach the bottom-most AldaSourceError
// that points to the immediate problem.
//
// TODO: Consider showing the entire stacktrace to the user. I'm undecided at
// this point whether or not this is a good idea. It seems like it probably is
// (providing more information is better, right?), but some thought needs to go
// into how to format them in a way that's easy to read and useful for the user.
// If we just barf the stacktrace into the user's face in a format that's hard
// to parse, it would actually be worse, not better.
func (ase *AldaSourceError) Error() string {
	var bottom *AldaSourceError
	var err error
	err = ase

	for {
		if !errors.As(err, &bottom) {
			break
		}

		err = bottom.Err
	}

	if bottom.Context.Filename == "" {
		bottom.Context.Filename = "<no file>"
	}

	return fmt.Sprintf(
		"%s:%d:%d %s",
		bottom.Context.Filename,
		bottom.Context.Line,
		bottom.Context.Column,
		bottom.Err.Error(),
	)
}

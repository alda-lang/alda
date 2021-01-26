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

	// Ideally, instead of tracking these separately, we would just use the values
	// from the source context of the bottom-most error.
	//
	// However, there are certain situations where the error is happening deep
	// within the internals of alda-lisp (see: defattribute in lisp.go, for
	// example), and it isn't really feasible (read: it's really hard and I
	// haven't figured out how to do it) to include source context on everything.
	//
	// So, to make this experience a little better, we keep track of the source
	// context at the deepest level where it was there. The chain of errors can
	// keep going deeper and lack context, but in that case, at least we can
	// include the higher-level context, which is typically correct anyway.
	var filename string
	var line int
	var column int

	var err error
	err = ase

	for {
		if bottom != nil && bottom.Context.Line != 0 {
			filename = bottom.Context.Filename
			line = bottom.Context.Line
			column = bottom.Context.Column
		}

		if !errors.As(err, &bottom) {
			break
		}

		err = bottom.Err
	}

	if filename == "" {
		filename = "<no file>"
	}

	return fmt.Sprintf(
		"%s:%d:%d %s",
		filename,
		line,
		column,
		bottom.Err.Error(),
	)
}

package help

import (
	"fmt"
	"os"

	"github.com/logrusorgru/aurora"
)

// UserFacingError is a custom error type that wraps an underlying error to
// provide better information to users about what went wrong and help to point
// them in the right direction.
type UserFacingError struct {
	Err     error
	Message string
}

// Unwrap implements error (un)wrapping as introduced in Go 1.13.
// See: https://blog.golang.org/go1.13-errors
func (ufe *UserFacingError) Unwrap() error {
	return ufe.Err
}

// Error returns a string representation of a UserFacingError.
func (ufe *UserFacingError) Error() string {
	msg := ufe.Message

	if msg == "" {
		msg = ufe.Err.Error()
	}

	return msg
}

// PresentError prints an error message in a way that is helpful for the user.
//
// Inspired by the guidance in https://clig.dev/#errors
func PresentError(err error) {
	switch e := err.(type) {
	case *UserFacingError:
		fmt.Println(e.Error())
	default:
		fmt.Printf(
			`%s
  %s

This might be a bug. For help, consider filing an issue at:
  %s

Or come chat with us on Slack:
  %s`+"\n",
			"Oops! Something went wrong:",
			aurora.BgRed(err.Error()),
			aurora.Underline("https://github.com/alda-lang/alda/issues/new"),
			aurora.Underline("https://slack.alda.io"),
		)
	}
}

// ExitOnError exits gracefully if the provided error is non-nil. Before
// exiting, we print a helpful, user-facing error message.
func ExitOnError(err error) {
	if err != nil {
		PresentError(err)
		os.Exit(1)
	}
}

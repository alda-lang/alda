package help

import (
	"fmt"
	"os"
	"strings"

	"alda.io/client/color"
	"alda.io/client/text"
	"github.com/spf13/cobra"
)

// UserFacingError is a custom error type that wraps an underlying error to
// provide better information to users about what went wrong and help to point
// them in the right direction.
type UserFacingError struct {
	Err     error
	Message string
}

// UserFacingErrorf is a convenient constructor for a UserFacingError that does
// not wrap an existing error. It allows you to concisely create a
// UserFacingError from a string in the same manner as fmt.Errorf.
func UserFacingErrorf(format string, args ...interface{}) error {
	return &UserFacingError{
		Err: fmt.Errorf(format, args...),
	}
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

// UsageError wraps an error to signify that it is caused by incorrect usage of
// the CLI commands/arguments/options.
type UsageError struct {
	Cmd *cobra.Command
	Err error
}

// Unwrap implements error (un)wrapping as introduced in Go 1.13.
// See: https://blog.golang.org/go1.13-errors
func (ue *UsageError) Unwrap() error {
	return ue.Err
}

// Error returns a string representation of a UsageError.
func (ue *UsageError) Error() string {
	return fmt.Sprintf(
		"%s\n\n---\n\nUsage error:\n\n%s\n",
		strings.TrimRight(ue.Cmd.UsageString(), "\n"),
		color.Aurora.Red(text.Indent(1, strings.TrimRight(ue.Err.Error(), "\n"))),
	)
}

// PresentError prints an error message in a way that is helpful for the user.
//
// Inspired by the guidance in https://clig.dev/#errors
func PresentError(err error) {
	switch e := err.(type) {
	case *UserFacingError, *UsageError:
		fmt.Fprintln(os.Stderr, e)
	default:
		fmt.Fprintf(
			os.Stderr,
			`Oops! Something went wrong:
  %s

This might be a bug. For help, consider filing an issue at:
  %s

Or come chat with us on Slack:
  %s`+"\n",
			color.Aurora.BgRed(err),
			color.Aurora.Underline("https://github.com/alda-lang/alda/issues/new/choose"),
			color.Aurora.Underline("http://slack.alda.io"),
		)
	}
}

package help

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
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

// Indent returns a modified version of the supplied string where each line is
// indented the desired amount.
//
// A single indent level is represented as two spaces.
//
// TODO: Move this function to a shared location if we end up needing it
// elsewhere.
func indent(amount int, str string) string {
	var buffer bytes.Buffer

	for _, line := range strings.Split(str, "\n") {
		buffer.WriteString(strings.Repeat("  ", amount) + line + "\n")
	}

	return strings.TrimRight(buffer.String(), "\n")
}

// Error returns a string representation of a UsageError.
func (ue *UsageError) Error() string {
	return fmt.Sprintf(
		"%s\n\n---\n\nUsage error:\n\n%s\n",
		strings.TrimRight(ue.Cmd.UsageString(), "\n"),
		aurora.Red(indent(1, strings.TrimRight(ue.Err.Error(), "\n"))),
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
			aurora.BgRed(err),
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

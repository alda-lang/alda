package color

import (
	"os"

	auroraLib "github.com/logrusorgru/aurora"
	"github.com/mattn/go-isatty"
	"golang.org/x/sys/windows/registry"
)

// We're using a couple of libraries that produce ANSI escape sequences to print
// colored text:
//
// * aurora
// * zerolog
//
// Ideally, both of these libraries would support the standard NO_COLOR
// environment variable, but they don't, so we need to handle it ourselves.
// Fortunately, both libraries include an option to disable colors, so we can
// check for NO_COLOR ourselves and disable color manually if needed.
//
// Another consideration is that not all terminal environments support ANSI
// escape codes (the Windows 7 CMD terminal doesn't, for example). So, we try to
// detect that scenario and disable colors. It would be good to do this in a
// more robust way, something like checking the terminfo/termcap capabilities of
// the terminal to see if it's capable of interpreting ANSI escape sequences,
// but that seems complex and error prone and I'm not really sure that it's
// worth it. So I'm taking a more basic approach here, where we just check to
// see if the terminal is a TTY and assume that if it is a TTY, then it can
// probably interpret ANSI escape sequences.
//
// Reference:
// https://github.com/logrusorgru/aurora/issues/2
// https://eklitzke.org/ansi-color-codes
var EnableColor = isatty.IsTerminal(os.Stdout.Fd()) &&
	len(os.Getenv("NO_COLOR")) == 0

var Aurora auroraLib.Aurora

func init() {
	// Check registry for enabled color printing(only for windows)

	var key, err = registry.OpenKey(registry.CURRENT_USER, "Console", registry.QUERY_VALUE)
	if err != nil {
		EnableColor = false
	} else {
		var val, _, err = key.GetIntegerValue("VirtualTerminalLevel")
		if err != nil {
			EnableColor = false
		} else if val == 0 {
			EnableColor = false
		}
	}

	// HACK: Ideally, aurora would support NO_COLOR, but at least they give us a
	// config option so that we can disable color manually.
	//
	// See the longer comment above EnableColor.

	Aurora = auroraLib.NewAurora(EnableColor)
}

package testing

import (
	"flag"
	// Apparently there was a breaking change in Go 1.13 related to the order that
	// flags are parsed in init functions, or something. When I tried upgrading
	// the Go version we're using in CI from 1.12.9 to 1.15.3, I got the same
	// error message as the one reported in this issue:
	//
	//   https://github.com/golang/go/issues/31859
	//
	// Someone pointed out that explicitly initiating the testing package would
	// fix the issue, so that's what this import is doing.
	_ "testing"

	log "alda.io/client/logging"
)

func init() {
	level := flag.String("log-level", "", "Logging level")
	flag.Parse()
	if *level != "" {
		log.SetGlobalLevel(*level)
	}
}

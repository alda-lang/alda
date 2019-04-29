package testing

import (
	log "alda.io/client/logging"
	"flag"
)

func init() {
	level := flag.String("log-level", "", "Logging level")
	flag.Parse()
	if *level != "" {
		log.SetGlobalLevel(*level)
	}
}

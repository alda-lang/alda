package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var output = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp}

var log = zerolog.New(output).With().Timestamp().Caller().Logger()

// Debug logs at the DEBUG level.
var Debug = log.Debug

// Info logs at the INFO level.
var Info = log.Info

// Warn logs at the WARN level.
var Warn = log.Warn

// Error logs at the ERROR level.
var Error = log.Error

// Fatal logs at the FATAL level.
var Fatal = log.Fatal

// Panic logs at the PANIC level.
var Panic = log.Panic

// SetGlobalLevel sets the global logging level.
func SetGlobalLevel(level string) {
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		panic(fmt.Sprintf("Unrecognized log level: %s", level))
	}
}

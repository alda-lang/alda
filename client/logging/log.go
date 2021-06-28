package logging

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

func logger(writer io.Writer) zerolog.Logger {
	output := zerolog.ConsoleWriter{
		Out:        writer,
		TimeFormat: time.Stamp,
		// HACK: Ideally, zerolog would support NO_COLOR, but at least they give us
		// a config option so that we can disable color manually.
		NoColor: len(os.Getenv("NO_COLOR")) > 0,
	}
	return zerolog.New(output).With().Timestamp().Caller().Logger()
}

var log = logger(os.Stderr)

// SetOutput sets the writer that we log to.
//
// ...OK, technically, we create a NEW logger that is logging to the new writer,
// because as far as I can tell, zerolog won't let you change the writer of a
// zerolog.Logger instance after the instance is created.
func SetOutput(writer io.Writer) {
	log = logger(writer)
}

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

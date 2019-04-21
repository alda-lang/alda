package logging

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

var output = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Stamp}

// Log is a global logger.
var log = zerolog.New(output).With().Caller().Logger()

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

package transmitter

import (
	"fmt"

	log "alda.io/client/logging"
	"alda.io/client/model"
)

// TransmissionContext provides context about how the score data should be
// transmitted.
type TransmissionContext struct {
	// A time marking (e.g. 0:30) or marker from which to start.
	from string
	// A time marking (e.g. 1:00) or marker at which to end.
	to string
	// The index of the first event to transmit. (default: 0)
	fromIndex int
	// The index (+ 1) of the last event to transmit. (default: len(events))
	toIndex int
	// An optional map of parts to offsets that is used for the purpose of
	// synchronization. For each part present in the map, any event transmitted
	// for that part will have the indicated offset subtracted from its offset.
	// The use case for this is REPL usage, where a score is built up
	// incrementally as the score is being played.
	syncOffsets map[string]float64
	// When true, no further transmissions are expected for this particular score.
	//
	// What this means can vary depending on the transmitter. For the OSC
	// transmitter, it means that we append a "shutdown" message at the end of the
	// OSC bundle, which tells the player process to shut down after playing the
	// score.
	oneOff bool
	// When true, the score will only be loaded, as opposed to being played,
	// displayed, performed, etc.
	loadOnly bool
}

// TransmissionOption is a function that customizes a TransmissionContext
// instance.
type TransmissionOption func(*TransmissionContext)

// TransmitFrom sets the time marking or marker from which to start.
func TransmitFrom(from string) TransmissionOption {
	return func(ctx *TransmissionContext) {
		log.Debug().
			Str("from", from).
			Msg("Applying transmission option")

		ctx.from = from
	}
}

// TransmitTo sets the time marking or marker at which to end.
func TransmitTo(to string) TransmissionOption {
	log.Debug().
		Str("to", to).
		Msg("Applying transmission option")

	return func(ctx *TransmissionContext) { ctx.to = to }
}

// TransmitFromIndex sets the index of the first event to transmit.
func TransmitFromIndex(i int) TransmissionOption {
	return func(ctx *TransmissionContext) {
		log.Debug().
			Interface("fromIndex", i).
			Msg("Applying transmission option")

		ctx.fromIndex = i
	}
}

// TransmitToIndex sets the index (+ 1) of the last event to transmit.
func TransmitToIndex(i int) TransmissionOption {
	log.Debug().
		Interface("toIndex", i).
		Msg("Applying transmission option")

	return func(ctx *TransmissionContext) { ctx.toIndex = i }
}

// SyncOffsets uses the provided map of parts to offsets for the purpose of
// synchronization. For each part present in the map, any event transmitted for
// that part will have the indicated offset subtracted from its offset. The use
// case for this is REPL usage, where a score is built up incrementally as the
// score is being played.
func SyncOffsets(syncOffsets map[string]float64) TransmissionOption {
	return func(ctx *TransmissionContext) {
		log.Debug().
			Str("syncOffsets", fmt.Sprintf("%#v", syncOffsets)).
			Msg("Applying transmission option")

		ctx.syncOffsets = syncOffsets
	}
}

// OneOff specifies that no further transmissions are expected for this
// particular score.
//
// What this means can vary depending on the transmitter. For the OSC
// transmitter, it means that we append a "shutdown" message at the end of the
// OSC bundle, which tells the player process to shut down after playing the
// score.
func OneOff() TransmissionOption {
	return func(ctx *TransmissionContext) {
		log.Debug().
			Bool("oneOff", true).
			Msg("Applying transmission option")

		ctx.oneOff = true
	}
}

// LoadOnly specifies that the score is to be loaded only, as opposed to being
// played, displayed, performed, etc.
func LoadOnly() TransmissionOption {
	return func(ctx *TransmissionContext) {
		log.Debug().
			Bool("loadOnly", true).
			Msg("Applying transmission option")

		ctx.loadOnly = true
	}
}

// A Transmitter sends score data somewhere for performance, visualization,
// etc.
type Transmitter interface {
	// TransmitScore sends score data somewhere.
	TransmitScore(score *model.Score, opts ...TransmissionOption) error
}

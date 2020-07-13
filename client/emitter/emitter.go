package emitter

import "alda.io/client/model"

// EmissionContext provides context about how the score data should be emitted.
type EmissionContext struct {
	// A time marking (e.g. 0:30) or marker from which to start.
	from string
	// A time marking (e.g. 1:00) or marker at which to end.
	to string
	// The index of the first event to emit. (default: 0)
	fromIndex int
	// The index (+ 1) of the last event to emit. (default: len(events))
	toIndex int
	// When true, no further emissions are expected for this particular score.
	//
	// What this means can vary depending on the emitter. For the OSC emitter, it
	// means that we append a "shutdown" message at the end of the OSC bundle,
	// which tells the player process to shut down after playing the score.
	oneOff bool
}

// EmissionOption is a function that customizes an EmissionContext
// instance.
type EmissionOption func(*EmissionContext)

// EmitFrom sets the time marking or marker from which to start.
func EmitFrom(from string) EmissionOption {
	return func(ctx *EmissionContext) { ctx.from = from }
}

// EmitTo sets the time marking or marker at which to end.
func EmitTo(to string) EmissionOption {
	return func(ctx *EmissionContext) { ctx.to = to }
}

// EmitFromIndex sets the index of the first event to emit.
func EmitFromIndex(i int) EmissionOption {
	return func(ctx *EmissionContext) { ctx.fromIndex = i }
}

// EmitToIndex sets the index (+ 1) of the last event to emit.
func EmitToIndex(i int) EmissionOption {
	return func(ctx *EmissionContext) { ctx.toIndex = i }
}

// OneOff specifies that no further emissions are expected for this particular
// score.
//
// What this means can vary depending on the emitter. For the OSC emitter, it
// means that we append a "shutdown" message at the end of the OSC bundle, which
// tells the player process to shut down after playing the score.
func OneOff() EmissionOption {
	return func(ctx *EmissionContext) { ctx.oneOff = true }
}

// An Emitter sends score data somewhere for performance, visualization, etc.
type Emitter interface {
	// EmitScore sends score data somewhere.
	EmitScore(score *model.Score, opts ...EmissionOption) error
}

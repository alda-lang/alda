package emitter

import "alda.io/client/model"

// EmissionContext provides context about how the score data should be emitted.
type EmissionContext struct {
	// A time marking (e.g. 0:30) or marker from which to start.
	from string
	// A time marking (e.g. 1:00) or marker at which to end.
	to string
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

// An Emitter sends score data somewhere for performance, visualization, etc.
type Emitter interface {
	// EmitScore sends score data somewhere.
	EmitScore(score *model.Score, opts ...EmissionOption) error
}

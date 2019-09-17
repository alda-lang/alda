package emitter

import "alda.io/client/model"

// An Emitter sends score data somewhere for performance, visualization, etc.
type Emitter interface {
	// EmitScore sends score data somewhere.
	EmitScore(score *model.Score) error
}

package model

// The ScoreUpdate interface is implemented by events that update a score.
type ScoreUpdate interface {
	// UpdateScore modifies a score and returns nil, or returns an error if
	// something went wrong.
	UpdateScore(score *Score) error
	// DurationMs returns a number of milliseconds representing how long an event
	// takes. For events where duration is not relevant (e.g. an octave change
	// event), this can return 0.
	//
	// The context for this is the Cram event, which involves summing the duration
	// of a number of events and then time-scaling them into a fixed duration.
	DurationMs(part *Part) float32

	// VariableValue returns the value that is captured when an event is part of a
	// variable definition.
	//
	// Returns an error if the value cannot be captured.
	VariableValue(score *Score) (ScoreUpdate, error)
}

// The ScoreEvent interface is implemented by events that occur at moments of
// time in a score.
type ScoreEvent interface{}

// A Score is a data structure representing a musical score.
//
// Scores are built up via events (structs which implement ScoreUpdate) that
// update aspects of the score data.
//
// chordMode: When true, notes/rests added to the score are placed at the same
// offset. Otherwise, they are appended sequentially.
type Score struct {
	Parts            []*Part
	CurrentParts     []*Part
	Aliases          map[string][]*Part
	Events           []ScoreEvent
	GlobalAttributes *GlobalAttributes
	Markers          map[string]OffsetMs
	Variables        map[string][]ScoreUpdate
	chordMode        bool
}

// NewScore returns an initialized score.
func NewScore() *Score {
	return &Score{
		Parts:            []*Part{},
		Aliases:          map[string][]*Part{},
		GlobalAttributes: NewGlobalAttributes(),
		Markers:          map[string]OffsetMs{},
		Variables:        map[string][]ScoreUpdate{},
	}
}

// Update applies a variable number of ScoreUpdates to a Score, short-circuiting
// and returning the first error that occurs.
//
// Returns nil if no error occurs.
func (score *Score) Update(updates ...ScoreUpdate) error {
	for _, update := range updates {
		if err := update.UpdateScore(score); err != nil {
			return err
		}
	}

	return nil
}

// Tracks returns a map of Part instances to track numbers for the purposes of
// emitting score data.
func (score *Score) Tracks() map[*Part]int32 {
	tracks := map[*Part]int32{}

	for i, part := range score.Parts {
		tracks[part.origin] = int32(i + 1)
	}

	return tracks
}

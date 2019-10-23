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
	chordMode        bool
}

// NewScore returns an initialized score.
func NewScore() *Score {
	return &Score{
		Parts:            []*Part{},
		Aliases:          map[string][]*Part{},
		GlobalAttributes: NewGlobalAttributes(),
	}
}

// NamedParts returns a list of Parts included in the score that correspond to
// the provided `alias`, or nil if there are no such parts.
func (score *Score) NamedParts(alias string) []*Part {
	return score.Aliases[alias]
}

// UnnamedParts returns the list of Parts in the score that are not included in
// any alias, and that are instances of the stock instrument identified by
// `name`.
func (score *Score) UnnamedParts(name string) []*Part {
	stock := "N/A"
	if stockInstrument, err := stockInstrumentName(name); err == nil {
		stock = stockInstrument
	}

	results := []*Part{}

	for _, part := range score.Parts {
		isNamedPart := false
		for _, namedParts := range score.Aliases {
			for _, namedPart := range namedParts {
				if namedPart == part {
					isNamedPart = true
				}
			}
		}

		if !isNamedPart && part.StockInstrument.Name() == stock {
			results = append(results, part)
		}
	}

	return results
}

// AliasedStockInstruments returns the list of Parts in the score that have a
// dedicated alias (e.g. 'piano "foo"'), and that are instances of the stock
// instrument identified by `name`.
func (score *Score) AliasedStockInstruments(name string) []*Part {
	stock := "N/A"
	if stockInstrument, err := stockInstrumentName(name); err == nil {
		stock = stockInstrument
	}

	results := []*Part{}

	for _, namedParts := range score.Aliases {
		if len(namedParts) == 1 {
			part := namedParts[0]
			if part.StockInstrument.Name() == stock {
				results = append(results, part)
			}
		}
	}

	return results
}

// AliasesFor returns the list of aliases in the score that correspond to a
// single part, the one provided.
func (score *Score) AliasesFor(part *Part) []string {
	results := []string{}

	for alias, parts := range score.Aliases {
		if len(parts) == 1 && parts[0] == part {
			results = append(results, alias)
		}
	}

	return results
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

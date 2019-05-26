package model

// A Score is a data structure representing a musical score.
//
// Scores are built up via events (structs which implement ScoreUpdate) that
// update aspects of the score data.
type Score struct{}

// Update applies a variable number of ScoreUpdates to a Score, short-circuiting
// and returning the first error that occurs.
//
// Returns nil if no error occurs.
func (score *Score) Update(updates ...ScoreUpdate) error {
	for _, update := range updates {
		if err := update.updateScore(score); err != nil {
			return err
		}
	}

	return nil
}

// The ScoreUpdate interface defines how something updates a score.
type ScoreUpdate interface {
	updateScore(score *Score) error
}

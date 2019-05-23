package model

// A Score is a data structure representing a musical score.
//
// Scores are built up via events (structs which implement ScoreUpdate) that
// update aspects of the score data.
type Score struct{}

// The ScoreUpdate interface defines how something updates a score.
type ScoreUpdate interface {
	updateScore(score *Score) error
}

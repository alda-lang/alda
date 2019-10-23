package model

import (
	"errors"
)

// A Marker gives a name to a point in time in a score.
type Marker struct {
	Name string
}

// UpdateScore implements ScoreUpdate.UpdateScore by storing the current point
// in time as a named marker.
//
// The assumption is that all current parts have the same current offset. In the
// rare event that this is not the case, an error is thrown because we don't
// have a single current offset to store under the marker.
func (Marker) UpdateScore(score *Score) error {
	return errors.New("Marker.UpdateScore not implemented")
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since defining a
// marker is conceptually instantaneous.
func (Marker) DurationMs(part *Part) float32 {
	return 0
}

// AtMarker is an action where a part's offset gets set to a point in time
// denoted previously by a Marker.
type AtMarker struct {
	Name string
}

// UpdateScore implements ScoreUpdate.UpdateScore by setting the current offset
// of all active parts to the offset stored in the marker with the provided
// name.
//
// If no such marker was previously defined, an error is thrown.
func (AtMarker) UpdateScore(score *Score) error {
	return errors.New("AtMarker.UpdateScore not implemented")
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0 because jumping
// to a marker is conceptually instantaneous.
//
// NB: I don't think it really makes any sense for an AtMarker event to occur
// within a Cram expression. Arguably, this should result in a syntax error.
func (AtMarker) DurationMs(part *Part) float32 {
	return 0
}

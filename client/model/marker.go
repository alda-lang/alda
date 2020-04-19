package model

import (
	"fmt"
)

// A Marker gives a name to a point in time in a score.
type Marker struct {
	Name string
}

// UpdateScore implements ScoreUpdate.UpdateScore by storing the current point
// in time as a named marker.
//
// The assumption is that all current parts have the same current offset. In the
// rare event that this is not the case, an error is returned because we don't
// have a single current offset to store under the marker.
func (marker Marker) UpdateScore(score *Score) error {
	if len(score.CurrentParts) == 0 {
		return fmt.Errorf(
			"can't define marker \"%s\" outside the context of a part", marker.Name,
		)
	}

	offset := score.CurrentParts[0].CurrentOffset

	if len(score.CurrentParts) > 1 {
		for _, part := range score.CurrentParts[1:len(score.CurrentParts)] {
			if part.CurrentOffset != offset {
				return fmt.Errorf("offset of marker \"%s\" unclear", marker.Name)
			}
		}
	}

	score.Markers[marker.Name] = offset

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since defining a
// marker is conceptually instantaneous.
func (Marker) DurationMs(part *Part) float64 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (marker Marker) VariableValue(score *Score) (ScoreUpdate, error) {
	return marker, nil
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
// If no such marker was previously defined, an error is returned.
func (atMarker AtMarker) UpdateScore(score *Score) error {
	offset, hit := score.Markers[atMarker.Name]
	if !hit {
		return fmt.Errorf("Marker undefined: %s", atMarker.Name)
	}

	for _, part := range score.CurrentParts {
		part.LastOffset = part.CurrentOffset
		part.CurrentOffset = offset
	}

	score.ApplyGlobalAttributes()

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0 because jumping
// to a marker is conceptually instantaneous.
//
// NB: I don't think it really makes any sense for an AtMarker event to occur
// within a Cram expression. Arguably, this should result in a syntax error.
func (AtMarker) DurationMs(part *Part) float64 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (atMarker AtMarker) VariableValue(score *Score) (ScoreUpdate, error) {
	return atMarker, nil
}

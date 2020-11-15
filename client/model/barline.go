package model

import (
	"alda.io/client/json"
)

// A Barline has no audible effect on a score. Its purpose is to visually
// separate elements in an Alda source file.
type Barline struct {
	SourceContext AldaSourceContext
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (barline Barline) GetSourceContext() AldaSourceContext {
	return barline.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (Barline) JSON() *json.Container {
	return json.Object("type", "barline")
}

// Beats implements DurationComponent.Beats by returning 0.
//
// A barline is considered a DurationComponent purely for syntactic reasons, for
// better or for worse.
func (Barline) Beats() float64 {
	return 0
}

// Ms implements DurationComponent.Ms by returning 0.
//
// A barline is considered a DurationComponent purely for syntactic reasons, for
// better or for worse.
func (Barline) Ms(tempo float64) float64 {
	return 0
}

// UpdateScore implements ScoreUpdate.UpdateScore by doing nothing. The purpose
// of a barline is to visually separate elements in an Alda source file.
func (Barline) UpdateScore(score *Score) error {
	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since a barline
// is conceptually instantaneous.
func (Barline) DurationMs(part *Part) float64 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (barline Barline) VariableValue(score *Score) (ScoreUpdate, error) {
	return barline, nil
}

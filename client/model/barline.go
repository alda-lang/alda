package model

// A Barline has no audible effect on a score. Its purpose is to visually
// separate elements in an Alda source file.
type Barline struct{}

// Beats implements DurationComponent.Beats by returning 0.
//
// A barline is considered a DurationComponent purely for syntactic reasons, for
// better or for worse.
func (Barline) Beats() (float64, error) {
	return 0, nil
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

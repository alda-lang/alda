package model

// A Barline has no audible effect on a score. Its purpose is to visually
// separate elements in an Alda source file.
type Barline struct{}

// Beats implements DurationComponent.Beats by returning 0.
//
// A barline is considered a DurationComponent purely for syntactic reasons, for
// better or for worse.
func (Barline) Beats() (float32, error) {
	return 0, nil
}

// Beats implement DurationComponent.Ms by returning 0.
//
// A barline is considered a DurationComponent purely for syntactic reasons, for
// better or for worse.
func (Barline) Ms(tempo float32) float32 {
	return 0
}

func (Barline) updateScore(score *Score) error {
	return nil
}

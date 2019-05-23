package model

// A Barline has no audible effect on a score. Its purpose is to visually
// separate elements in an Alda source file.
type Barline struct{}

func (Barline) updateScore(score *Score) error {
	return nil
}

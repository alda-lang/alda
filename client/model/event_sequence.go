package model

type EventSequence struct {
	Events []ScoreUpdate
}

func (es EventSequence) updateScore(score *Score) error {
	return score.Update(es.Events...)
}

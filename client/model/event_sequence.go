package model

type EventSequence struct {
	Events []ScoreUpdate
}

func (es EventSequence) updateScore(score *Score) error {
	for _, event := range es.Events {
		if err := event.updateScore(score); err != nil {
			return err
		}
	}

	return nil
}

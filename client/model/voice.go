package model

import (
	"errors"
)

// A VoiceGroupEndMarker denotes the end of a voice group, i.e. the point at
// which we are just dealing with a single voice.
type VoiceGroupEndMarker struct{}

// UpdateScore implements ScoreUpdate.UpdateScore by ... FIXME
func (VoiceGroupEndMarker) UpdateScore(score *Score) error {
	return errors.New("VoiceGroupEndMarker.UpdateScore not implemented")
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since a voice
// marker is conceptually instantaneous.
func (VoiceGroupEndMarker) DurationMs(part *Part) float32 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (vgem VoiceGroupEndMarker) VariableValue(
	score *Score,
) (ScoreUpdate, error) {
	return vgem, nil
}

// A VoiceMarker indicates that the following events belong to one voice in a
// group of voices.
type VoiceMarker struct {
	VoiceNumber int32
}

// UpdateScore implements ScoreUpdate.UpdateScore by ... FIXME
func (VoiceMarker) UpdateScore(score *Score) error {
	return errors.New("VoiceMarker.UpdateScore not implemented")
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since a voice
// marker is conceptually instantaneous.
func (VoiceMarker) DurationMs(part *Part) float32 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (vm VoiceMarker) VariableValue(score *Score) (ScoreUpdate, error) {
	return vm, nil
}

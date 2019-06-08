package model

import (
	"errors"
)

type VoiceGroupEndMarker struct{}

func (VoiceGroupEndMarker) updateScore(score *Score) error {
	return errors.New("VoiceGroupEndMarker.updateScore not implemented")
}

type VoiceMarker struct {
	VoiceNumber int32
}

func (VoiceMarker) updateScore(score *Score) error {
	return errors.New("VoiceMarker.updateScore not implemented")
}

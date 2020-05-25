package model

import (
	"fmt"

	log "alda.io/client/logging"
)

// Voices wraps a map of voice numbers to Part instances and an insertion order,
// so that we can not only look up a Part by voice number, but also know which
// voice was the last to be added.
type Voices struct {
	voices         map[int32]*Part
	insertionOrder []int32
}

// NewVoices returns an initialized Voices structure.
func NewVoices() *Voices {
	return &Voices{
		voices:         map[int32]*Part{},
		insertionOrder: []int32{},
	}
}

// AddVoice adds a voice to a Voices structure.
func (v *Voices) AddVoice(voiceNumber int32, voice *Part) {
	v.voices[voiceNumber] = voice
	v.insertionOrder = append(v.insertionOrder, voiceNumber)
}

// NewVoice creates and returns a new voice.
func (part *Part) NewVoice(voiceNumber int32) *Part {
	if len(part.voices.voices) == 0 {
		if part.voiceTemplate != nil {
			panic(fmt.Sprintf(
				"Part has no voices, but part.voiceTemplate is %#v",
				part.voiceTemplate,
			))
		}

		// part.voiceTemplate serves as a template for the state of the part as of
		// the start of each voice.
		voiceTemplate := part.Clone()
		voiceTemplate.voiceTemplate = voiceTemplate
		part.voiceTemplate = voiceTemplate
	}

	voice := part.voiceTemplate.Clone()

	part.voices.AddVoice(voiceNumber, voice)

	return voice
}

// GetVoice returns an existing voice or creates a new one.
func (part *Part) GetVoice(voiceNumber int32) *Part {
	if existingVoice, hit := part.voices.voices[voiceNumber]; hit {
		return existingVoice
	}

	return part.NewVoice(voiceNumber)
}

// A VoiceMarker indicates that the following events belong to one voice in a
// group of voices.
type VoiceMarker struct {
	VoiceNumber int32
}

// UpdateScore implements ScoreUpdate.UpdateScore by initializing a voice for
// each current part. This initiates a "voice group" if it wasn't already done
// previously.
//
// A voice group effectively forks a part into N copies of itself, one per
// voice.
func (vm VoiceMarker) UpdateScore(score *Score) error {
	log.Debug().
		Int32("VoiceNumber", vm.VoiceNumber).
		Msg("Voice marker")

	for _, part := range score.CurrentParts {
		voice := part.GetVoice(vm.VoiceNumber)

		for i, currentPart := range score.CurrentParts {
			if currentPart.origin == part.origin {
				score.CurrentParts[i] = voice
			}
		}
	}

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since a voice
// marker is conceptually instantaneous.
func (VoiceMarker) DurationMs(part *Part) float64 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (vm VoiceMarker) VariableValue(score *Score) (ScoreUpdate, error) {
	return vm, nil
}

// A VoiceGroupEndMarker denotes the end of a voice group, i.e. the point at
// which we are just dealing with a single voice.
type VoiceGroupEndMarker struct{}

// UpdateScore implements ScoreUpdate.UpdateScore by updating the "origin" part
// of each current part (i.e. the part that was forked into N voices) to be
// equal to the last voice to finish (offset-wise), and then updating the
// pointers to the part (e.g. in score.CurrentParts and score.Parts) to point to
// the last voice to finish, effectively making it "the" voice of that part in
// the score, going forward.
func (VoiceGroupEndMarker) UpdateScore(score *Score) error {
	for i, part := range score.CurrentParts {
		if len(part.voices.voices) == 0 {
			continue
		}

		insertionOrder := part.voices.insertionOrder
		lastInsertedVoiceNumber := insertionOrder[len(insertionOrder)-1]

		lastVoiceToFinish := part.GetVoice(lastInsertedVoiceNumber)

		if len(part.voices.voices) > 1 {
			for _, voiceNumber := range insertionOrder[0 : len(insertionOrder)-1] {
				voice := part.GetVoice(voiceNumber)

				if voice.CurrentOffset > lastVoiceToFinish.CurrentOffset {
					lastVoiceToFinish = voice
				}
			}
		}

		lastVoiceToFinish.voices = NewVoices()
		lastVoiceToFinish.voiceTemplate = nil

		score.CurrentParts[i] = lastVoiceToFinish

		for i, partsPart := range score.Parts {
			if partsPart.origin == part.origin {
				score.Parts[i] = lastVoiceToFinish
			}
		}

		for _, parts := range score.Aliases {
			for i, aliasPart := range parts {
				if aliasPart.origin == part.origin {
					parts[i] = lastVoiceToFinish
				}
			}
		}
	}

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since a voice
// marker is conceptually instantaneous.
func (VoiceGroupEndMarker) DurationMs(part *Part) float64 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (vgem VoiceGroupEndMarker) VariableValue(
	score *Score,
) (ScoreUpdate, error) {
	return vgem, nil
}

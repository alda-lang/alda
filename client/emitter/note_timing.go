package emitter

import (
	"fmt"
	"math"

	"alda.io/client/model"
)

// NoteTimingEmitter prints the offset, duration, and pitch of every note in a
// score.
//
// The goal is to compare the same output from Alda v1 to find discrepancies.
//
// See: https://github.com/daveyarwood/alda-v1-v2-comparer
type NoteTimingEmitter struct{}

// EmitScore implements Emitter.EmitScore by printing the offset, duration, and
// pitch of every note in the score.
func (nte NoteTimingEmitter) EmitScore(score *model.Score) error {
	tracks := score.Tracks()

	fmt.Println("track, offset, duration, midi note")

	for _, event := range score.Events {
		switch event.(type) {
		case model.NoteEvent:
			noteEvent := event.(model.NoteEvent)

			track := tracks[noteEvent.Part]
			offset := int32(math.Round(float64(noteEvent.Offset)))
			duration := int32(math.Round(float64(noteEvent.AudibleDuration)))

			fmt.Printf("%d, %d, %d, %d\n", track, offset, duration, noteEvent.MidiNote)
		default:
			return fmt.Errorf("unsupported event: %#v", event)
		}
	}

	return nil
}

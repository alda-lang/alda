package transmitter

import (
	"fmt"
	"math"

	"alda.io/client/model"
)

// NoteTimingTransmitter prints the offset, duration, and pitch of every note in
// a score.
//
// The goal is to compare the same output from Alda v1 to find discrepancies.
//
// See: https://github.com/daveyarwood/alda-v1-v2-comparer
type NoteTimingTransmitter struct{}

// TransmitScore implements Transmitter.TransmitScore by printing the offset,
// duration, and pitch of every note in the score.
func (nte NoteTimingTransmitter) TransmitScore(score *model.Score) error {
	fmt.Println("offset,duration,midi note")

	for _, event := range score.Events {
		switch event := event.(type) {
		case model.NoteEvent:
			offset := int32(math.Round(event.Offset))
			duration := int32(math.Round(event.AudibleDuration))

			fmt.Printf("%d,%d,%d\n", offset, duration, event.MidiNote)
		default:
			return fmt.Errorf("unsupported event: %#v", event)
		}
	}

	return nil
}

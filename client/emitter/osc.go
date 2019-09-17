package emitter

import (
	"fmt"
	"math"
	"time"

	"alda.io/client/model"
	"github.com/hypebeast/go-osc/osc"
)

// OSCEmitter sends OSC messages to a player process.
type OSCEmitter struct {
	Port int
}

func systemPlayMsg() *osc.Message {
	return osc.NewMessage("/system/play")
}

func midiPatchMsg(track int32, offset int32, patch int32) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/patch", track))
	msg.Append(offset)
	msg.Append(patch)
	return msg
}

func midiNoteMsg(
	track int32, offset int32, note int32, duration int32, audibleDuration int32,
	velocity int32,
) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/note", track))
	msg.Append(offset)
	msg.Append(note)
	msg.Append(duration)
	msg.Append(audibleDuration)
	msg.Append(velocity)
	return msg
}

// EmitScore implements Emitter.EmitScore by sending OSC messages to instruct a
// player process how to perform the score.
func (oe OSCEmitter) EmitScore(score *model.Score) error {
	client := osc.NewClient("localhost", int(oe.Port))
	bundle := osc.NewBundle(time.Now())

	tracks := map[*model.Part]int32{}
	trackNumber := int32(1)

	for _, part := range score.Parts {
		tracks[part] = trackNumber
		patchNumber := part.StockInstrument.(model.MidiInstrument).PatchNumber
		bundle.Append(midiPatchMsg(trackNumber, 0, patchNumber))
		trackNumber++
	}

	for _, event := range score.Events {
		switch event.(type) {
		case model.NoteEvent:
			noteEvent := event.(model.NoteEvent)
			bundle.Append(midiNoteMsg(
				tracks[noteEvent.Part],
				int32(math.Round(float64(noteEvent.Offset))),
				noteEvent.MidiNote,
				int32(math.Round(float64(noteEvent.Duration))),
				int32(math.Round(float64(noteEvent.AudibleDuration))),
				127, // FIXME: include velocity on note events and use it here
			))
		default:
			return fmt.Errorf("unsupported event: %#v", event)
		}
	}

	bundle.Append(systemPlayMsg())

	client.Send(bundle)

	return nil
}

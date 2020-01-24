package emitter

import (
	"fmt"
	"math"
	"time"

	"alda.io/client/model"
	"github.com/daveyarwood/go-osc/osc"
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

func midiPercussionMsg(track int32, offset int32) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/percussion", track))
	msg.Append(offset)
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

func midiVolumeMsg(track int32, offset int32, volume int32) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/volume", track))
	msg.Append(offset)
	msg.Append(volume)
	return msg
}

func midiPanningMsg(track int32, offset int32, panning int32) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/panning", track))
	msg.Append(offset)
	msg.Append(panning)
	return msg
}

// EmitScore implements Emitter.EmitScore by sending OSC messages to instruct a
// player process how to perform the score.
func (oe OSCEmitter) EmitScore(score *model.Score) error {
	client := osc.NewClient("localhost", int(oe.Port))
	client.SetNetworkProtocol(osc.TCP)
	bundle := osc.NewBundle(time.Now())

	tracks := score.Tracks()

	for part, trackNumber := range tracks {
		// We currently only have MIDI instruments. This might change in the future,
		// which is why Instrument is an interface instead of a plain struct. For
		// now, we're operating under the assumption that all instruments are MIDI
		// instruments.
		stockInstrument := part.StockInstrument.(model.MidiInstrument)

		patchNumber := stockInstrument.PatchNumber
		bundle.Append(midiPatchMsg(trackNumber, 0, patchNumber))

		if stockInstrument.IsPercussion {
			bundle.Append(midiPercussionMsg(trackNumber, 0))
		}
	}

	for _, event := range score.Events {
		switch event.(type) {
		case model.NoteEvent:
			noteEvent := event.(model.NoteEvent)
			track := tracks[noteEvent.Part]
			offset := int32(math.Round(float64(noteEvent.Offset)))

			// Scheduling volume and panning control change messages before every note
			// is terribly inefficient, but it's what we've always done in alda v1 and
			// it gets the job done.
			//
			// At some point, it would be nice to make it so that we only send a
			// volume or panning control change message when the values change.
			bundle.Append(
				midiVolumeMsg(
					track,
					offset,
					int32(math.Round(float64(noteEvent.TrackVolume*127))),
				),
			)

			bundle.Append(
				midiPanningMsg(
					track,
					offset,
					int32(math.Round(float64(noteEvent.Panning*127))),
				),
			)

			bundle.Append(midiNoteMsg(
				track,
				offset,
				noteEvent.MidiNote,
				int32(math.Round(float64(noteEvent.Duration))),
				int32(math.Round(float64(noteEvent.AudibleDuration))),
				int32(math.Round(float64(noteEvent.Volume*127))),
			))
		default:
			return fmt.Errorf("unsupported event: %#v", event)
		}
	}

	bundle.Append(systemPlayMsg())

	client.Send(bundle)

	return nil
}

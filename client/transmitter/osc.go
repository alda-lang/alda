package transmitter

import (
	"fmt"
	"math"
	"sort"
	"time"

	log "alda.io/client/logging"
	"alda.io/client/model"
	"github.com/daveyarwood/go-osc/osc"
)

// OSCTransmitter sends OSC messages to a player process.
type OSCTransmitter struct {
	Port int
}

func pingMsg() *osc.Message {
	return osc.NewMessage("/ping")
}

func systemMidiExportMsg(filename string) *osc.Message {
	msg := osc.NewMessage("/system/midi/export")
	msg.Append(filename)
	return msg
}

func systemPlayMsg() *osc.Message {
	return osc.NewMessage("/system/play")
}

func systemStopMsg() *osc.Message {
	return osc.NewMessage("/system/stop")
}

func systemShutdownMsg(offset int32) *osc.Message {
	msg := osc.NewMessage("/system/shutdown")
	msg.Append(offset)
	return msg
}

func systemOffsetMsg(offset int32) *osc.Message {
	msg := osc.NewMessage("/system/offset")
	msg.Append(offset)
	return msg
}

func systemTempoMsg(offset int32, tempo float32) *osc.Message {
	msg := osc.NewMessage("/system/tempo")
	msg.Append(offset)
	msg.Append(tempo)
	return msg
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

func oscClient(port int) *osc.Client {
	return osc.NewClient("localhost", int(port), osc.ClientProtocol(osc.TCP))
}

// TransmitMidiExportMessage sends a "MIDI export" message to a player process.
func (oe OSCTransmitter) TransmitMidiExportMessage(filename string) error {
	return oscClient(oe.Port).Send(systemMidiExportMsg(filename))
}

// TransmitPingMessage sends a "ping" message to a player process.
func (oe OSCTransmitter) TransmitPingMessage() error {
	return oscClient(oe.Port).Send(pingMsg())
}

// TransmitPlayMessage sends a "play" message to a player process.
func (oe OSCTransmitter) TransmitPlayMessage() error {
	return oscClient(oe.Port).Send(systemPlayMsg())
}

// TransmitStopMessage sends a "stop" message to a player process.
func (oe OSCTransmitter) TransmitStopMessage() error {
	return oscClient(oe.Port).Send(systemStopMsg())
}

// TransmitShutdownMessage sends a "shutdown" message to a player process.
func (oe OSCTransmitter) TransmitShutdownMessage(offset int32) error {
	return oscClient(oe.Port).Send(systemShutdownMsg(offset))
}

// TransmitOffsetMessage sends an "offset" message to a player process.
func (oe OSCTransmitter) TransmitOffsetMessage(offset int32) error {
	return oscClient(oe.Port).Send(systemOffsetMsg(offset))
}

func tempoMessages(
	score *model.Score, startOffset float64, endOffset float64,
) []*osc.Message {
	tempoItinerary := score.TempoItinerary()

	tempoOffsets := []float64{}
	for offset := range tempoItinerary {
		tempoOffsets = append(tempoOffsets, offset)
	}
	sort.Float64s(tempoOffsets)

	// In the case where we're starting a ways into the score (i.e. if the
	// `--from` option is supplied), we want to skip any extraneous tempo changes
	// that happened before that point in the score. Except we do want the last
	// one before or at the start offset, so that the initial tempo is correct.
	firstTempoOffset := 0.0
	for _, tempoOffset := range tempoOffsets {
		if tempoOffset > startOffset {
			break
		}

		// Keep going until we reach a tempo offset past the start offset. At that
		// point, we'll use the previous tempo offset recorded here, because that
		// would be the last tempo change before the start offset.
		firstTempoOffset = tempoOffset
	}

	messages := []*osc.Message{}

	// Now, we want to emit tempo messages for each tempo change within the time
	// range of the excerpt of the score that we're playing.
	for _, tempoOffset := range tempoOffsets {
		// Filter out any tempo changes prior to the `--from` time marking / marker,
		// when supplied. (...except for the last tempo offset prior to that point
		// in time; see the comment where we defined `firstTempoOffset` above.)
		if tempoOffset < firstTempoOffset {
			continue
		}

		// Filter out any tempo changes after the `--to` time marking / marker, when
		// supplied.
		if tempoOffset >= endOffset {
			break
		}

		tempo := tempoItinerary[tempoOffset]

		// We subtract `startOffset` from the offset because we're about to do the
		// same thing to the offset of every note event, for reasons that are
		// explained below.
		//
		// By default, `startOffset` is 0, so the usual scenario is that the offset
		// is not adjusted.
		offset := tempoOffset - startOffset

		// If the effective offset is earlier than the notional start offset (0),
		// then we'll place the tempo change right at the beginning (0).
		if offset < 0 {
			offset = 0
		}

		// The OSC API works with int offsets and float tempos, so we do the
		// necessary conversions here.
		offsetRounded := int32(math.Round(offset))
		tempo32 := float32(tempo)
		messages = append(messages, systemTempoMsg(offsetRounded, tempo32))
	}

	return messages
}

// ScoreToOSCBundle returns the OSC bundle that should be sent to an Alda player
// process in order to transmit the provided score.
func (oe OSCTransmitter) ScoreToOSCBundle(
	score *model.Score, opts ...TransmissionOption,
) (*osc.Bundle, error) {
	ctx := &TransmissionContext{toIndex: -1}
	for _, opt := range opts {
		opt(ctx)
	}

	if ctx.toIndex == -1 {
		ctx.toIndex = len(score.Events)
	}

	log.Debug().
		Str("ctx", fmt.Sprintf("%#v", ctx)).
		Msg("Transmission options applied.")

	events := score.Events[ctx.fromIndex:ctx.toIndex]

	startOffset := 0.0
	endOffset := math.MaxFloat64

	if ctx.from != "" {
		offset, err := score.InterpretOffsetReference(ctx.from)
		if err != nil {
			return nil, err
		}

		startOffset = offset
	}

	if ctx.to != "" {
		offset, err := score.InterpretOffsetReference(ctx.to)
		if err != nil {
			return nil, err
		}

		endOffset = offset
	}

	bundle := osc.NewBundle(time.Now())

	// In order to support features like:
	//
	// * Avoiding scheduling more volume and panning control change messages than
	//   we have to (see below).
	//
	// * Playing just a slice of a score, e.g. `alda play --from 0:05 --to 0:10`
	//
	// ...we sort the events in the score by offset and schedule them in
	// chronological order.
	sort.Slice(events, func(i, j int) bool {
		return events[i].EventOffset() < events[j].EventOffset()
	})

	// In Alda's model, the track volume and panning are an attribute of each
	// individual note. However, in MIDI, these attributes are set persistently on
	// a channel via a control change message.
	//
	// To make this work, as we're scheduling the events of the score in
	// chonological order, we keep track of the volume and panning attributes for
	// each track, so that we can send volume and panning control changes only
	// when necessary (when the values change).
	currentVolume := map[int32]float64{}
	currentPanning := map[int32]float64{}

	tracks := score.Tracks()

	for part, trackNumber := range tracks {
		currentVolume[trackNumber] = -1
		currentPanning[trackNumber] = -1

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

	// Append tempo messages to the score, based on the tempo changes in the
	// score. (See *Score.TempoItinerary.)
	//
	// We avoid doing this if there are any sync offsets, which is the case if the
	// score is being emitted as an incremental update in an Alda REPL session. It
	// would be complicated (maybe even impossible?) to get tempo messages right
	// in this context, so we punt on it.
	//
	// NOTE: We _do_ include tempo messages in every other context, including
	// playing an entire score from the REPL via the :play command, or exporting a
	// score via the :export command. It is important especially for the MIDI
	// export use case that we include tempo messages in the MIDI sequence, so
	// that the MIDI file can include context about the tempo when it's imported
	// into other tools.
	if len(ctx.syncOffsets) == 0 {
		for _, tempoMsg := range tempoMessages(score, startOffset, endOffset) {
			bundle.Append(tempoMsg)
		}
	}

	// We keep track of the known (audible) length of the score as we iterate
	// through the events. That way, at the end, if we want to schedule a shutdown
	// message to clean up, we can schedule it for shortly after the audible end
	// of the score.
	scoreLength := 0.0

	for index, event := range events {
		eventOffset := event.EventOffset()

		// Filter out events before the `--from` time marking / marker, when
		// supplied.
		if eventOffset < startOffset {
			continue
		}

		// Filter out events after the `--to` time marking / marker, when supplied.
		if eventOffset >= endOffset {
			break
		}

		switch event := event.(type) {
		case model.NoteEvent:

			// Add offset for any rests done after the last Note
			// Additional rests at the end of a code sequence create a timing mismatch
			// between Part.CurrentOffset and this note event's Offset + duration
			// Part.CurrentOffset should equal Note Offset + Note Duration, so this statement
			// Increments Note Duration so that they become equal
			restOffset := int32(0)
			if index == len(events)-1 {
				lastEvent := event
				// Last event offset + duration must equal score.duration
				currNoteEnd := int32(math.Round(lastEvent.Duration)) + int32(lastEvent.EventOffset())
				partOffset := int32(lastEvent.Part.CurrentOffset)
				restOffset = partOffset - currNoteEnd
			}

			track := tracks[event.Part]

			// We subtract `startOffset` from the offset so that when the `--from`
			// option is used (e.g. `--from 0:30`), we will shift all of the events
			// back by that amount so that playback starts as if those events were at
			// the beginning of the score. Otherwise, `--from 0:30` would result in
			// you having to wait 30 seconds before you hear anything.
			//
			// By default, `startOffset` is 0, so the usual scenario is that the event
			// offsets are not adjusted.
			offset := event.Offset - startOffset

			// When sync offsets are provided, we subtract the specified offset for
			// each part from its events. (When syncOffsets isn't provided, or when
			// the sync offset for a part isn't specified, the result is that 0 is
			// subtracted from the part's events' offsets, i.e. the offsets for that
			// part are not adjusted.)
			//
			// NB: ctx.from and ctx.syncOffsets are not intended to be used together.
			// If they are used together, the behavior is unspecified (we would
			// probably subtract too much from each offset and the features wouldn't
			// work the way they're supposed to.)
			offset -= ctx.syncOffsets[event.Part]

			// The OSC API works with offsets that are ints, not floats, so we do the
			// rounding here and work with the int value from here onward.
			offsetRounded := int32(math.Round(offset))

			if event.TrackVolume != currentVolume[track] {
				currentVolume[track] = event.TrackVolume

				bundle.Append(
					midiVolumeMsg(
						track,
						offsetRounded,
						int32(math.Round(event.TrackVolume*127)),
					),
				)
			}

			if event.Panning != currentPanning[track] {
				currentPanning[track] = event.Panning

				bundle.Append(
					midiPanningMsg(
						track,
						offsetRounded,
						int32(math.Round(event.Panning*127)),
					),
				)
			}

			bundle.Append(midiNoteMsg(
				track,
				offsetRounded,
				event.MidiNote,
				int32(math.Round(event.Duration)+float64(restOffset)),
				int32(math.Round(event.AudibleDuration)),
				int32(math.Round(event.Volume*127)),
			))

			scoreLength = math.Max(scoreLength, offset+event.AudibleDuration)
		default:
			return nil, fmt.Errorf("unsupported event: %#v", event)
		}
	}

	if !ctx.loadOnly {
		bundle.Append(systemPlayMsg())
	}

	if ctx.oneOff {
		bundle.Append(systemShutdownMsg(int32(math.Round(scoreLength + 10000))))
	}

	return bundle, nil
}

// TransmitScore implements Transmitter.TransmitScore by sending OSC messages to
// instruct a player process how to perform the score.
func (oe OSCTransmitter) TransmitScore(
	score *model.Score, opts ...TransmissionOption,
) error {
	bundle, err := oe.ScoreToOSCBundle(score, opts...)
	if err != nil {
		return err
	}

	log.Debug().
		Interface("bundle", bundle).
		Msg("Sending OSC bundle.")

	return oscClient(oe.Port).Send(bundle)
}

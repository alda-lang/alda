package model

import (
	"alda.io/client/help"
)

// midiChannelUsage keeps track of which parts are using each of the 16
// available MIDI channels.
type midiChannelUsage = [16][]*Part

func (s *Score) partHasExclusiveAccess(part *Part, midiChannel int32) bool {
	channelUsage := s.midiChannelUsage[midiChannel]

	return len(channelUsage) == 1 && channelUsage[0] == part.origin
}

func (s *Score) simpleCheck(part *Part) (bool, int32) {
	for channel, parts := range s.midiChannelUsage {
		// Channel 9 is reserved for percussion, but this function is for finding
		// an available MIDI channel for a non-percussion instrument.
		if channel == 9 {
			continue
		}

		if len(parts) == 0 || s.partHasExclusiveAccess(part, int32(channel)) {
			return true, int32(channel)
		}
	}

	return false, -1
}

func overlaps(
	note1start float64, note1duration float64, note2start float64,
	note2duration float64,
) bool {
	note1end := note1start + note1duration
	note2end := note2start + note2duration

	return note2start < note1end && note2end > note1start
}

func (s *Score) complexCheck(part *Part, noteDurationMs float64) (bool, int32) {
	unavailableChannels := map[int32]bool{}

	// Go through every note in the score and make a note of which channels have
	// notes belonging to other parts that overlap with the proposed note,
	// recording them in `unavailableChannels`.
	for _, event := range s.Events {
		note, isNote := event.(NoteEvent)

		if !isNote {
			continue
		}

		if note.Part.origin == part.origin {
			continue
		}

		if _, unavailable := unavailableChannels[note.MidiChannel]; unavailable {
			continue
		}

		if overlaps(
			note.Offset, note.AudibleDuration, part.CurrentOffset, noteDurationMs,
		) {
			unavailableChannels[note.MidiChannel] = true
		}
	}

	// Prefer the channel already assigned, if it's still available.
	if part.MidiChannel != -1 {
		if _, unavailable := unavailableChannels[part.MidiChannel]; !unavailable {
			return true, part.MidiChannel
		}
	}

	for channel := range s.midiChannelUsage {
		// Channel 9 is reserved for percussion, but this function is for finding
		// an available MIDI channel for a non-percussion instrument.
		if channel == 9 {
			continue
		}

		if _, unavailable := unavailableChannels[int32(channel)]; !unavailable {
			return true, int32(channel)
		}
	}

	return false, -1
}

func (s *Score) availableMidiChannel(
	part *Part, noteDurationMs float64,
) (int32, error) {
	if available, channel := s.simpleCheck(part); available {
		return channel, nil
	}

	if available, channel := s.complexCheck(part, noteDurationMs); available {
		return channel, nil
	}

	return -1, help.UserFacingErrorf(
		`No MIDI channel available for part "%s" at offset %f.

This means that your score has more than 16 instruments, and we tried to map
the instruments' notes to the 16 MIDI channels by having multiple instruments
share a channel as needed, but it wasn't logistically possible.`,
		part.Name,
		part.CurrentOffset,
	)
}

func (s *Score) determineMidiChannel(
	part *Part, noteDurationMs float64,
) (int32, error) {
	// Channel 9 is the only channel that can be used for percussion.
	if part.StockInstrument.(MidiInstrument).IsPercussion {
		return 9, nil
	}

	if part.MidiChannel != -1 &&
		s.partHasExclusiveAccess(part, part.MidiChannel) {
		return part.MidiChannel, nil
	}

	return s.availableMidiChannel(part, noteDurationMs)
}

func (s *Score) assignMidiChannel(
	part *Part, noteDurationMs float64,
) (int32, error) {
	channel, err := s.determineMidiChannel(part, noteDurationMs)
	if err != nil {
		return -1, err
	}

	alreadyRecorded := false
	for _, recordedPart := range s.midiChannelUsage[channel] {
		if recordedPart == part.origin {
			alreadyRecorded = true
		}
	}

	if !alreadyRecorded {
		s.midiChannelUsage[channel] = append(
			s.midiChannelUsage[channel], part.origin,
		)
	}

	return channel, nil
}

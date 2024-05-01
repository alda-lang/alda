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

func (s *Score) availableMidiChannel(
	part *Part, duration float64,
) (int32, error) {
	for channel, parts := range s.midiChannelUsage {
		// Channel 9 is reserved for percussion, but this function is for finding
		// an available MIDI channel for a non-percussion instrument.
		if channel == 9 {
			continue
		}

		if len(parts) == 0 || (len(parts) == 1 && parts[0] == part.origin) {
			return int32(channel), nil
		}
	}

	// TODO: Implement complex check
	return -1, help.UserFacingErrorf(
		`No MIDI channel available at offset %f.

This means that your score has more than 16 instruments, and we tried to map
the instruments' notes to the 16 MIDI channels by having multiple instruments
share a channel as needed, but it wasn't logistically possible.`,
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

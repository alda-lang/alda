package model

import (
	"fmt"

	_ "alda.io/client/testing"
)

func expectNoteOffsets(expectedOffsets ...OffsetMs) func(*Score) error {
	return func(s *Score) error {
		if len(s.Events) != len(expectedOffsets) {
			return fmt.Errorf(
				"expected %d events, got %d",
				len(expectedOffsets),
				len(s.Events),
			)
		}

		for i := 0; i < len(expectedOffsets); i++ {
			expectedOffset := expectedOffsets[i]
			actualOffset := s.Events[i].(NoteEvent).Offset
			if expectedOffset != actualOffset {
				return fmt.Errorf(
					"expected note #%d to have offset %f, but it was %f",
					i+1,
					expectedOffset,
					actualOffset,
				)
			}
		}

		return nil
	}
}

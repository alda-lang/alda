package interop

import "alda.io/client/model"

// Translates MusicXML duration with divisions into Alda duration
func TranslateDuration(divisions int, duration int) model.Duration {
	return model.Duration{
		Components: []model.DurationComponent{
			model.NoteLength{
				Denominator: 4 * float64(divisions) / float64(duration),
			},
		},
	}
}

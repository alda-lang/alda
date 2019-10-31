package model

import (
	"fmt"
	"math"
)

// An OffsetMs describes a point in time a specific number of milliseconds into
// a score.
//
// NB: This is a float64 instead of a float32 so that we can use `sort.Float64s`
// to maintain the GlobalAttributes data structure.
type OffsetMs = float64

// DurationComponent instances are added together and the sum is the total
// duration of the Duration.
type DurationComponent interface {
	// Beats returns the number of beats represented by a DurationComponent.
	//
	// An error is returned if the component's duration cannot be expressed in
	// beats.
	Beats() (float32, error)
	// Ms returns the duration of a DurationComponent in milliseconds.
	Ms(tempo float32) float32
}

// A NoteLength represents a standard note length in Western classical music
// notation.
type NoteLength struct {
	// Denominator is analogous to the "American names" for note values, e.g. 4 is
	// a quarter note (1/4), 2 is a half note (1/2), etc.
	//
	// Mathematically, it is the denominator in the fraction "1/N," which
	// describes ratio of the note length to a whole note (1).
	//
	// https://en.wikipedia.org/wiki/Note_value
	//
	// Note that Alda allows composers to use non-standard note lengths like a
	// "sixth note" (1/6), i.e. a note with a length such that it takes 6 of them
	// to fill a bar of 4/4. In standard musical notation, this note length would
	// be represented as a quarter note triplet, and you would typically only see
	// 3 or 6 of them grouped together, whereas in Alda it's just a note length
	// with a denominator of 6.
	Denominator float32
	// Dots is the number of dots to be added to a note length to modify its
	// value. The first dot added to a note length adds half of the original
	// value, e.g. a dotted half note ("2.") adds 1 beat to the value of a half
	// note (2 beats), for a total of 3 beats. Each subsequent dot adds a
	// progressively halved value, e.g. a dotted half note ("2..") adds an
	// additional 1/2 beat, for a total of 3-1/2 beats.
	//
	// https://en.wikipedia.org/wiki/Dotted_note
	Dots int32
}

// Beats implements DurationComponent.Beats by calculating the number of beats
// represented by a standard note length.
func (nl NoteLength) Beats() (float32, error) {
	return float32(
		float64(4/nl.Denominator) * (2 - math.Pow(2, float64(-nl.Dots))),
	), nil
}

// Ms implements DurationComponent.Ms by calculating the duration in
// milliseconds within the context of a tempo.
func (nl NoteLength) Ms(tempo float32) float32 {
	beats := float32(
		float64(4/nl.Denominator) * (2 - math.Pow(2.0, float64(-nl.Dots))),
	)

	return NoteLengthBeats{Quantity: beats}.Ms(tempo)
}

// NoteLengthBeats expresses a duration as a specific number of beats.
type NoteLengthBeats struct {
	Quantity float32
}

// Beats implements DurationComponent.Beats by describing a specific number of
// beats.
func (nl NoteLengthBeats) Beats() (float32, error) {
	return nl.Quantity, nil
}

// Ms implements DurationComponent.Ms by calculating the duration in
// milliseconds within the context of a tempo.
func (nl NoteLengthBeats) Ms(tempo float32) float32 {
	return nl.Quantity * (60000 / tempo)
}

// NoteLengthMs expresses a duration as a specific number of milliseconds.
type NoteLengthMs struct {
	Quantity float32
}

// Beats implements DurationComponent.Beats by returning an error.
//
// NB: It is mathematically possible to convert a number of milliseconds into a
// number of beats, given a tempo. TODO: Consider whether Beats() should take a
// tempo argument and NoteLengthMs.Beats(tempo float32) should be implemented.
// At this point, I'm not sure if that functionality is needed.
func (nl NoteLengthMs) Beats() (float32, error) {
	return 0, fmt.Errorf("A millisecond note length cannot be expressed in beats")
}

// Ms implements DurationComponent.Ms by describing a specific number of
// milliseconds.
func (nl NoteLengthMs) Ms(tempo float32) float32 {
	return nl.Quantity
}

// Duration describes the length of time occupied by a note or other event.
type Duration struct {
	Components []DurationComponent
}

// Beats implements DurationComponent.Beats by summing the beats of the
// components.
func (d Duration) Beats() (float32, error) {
	beats := float32(0)

	for _, component := range d.Components {
		componentBeats, err := component.Beats()

		if err != nil {
			return 0, err
		}

		beats += componentBeats
	}

	return beats, nil
}

// Ms implements DurationComponent.Ms by summing the millisecond durations of
// the components.
func (d Duration) Ms(tempo float32) float32 {
	ms := float32(0)

	for _, component := range d.Components {
		ms += component.Ms(tempo)
	}

	return ms
}

// TempoRole describes the relationship a part has with the global tempo of a
// score.
type TempoRole int

const (
	// TempoRoleUnspecified means that the part has no special relationship with
	// the global tempo.
	TempoRoleUnspecified TempoRole = 0
	// TempoRoleMaster means that any tempo changes in the part apply to the score
	// as a whole.
	TempoRoleMaster TempoRole = 1
)

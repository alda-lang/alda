package model

import (
	"fmt"
)

// NoteLetter represents a note letter in Western standard musical notation.
type NoteLetter int

const (
	// A is the note "A" in Western standard musical notation.
	A NoteLetter = iota
	// B is the note "B" in Western standard musical notation.
	B
	// C is the note "C" in Western standard musical notation.
	C
	// D is the note "D" in Western standard musical notation.
	D
	// E is the note "E" in Western standard musical notation.
	E
	// F is the note "F" in Western standard musical notation.
	F
	// G is the note "G" in Western standard musical notation.
	G
)

// NewNoteLetter returns the NoteLetter that corresponds to the provided
// character. e.g. 'a' => A
//
// Returns an error if there is no corresponding NoteLetter.
func NewNoteLetter(letter rune) (NoteLetter, error) {
	switch letter {
	case 'a':
		return A, nil
	case 'b':
		return B, nil
	case 'c':
		return C, nil
	case 'd':
		return D, nil
	case 'e':
		return E, nil
	case 'f':
		return F, nil
	case 'g':
		return G, nil
	default:
		return -1, fmt.Errorf("Invalid note letter: %c", letter)
	}
}

// An Accidental is an accidental (flat, sharp, or natural) from Western
// standard musical notation.
type Accidental int

const (
	// Flat is the "flat" accidental.
	Flat Accidental = iota
	// Natural is the "natural" accidental.
	Natural
	// Sharp is the "sharp" accidental.
	Sharp
)

// NewAccidental returns the Accidental that corresponds to the provided string.
// e.g. "flat" => Flat
//
// Returns an error if there is no corresponding Accidental.
func NewAccidental(accidental string) (Accidental, error) {
	switch accidental {
	case "flat":
		return Flat, nil
	case "natural":
		return Natural, nil
	case "sharp":
		return Sharp, nil
	default:
		return -1, fmt.Errorf("Invalid accidental: %s", accidental)
	}
}

// The PitchIdentifier interface defines how a pitch is specified and
// determined.
//
// This is a multi-step process. The first step is syntax, which provides only
// partial information, e.g. a note letter like C. Then we gain additional
// information (e.g. octave, key signature) as we build up the score from the
// AST, and the methods of this interface are used to determine the precise
// pitch of the note.
type PitchIdentifier interface {
	// CalculateMidiNote returns the MIDI note number of a note, given contextual
	// information about the part playing the note (e.g. octave, key signature,
	// transposition).
	CalculateMidiNote(
		octave int32, keySignature KeySignature, transposition int32,
	) int32
}

// LetterAndAccidentals specifies a pitch as a note letter and (optional)
// accidentals.
type LetterAndAccidentals struct {
	NoteLetter  NoteLetter
	Accidentals []Accidental
}

// CalculateMidiNote implements PitchIdentifier.CalculateMidiNote by placing the
// note in the given octave, applying the key signature, and applying the
// transposition.
func (laa LetterAndAccidentals) CalculateMidiNote(
	octave int32, keySignature KeySignature, transposition int32,
) int32 {
	intervals := map[NoteLetter]int32{
		C: 0, D: 2, E: 4, F: 5, G: 7, A: 9, B: 11,
	}

	baseMidiNoteNumber := ((octave + 1) * 12) + intervals[laa.NoteLetter]

	var accidentals []Accidental
	if laa.Accidentals == nil {
		accidentals = keySignature[laa.NoteLetter]
	} else {
		accidentals = laa.Accidentals
	}

	for _, accidental := range accidentals {
		switch accidental {
		case Flat:
			baseMidiNoteNumber--
		case Sharp:
			baseMidiNoteNumber++
		}
	}

	return baseMidiNoteNumber + transposition
}

// MidiNoteNumber specifies a pitch as a MIDI note number.
type MidiNoteNumber struct {
	MidiNote int32
}

// CalculateMidiNote implements PitchIdentifier.CalculateMidiNote by returning
// an explicit MIDI note number and applying the transposition.
func (mnn MidiNoteNumber) CalculateMidiNote(
	octave int32, keySignature KeySignature, transposition int32,
) int32 {
	return mnn.MidiNote + transposition
}

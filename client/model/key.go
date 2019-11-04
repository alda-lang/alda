package model

import (
	"math"
)

// ScaleType represents a type of scale or mode, e.g. major, minor.
type ScaleType int

const (
	// Ionian represents a major/Ionian scale.
	Ionian ScaleType = iota

	// Dorian represents a Dorian scale.
	Dorian

	// Phrygian represents a Phrygian scale.
	Phrygian

	// Lydian represents a Lydian scale.
	Lydian

	// Mixolydian represents a Mixolydian scale.
	Mixolydian

	// Aeolian represents a minor/Aeolian scale.
	Aeolian

	// Locrian represents a Locrian scale.
	Locrian
)

// KeySignature is a key signature in Western standard musical notation.
//
// Alda models a key signature as a map from NoteLetter to []Accidental. This
// allows us to represent the "standard" key signatures, e.g.:
//
//   (A major)
//   {F: [sharp], C: [sharp], G: [sharp]}
//
// as well as unconventional ones like:
//
//   {B: [flat, flat], G: [sharp], E: [flat]}
type KeySignature map[NoteLetter][]Accidental

func shiftKey(k KeySignature, this Accidental, that Accidental) KeySignature {
	result := KeySignature{}

	for _, letter := range []NoteLetter{A, B, C, D, E, F, G} {
		accidentals, inKeySig := k[letter]

		if !inKeySig {
			result[letter] = []Accidental{this}
			continue
		}

		if len(accidentals) == 1 && accidentals[0] == that {
			// shifting the other way (towards `this`) cancels `that` out, so this
			// letter is not part of the new key signature
			//
			// e.g. if you sharpen the key of Bb (two flats, B and E), you get the key
			// of B, which has 5 sharps (F, C, G, D, A), and B and E are not included
			// in the key signature.
			continue
		}

		result[letter] = []Accidental{}

		containsThat := false
		for _, accidental := range accidentals {
			if accidental == this || containsThat {
				result[letter] = append(result[letter], accidental)
			} else {
				// Don't copy over this accidental; this has the effect of shifting one
				// semitone toward `this`.
				//
				// Set `containsThat` to `true` so that we know to copy the remaining
				// `that` accidentals.
				containsThat = true
			}
		}

		// If we didn't adjust the accidentals just now (i.e. it's all `this`)
		// accidentals, then the way to shift the key further toward `this` is to
		// add an additional `this` accidental.
		if !containsThat {
			result[letter] = append(result[letter], this)
		}
	}

	return result
}

// Flatten returns the key signature one semitone lower.
func (k KeySignature) Flatten() KeySignature {
	return shiftKey(k, Flat, Sharp)
}

// Sharpen returns the key signature one semitone higher.
func (k KeySignature) Sharpen() KeySignature {
	return shiftKey(k, Sharp, Flat)
}

func orderOfFlats() []NoteLetter {
	return []NoteLetter{B, E, A, D, G, C, F}
}

func orderOfSharps() []NoteLetter {
	return []NoteLetter{F, C, G, D, A, E, B}
}

// Given a scale type, returns a map of note letter to an integer representing a
// number of sharps or flats. A negative integer represents a number of flats,
// and a positive integer represents a number of sharps.
func partialCircleOfFifths(scaleType ScaleType) map[NoteLetter]int {
	var start int
	switch scaleType {
	case Ionian:
		start = -1
	case Dorian:
		start = -3
	case Phrygian:
		start = -5
	case Lydian:
		start = 0
	case Mixolydian:
		start = -2
	case Aeolian:
		start = -4
	case Locrian:
		start = -6
	}

	result := map[NoteLetter]int{}

	for i, letter := range orderOfSharps() {
		result[letter] = start + i
	}

	return result
}

// KeySignatureFromScale returns a key signature given a tonic note and a scale
// type.
func KeySignatureFromScale(
	tonic LetterAndAccidentals, scaleType ScaleType,
) KeySignature {
	n := partialCircleOfFifths(scaleType)[tonic.NoteLetter]

	var order []NoteLetter
	var accidental Accidental

	if n > 0 {
		order = orderOfSharps()
		accidental = Sharp
	} else {
		order = orderOfFlats()
		accidental = Flat
	}

	letters := order[0:int(math.Abs(float64(n)))]

	keySignature := KeySignature{}

	for _, letter := range letters {
		keySignature[letter] = []Accidental{accidental}
	}

	for _, accidental := range tonic.Accidentals {
		switch accidental {
		case Flat:
			keySignature = keySignature.Flatten()
		case Sharp:
			keySignature = keySignature.Sharpen()
		}
	}

	return keySignature
}

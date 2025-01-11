package model

import (
	"math"
	"strings"

	"alda.io/client/json"
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
//	(A major)
//	{F: [sharp], C: [sharp], G: [sharp]}
//
// as well as unconventional ones like:
//
//	{B: [flat, flat], G: [sharp], E: [flat]}
type KeySignature map[NoteLetter][]Accidental

func (ks KeySignature) String() string {
	laas := []string{}

	for note, accidentals := range ks {
		laa := strings.Builder{}
		laa.WriteRune(rune(note + 'a'))

		for _, acc := range accidentals {
			switch acc {
			case Flat:
				laa.WriteString("-")
			case Natural:
				laa.WriteString("_")
			case Sharp:
				laa.WriteString("+")
			}
		}

		laas = append(laas, laa.String())
	}

	return strings.Join(laas, " ")
}

// JSON implements RepresentableAsJSON.JSON.
func (ks KeySignature) JSON() *json.Container {
	object := json.Object()

	for letter, accidentals := range ks {
		accidentalsArray := json.Array()
		for _, accidental := range accidentals {
			accidentalsArray.ArrayAppend(accidental.String())
		}

		object.Set(accidentalsArray, letter.String())
	}

	return object
}

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

// partialCircleOfFifths returns a map of note letter to placement in the circle
// of fifths given a scaleType. A negative integer represents a number of flats,
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

// KeySignatureFromCircleOfFifths returns a key signature given a placement in
// the circle of fifths. A negative integer represents a number of flats,
// and a positive integer represents a number of sharps.
func KeySignatureFromCircleOfFifths(fifths int) KeySignature {
	var order []NoteLetter
	var accidental Accidental

	if fifths == 0 {
		return KeySignature{}
	} else if fifths > 0 {
		order = orderOfSharps()
		accidental = Sharp
	} else {
		order = orderOfFlats()
		accidental = Flat
	}

	keySignature := KeySignature{}

	for i := 0; i < int(math.Abs(float64(fifths))); i++ {
		keySignature[order[i%len(order)]] = append(
			keySignature[order[i%len(order)]],
			accidental,
		)
	}

	return keySignature
}

// KeySignatureFromScale returns a key signature given a tonic note and a scale
// type.
func KeySignatureFromScale(
	tonic LetterAndAccidentals, scaleType ScaleType,
) KeySignature {
	fifths := partialCircleOfFifths(scaleType)[tonic.NoteLetter]
	keySignature := KeySignatureFromCircleOfFifths(fifths)

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

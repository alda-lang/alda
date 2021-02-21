package model

import (
	"github.com/go-test/deep"
	"testing"
)

type keySignatureTestCase struct {
	label    string
	expected KeySignature
	actual   KeySignature
}

func executeKeySignatureTestCase(
	t *testing.T, testCases ...keySignatureTestCase,
) {
	for _, testCase := range testCases {
		if diff := deep.Equal(testCase.expected, testCase.actual); diff != nil {
			t.Error(testCase.label)
			for _, diffItem := range diff {
				t.Errorf("%v", diffItem)
			}
		}
	}
}

func TestKeySignatureFromCircleOfFifthsSharps(t *testing.T) {
	key := KeySignature{}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "C major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(0),
	})

	key[F] = []Accidental{Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "G major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(1),
	})

	key[C] = []Accidental{Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "D major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(2),
	})

	key[G] = []Accidental{Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "A major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(3),
	})

	key[D] = []Accidental{Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "E major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(4),
	})

	key[A] = []Accidental{Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "B major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(5),
	})

	key[E] = []Accidental{Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "F sharp major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(6),
	})

	key[B] = []Accidental{Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "C sharp major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(7),
	})

	key[F] = []Accidental{Sharp, Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "G sharp major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(8),
	})

	key[C] = []Accidental{Sharp, Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "D sharp major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(9),
	})

	key[G] = []Accidental{Sharp, Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "A sharp major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(10),
	})

	key[D] = []Accidental{Sharp, Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "E sharp major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(11),
	})

	key[A] = []Accidental{Sharp, Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "B sharp major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(12),
	})

	key[E] = []Accidental{Sharp, Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "F double sharp major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(13),
	})

	key[B] = []Accidental{Sharp, Sharp}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "C double sharp major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(14),
	})
}

func TestKeySignatureFromCircleOfFifthsFlats(t *testing.T) {
	key := KeySignature{}

	key[B] = []Accidental{Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "F major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-1),
	})

	key[E] = []Accidental{Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "B flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-2),
	})

	key[A] = []Accidental{Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "E flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-3),
	})

	key[D] = []Accidental{Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "A flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-4),
	})

	key[G] = []Accidental{Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "D flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-5),
	})

	key[C] = []Accidental{Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "G flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-6),
	})

	key[F] = []Accidental{Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "C flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-7),
	})

	key[B] = []Accidental{Flat, Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "F flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-8),
	})

	key[E] = []Accidental{Flat, Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "B double flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-9),
	})

	key[A] = []Accidental{Flat, Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "E double flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-10),
	})

	key[D] = []Accidental{Flat, Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "A double flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-11),
	})

	key[G] = []Accidental{Flat, Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "D double flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-12),
	})

	key[C] = []Accidental{Flat, Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "G double flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-13),
	})

	key[F] = []Accidental{Flat, Flat}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "C double flat major circle of fifths",
		expected: key,
		actual:   KeySignatureFromCircleOfFifths(-14),
	})
}

func TestKeySignatureFromScale(t *testing.T) {
	key := KeySignature{}
	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "C Ionian scale",
		expected: key,
		actual: KeySignatureFromScale(
			LetterAndAccidentals{NoteLetter: C}, Ionian,
		),
	}, keySignatureTestCase{
		label:    "D Dorian scale",
		expected: key,
		actual: KeySignatureFromScale(
			LetterAndAccidentals{NoteLetter: D}, Dorian,
		),
	}, keySignatureTestCase{
		label:    "E Phrygian scale",
		expected: key,
		actual: KeySignatureFromScale(
			LetterAndAccidentals{NoteLetter: E}, Phrygian,
		),
	}, keySignatureTestCase{
		label:    "F Lydian scale",
		expected: key,
		actual: KeySignatureFromScale(
			LetterAndAccidentals{NoteLetter: F}, Lydian,
		),
	}, keySignatureTestCase{
		label:    "G Mixolydian scale",
		expected: key,
		actual: KeySignatureFromScale(
			LetterAndAccidentals{NoteLetter: G}, Mixolydian,
		),
	}, keySignatureTestCase{
		label:    "A Aeolian scale",
		expected: key,
		actual: KeySignatureFromScale(
			LetterAndAccidentals{NoteLetter: A}, Aeolian,
		),
	}, keySignatureTestCase{
		label:    "B Locrian scale",
		expected: key,
		actual: KeySignatureFromScale(
			LetterAndAccidentals{NoteLetter: B}, Locrian,
		),
	})
}

func TestKeySignatureFromScaleAccidentals(t *testing.T) {
	key := KeySignature{}
	key[C] = []Accidental{Sharp}
	key[D] = []Accidental{Sharp}
	key[E] = []Accidental{Sharp}
	key[F] = []Accidental{Sharp}
	key[G] = []Accidental{Sharp}
	key[A] = []Accidental{Sharp}
	key[B] = []Accidental{Sharp}

	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "C sharp Ionian scale",
		expected: key,
		actual: KeySignatureFromScale(
			LetterAndAccidentals{
				NoteLetter: C, Accidentals: []Accidental{Sharp},
			}, Ionian,
		),
	})

	key[C] = []Accidental{Flat}
	key[D] = []Accidental{Flat}
	key[E] = []Accidental{Flat}
	key[F] = []Accidental{Flat}
	key[G] = []Accidental{Flat}
	key[A] = []Accidental{Flat}
	key[B] = []Accidental{Flat}

	executeKeySignatureTestCase(t, keySignatureTestCase{
		label:    "C flat Ionian scale",
		expected: key,
		actual: KeySignatureFromScale(
			LetterAndAccidentals{
				NoteLetter: C, Accidentals: []Accidental{Flat},
			}, Ionian,
		),
	})
}

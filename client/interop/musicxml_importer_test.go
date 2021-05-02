package interop

import (
	"testing"
)

func TestNote(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label:    "simple notes",
		file:     "examples/note.musicxml",
		expected: `piano: (key-signature "") > c4 d4 e4 f4 g1`,
	})
}

func TestKeySignature(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label:    "simple traditional key signature",
		file:     "examples/key_signature.musicxml",
		expected: `piano: (key-signature "f+ c+ g+ d+") > c+4 d+4 e4 f+4 g+1`,
	})
}

func TestAccidental(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label:    "simple accidentals",
		file:     "examples/accidental.musicxml",
		expected: `piano: (key-signature "") > c4 d+4 e4 f-4 g++1`,
	})
}

func TestOctave(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label:    "simple octave switching",
		file:     "examples/octave.musicxml",
		expected: `piano: (key-signature "") < b4 > > c4 < a4 > d4 < < b4 > > > d4 < < b4 > e4`,
	})
}

func TestRest(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label:    "simple rests",
		file:     "examples/rest.musicxml",
		expected: `piano: (key-signature "") > c2 r2 d4 r4 e8 r8 f16 r16 g32 r32 r16 a4 r1.3333333333`,
	})
}

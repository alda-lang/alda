package importer

import (
	"alda.io/client/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNote(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple notes",
		file:  "../examples/note.musicxml",
		expected: `
midi-acoustic-grand-piano: 
	(key-signature "") 
	> c4 d4 e4 f4 | g1
`,
	})
}

func TestKeySignature(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple traditional key signature",
		file:  "../examples/key_signature.musicxml",
		expected: `
midi-acoustic-grand-piano: 
	(key-signature "f+ c+ g+ d+") 
	> c+4 d+4 e4 f+4 | g+1
`,
	})
}

func TestAccidental(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple accidentals",
		file:  "../examples/accidental.musicxml",
		expected: `
midi-acoustic-grand-piano: 
	(key-signature "") 
	> c4 d+4 e4 f-4 | g++1
`,
	})
}

func TestOctave(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple octave switching",
		file:  "../examples/octave.musicxml",
		expected: `
midi-acoustic-grand-piano: 
	(key-signature "") 
	< b4 >> c4 < a4 > d4 | << b4 >>> d4 << b4 > e4
`,
	})
}

func TestRest(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple rests",
		file:  "../examples/rest.musicxml",
		expected: `
midi-acoustic-grand-piano: 
	(key-signature "") 
	> c2 r2 | d4 r4 e8 r8 f16 r16 g32 r32 r16 | a4 r2.
`,
	})
}

func TestSlurs(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple slurs",
		file:  "../examples/slurs.musicxml",
		expected: `
midi-acoustic-grand-piano: 
	(key-signature "") 
	c4 d4~ e4~ f4~ | g4~ a4 b4~ > c4
`,
	})
}

func TestChord(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple chords",
		file:  "../examples/chord.musicxml",
		expected: `
midi-acoustic-grand-piano: 
	(key-signature "") 
	d2/g2/b2 f2 | c1/e1/g1/>c1
`,
	})
}

func TestTies1(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple ties",
		file:  "../examples/ties1.musicxml",
		expected: `
midi-acoustic-grand-piano: 
	(key-signature "") 
	> d1~ | 1~ | 4 f8~8 f8 f8 < f8~8
`,
	})
}

func TestTies2(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "ties with chords",
		file:  "../examples/ties2.musicxml",
		expected: `
midi-acoustic-grand-piano: 
	(key-signature "") 
	a4/>c4~4 < g4 > d2~ | 4~4 e4~4/g4~4
`,
		postprocess: func(updates []model.ScoreUpdate) []model.ScoreUpdate {
			// The OctaveDown before the g is parsed into the chord
			// We take it out
			indexOfChord := 2
			updates = apply(
				updates,
				nestedIndex{index: indexOfChord, chordIndex: -1},
				func(update model.ScoreUpdate) model.ScoreUpdate {
					chord := update.(model.Chord)
					chord.Events = chord.Events[:len(chord.Events)-1]
					return chord
				},
			)

			// Then we insert it in the parsed location
			updates = insert(
				updates,
				nestedIndex{index: indexOfChord + 1, chordIndex: -1},
				model.AttributeUpdate{PartUpdate: model.OctaveDown{}},
			)

			// This change produces equivalent Alda representation
			// As we're just moving an octave change outside of the note itself
			return updates
		},
	})
}

func TestTies3(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "complex ties",
		file:  "../examples/ties3.musicxml",
		expected: `
midi-acoustic-grand-piano: 
	(key-signature "") 
	c4/e4/g4/>c4~4 e8/g8~8 d8~8
`,
	})
}

func TestDots(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple dotted notes",
		file:  "../examples/dots.musicxml",
		expected: `
midi-acoustic-grand-piano: 
	(key-signature "") 
	g4.... r64 a4... r32 | r4. r8 b2 ~ | 2. r8.. r32
`,
	})
}

func TestVoices1(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple voices",
		file:  "../examples/voices1.musicxml",
		expected: `
midi-acoustic-grand-piano:  
	V1: (key-signature "") e4 f4 g4 a4 | b4 > c4 d4 e4
	V2: c4 d4 e4 f4 | g4 a4 b4 > c4
`,
	})
}

func TestVoices2(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "complex voices with padding, backup, and forward",
		file:  "../examples/voices2.musicxml",
		expected: `
midi-acoustic-grand-piano:  
	V1: (key-signature "") > c4 d4 e4 f4 | g4 a4 b4 > c4
	V2: r1 | r2 > g2
	V3: r4 g2 r4 | g1 
	V4: c4 r4 c4 r4 | r1 
`,
	})
}

func TestParts(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple parts (wind quintet) with transpositions",
		file:  "../examples/parts.musicxml",
		expected: `
midi-flute: 
	(key-signature "") 
	c4 d4 e4 f4 | g4 a4 b4 > c4 | < b4 a4 g4 f4 | e4 d4 c2
midi-oboe: 
	(key-signature "") 
	c4 e4 g4 > c4 | < g4 e4 c2 | g1 | c1
midi-clarinet: 
	(key-signature "f+ c+") 
	(transpose -2)
	d4 d4 d4 d4 | r1 | d4 d4 d4 d4 | r1
midi-bassoon: 
	(key-signature "") 
	< c1 | g1 | c1 | c1
midi-french-horn: 
	(key-signature "f+") 
	(transpose -7)
	g4 b4 g4 b4 | > d4 < b4 g2 | r1 | g1
`,
	})
}

func TestRepeat1(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple repeat",
		file:  "../examples/repeat1.musicxml",
		expected: `
midi-acoustic-grand-piano:
	[(key-signature "") > c1 | g1 <]*2
`,
	})
}

func TestRepeat2(t *testing.T) {
	executeImporterTestCases(t, importerTestCase{
		label: "simple repeat with forward repeat",
		file:  "../examples/repeat2.musicxml",
		expected: `
midi-acoustic-grand-piano:
	[(key-signature "") > c1 | g1 <]*2
`,
	})
}

//func TestRepeats2(t *testing.T) {
//	executeImporterTestCases(t, importerTestCase{
//		label: "repeats with first and second ending",
//		file:  "../examples/repeats2.musicxml",
//		expected: `
//midi-acoustic-grand-piano:
//	(key-signature "")
//	> c1
//	[
//		[e1]'1
//		[g1]'2
//	]*2
//`,
//	})
//}

//func TestRepeats3(t *testing.T) {
//	executeImporterTestCases(t, importerTestCase{
//		label: "repeats with complex octave updates",
//		file:  "../examples/repeats3.musicxml",
//		expected: `
//midi-acoustic-grand-piano:
//	(key-signature "")
//	[> d1 > d1 <<]*2 >> c1
//`,
//	})
//}

type Foo struct {
	info []interface{}
}

func Test(t *testing.T) {
	foo := Foo{
		info: []interface{}{"a", "b", "c"},
	}

	info := foo.info
	str := info[0]
	_ = str
	str = "aaa"

	assert.Equal(t, "aaa", foo.info[0])
}
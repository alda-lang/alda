package model

import (
	"fmt"
)

// An Instrument is a template for a Part.
type Instrument interface {
	Name() string
}

// A MidiInstrument uses one of the 128 instruments defined in the General MIDI
// spec.
//
// Reference: http://www.jimmenard.com/midi_ref.html#General_MIDI
type MidiInstrument struct {
	NameImpl     string
	PatchNumber  int32
	IsPercussion bool
}

// Name implements Instrument.Name by returning the name of the instrument.
func (mi MidiInstrument) Name() string {
	return mi.NameImpl
}

type midiInstrumentNames struct {
	name    string
	aliases []string
}

func mi(name string, aliases ...string) midiInstrumentNames {
	return midiInstrumentNames{name: name, aliases: aliases}
}

var midiNonPercussionInstruments = []midiInstrumentNames{
	// 1-8: piano
	mi("midi-acoustic-grand-piano", "midi-piano", "piano"),
	mi("midi-bright-acoustic-piano"),
	mi("midi-electric-grand-piano"),
	mi("midi-honky-tonk-piano"),
	mi("midi-electric-piano-1"),
	mi("midi-electric-piano-2"),
	mi("midi-harpsichord", "harpsichord"),
	mi("midi-clavi", "midi-clavinet", "clavinet"),
	// 9-16: chromatic percussion
	mi("midi-celesta", "celesta", "celeste", "midi-celeste"),
	mi("midi-glockenspiel", "glockenspiel"),
	mi("midi-music-box", "music-box"),
	mi("midi-vibraphone", "vibraphone", "vibes", "midi-vibes"),
	mi("midi-marimba", "marimba"),
	mi("midi-xylophone", "xylophone"),
	mi("midi-tubular-bells", "tubular-bells"),
	mi("midi-dulcimer", "dulcimer"),
	// 17-24: organ
	mi("midi-drawbar-organ"),
	mi("midi-percussive-organ"),
	mi("midi-rock-organ"),
	mi("midi-church-organ", "organ"),
	mi("midi-reed-organ"),
	mi("midi-accordion", "accordion"),
	mi("midi-harmonica", "harmonica"),
	mi("midi-tango-accordion"),
	// 25-32: guitar
	mi(
		"midi-acoustic-guitar-nylon", "midi-acoustic-guitar", "acoustic-guitar",
		"guitar",
	),
	mi("midi-acoustic-guitar-steel"),
	mi("midi-electric-guitar-jazz"),
	mi("midi-electric-guitar-clean", "electric-guitar-clean"),
	mi("midi-electric-guitar-palm-muted"),
	mi("midi-electric-guitar-overdrive", "electric-guitar-overdrive"),
	mi("midi-electric-guitar-distorted", "electric-guitar-distorted"),
	mi("midi-electric-guitar-harmonics", "electric-guitar-harmonics"),
	// 33-40: bass
	mi("midi-acoustic-bass", "acoustic-bass", "upright-bass"),
	mi("midi-electric-bass-finger", "electric-bass-finger", "electric-bass"),
	mi("midi-electric-bass-pick", "electric-bass-pick"),
	mi("midi-fretless-bass", "fretless-bass"),
	mi("midi-bass-slap"),
	mi("midi-bass-pop"),
	mi("midi-synth-bass-1"),
	mi("midi-synth-bass-2"),
	// 41-48: strings
	mi("midi-violin", "violin"),
	mi("midi-viola", "viola"),
	mi("midi-cello", "cello"),
	mi(
		"midi-contrabass", "string-bass", "arco-bass", "double-bass", "contrabass",
		"midi-string-bass", "midi-arco-bass", "midi-double-bass",
	),
	mi("midi-tremolo-strings"),
	mi("midi-pizzicato-strings"),
	mi("midi-orchestral-harp", "harp", "orchestral-harp", "midi-harp"),
	// no idea why this is in strings, but ok! ¯\_(ツ)_/¯
	mi("midi-timpani", "timpani"),
	// 49-56: ensemble
	mi("midi-string-ensemble-1"),
	mi("midi-string-ensemble-2"),
	mi("midi-synth-strings-1"),
	mi("midi-synth-strings-2"),
	mi("midi-choir-aahs"),
	mi("midi-voice-oohs"),
	mi("midi-synth-voice"),
	mi("midi-orchestra-hit"),
	// 57-64: brass
	mi("midi-trumpet", "trumpet"),
	mi("midi-trombone", "trombone"),
	mi("midi-tuba", "tuba"),
	mi("midi-muted-trumpet"),
	mi("midi-french-horn", "french-horn"),
	mi("midi-brass-section"),
	mi("midi-synth-brass-1"),
	mi("midi-synth-brass-2"),
	// 65-72: reed
	mi(
		"midi-soprano-saxophone", "midi-soprano-sax", "soprano-saxophone",
		"soprano-sax",
	),
	mi("midi-alto-saxophone", "midi-alto-sax", "alto-saxophone", "alto-sax"),
	mi("midi-tenor-saxophone", "midi-tenor-sax", "tenor-saxophone", "tenor-sax"),
	mi(
		"midi-baritone-saxophone", "midi-baritone-sax", "midi-bari-sax",
		"baritone-saxophone", "baritone-sax", "bari-sax",
	),
	mi("midi-oboe", "oboe"),
	mi("midi-english-horn", "english-horn"),
	mi("midi-bassoon", "bassoon"),
	mi("midi-clarinet", "clarinet"),
	// 73-80: pipe
	mi("midi-piccolo", "piccolo"),
	mi("midi-flute", "flute"),
	mi("midi-recorder", "recorder"),
	mi("midi-pan-flute", "pan-flute"),
	mi("midi-bottle", "bottle"),
	mi("midi-shakuhachi", "shakuhachi"),
	mi("midi-whistle", "whistle"),
	mi("midi-ocarina", "ocarina"),
	// 81-88: synth lead
	mi(
		"midi-square-lead", "square", "square-wave", "square-lead", "midi-square",
		"midi-square-wave",
	),
	mi(
		"midi-saw-wave", "sawtooth", "saw-wave", "saw-lead", "midi-sawtooth",
		"midi-saw-lead",
	),
	mi("midi-calliope-lead", "calliope-lead", "calliope", "midi-calliope"),
	mi(
		"midi-chiffer-lead", "chiffer-lead", "chiffer", "chiff", "midi-chiffer",
		"midi-chiff",
	),
	mi("midi-charang", "charang"),
	mi("midi-solo-vox"),
	mi("midi-fifths", "midi-sawtooth-fifths"),
	mi("midi-bass-and-lead", "midi-bass+lead"),
	// 89-96: synth pad
	mi("midi-synth-pad-new-age", "midi-pad-new-age", "midi-new-age-pad"),
	mi("midi-synth-pad-warm", "midi-pad-warm", "midi-warm-pad"),
	mi("midi-synth-pad-polysynth", "midi-pad-polysynth", "midi-polysynth-pad"),
	mi("midi-synth-pad-choir", "midi-pad-choir", "midi-choir-pad"),
	mi(
		"midi-synth-pad-bowed", "midi-pad-bowed", "midi-bowed-pad",
		"midi-pad-bowed-glass", "midi-bowed-glass-pad",
	),
	mi(
		"midi-synth-pad-metallic", "midi-pad-metallic", "midi-metallic-pad",
		"midi-pad-metal", "midi-metal-pad",
	),
	mi("midi-synth-pad-halo", "midi-pad-halo", "midi-halo-pad"),
	mi("midi-synth-pad-sweep", "midi-pad-sweep", "midi-sweep-pad"),
	// 97-104: synth effects
	mi("midi-fx-rain", "midi-fx-ice-rain", "midi-rain", "midi-ice-rain"),
	mi("midi-fx-soundtrack", "midi-soundtrack"),
	mi("midi-fx-crystal", "midi-crystal"),
	mi("midi-fx-atmosphere", "midi-atmosphere"),
	mi("midi-fx-brightness", "midi-brightness"),
	mi("midi-fx-goblins", "midi-fx-goblin", "midi-goblins", "midi-goblin"),
	mi("midi-fx-echoes", "midi-fx-echo-drops", "midi-echoes", "midi-echo-drops"),
	mi("midi-fx-sci-fi", "midi-sci-fi"),
	// 105-112: "ethnic" (sigh)
	mi("midi-sitar", "sitar"),
	mi("midi-banjo", "banjo"),
	mi("midi-shamisen", "shamisen"),
	mi("midi-koto", "koto"),
	mi("midi-kalimba", "kalimba"),
	mi("midi-bagpipes", "bagpipes"),
	mi("midi-fiddle"),
	mi(
		"midi-shehnai", "shehnai", "shahnai", "shenai", "shanai", "midi-shahnai",
		"midi-shenai", "midi-shanai",
	),
	// 113-120: percussive
	mi("midi-tinkle-bell", "midi-tinker-bell"),
	mi("midi-agogo"),
	mi("midi-steel-drums", "midi-steel-drum", "steel-drums", "steel-drum"),
	mi("midi-woodblock"),
	mi("midi-taiko-drum"),
	mi("midi-melodic-tom"),
	mi("midi-synth-drum"),
	mi("midi-reverse-cymbal"),
	// 121-128: sound effects
	mi("midi-guitar-fret-noise"),
	mi("midi-breath-noise"),
	mi("midi-seashore"),
	mi("midi-bird-tweet"),
	mi("midi-telephone-ring"),
	mi("midi-helicopter"),
	mi("midi-applause"),
	mi("midi-gunshot", "midi-gun-shot"),
}

var midiPercussionInstruments = []midiInstrumentNames{
	mi("midi-percussion", "percussion"),
}

// InstrumentsList returns the list of instruments available to use in an Alda
// score.
func InstrumentsList() []string {
	list := []string{}

	for _, instrument := range midiNonPercussionInstruments {
		list = append(list, instrument.name)
	}

	for _, instrument := range midiPercussionInstruments {
		list = append(list, instrument.name)
	}

	return list
}

var stockInstruments = map[string]Instrument{}

func init() {
	for i, instrumentNames := range midiNonPercussionInstruments {
		instrument := MidiInstrument{
			NameImpl: instrumentNames.name, PatchNumber: int32(i),
		}
		stockInstruments[instrumentNames.name] = instrument
		for _, alias := range instrumentNames.aliases {
			stockInstruments[alias] = instrument
		}
	}

	for _, instrumentNames := range midiPercussionInstruments {
		instrument := MidiInstrument{
			NameImpl: instrumentNames.name, IsPercussion: true,
		}
		stockInstruments[instrumentNames.name] = instrument
		for _, alias := range instrumentNames.aliases {
			stockInstruments[alias] = instrument
		}
	}
}

// stockInstrument returns a stock instrument, given an identifier which is the
// name or alias of a stock instrument.
//
// Returns an error if the identifier is not recognized as the name or alias of
// a stock instrument.
func stockInstrument(identifier string) (Instrument, error) {
	instrument, hit := stockInstruments[identifier]

	if !hit {
		return nil, fmt.Errorf("Unrecognized instrument: %s", identifier)
	}

	return instrument, nil
}

// stockInstrumentName returns the name of a stock instrument, given an
// identifier which is the name or alias of a stock instrument.
//
// Returns an error if the identifier is not recognized as the name or alias of
// a stock instrument.
func stockInstrumentName(identifier string) (string, error) {
	instrument, err := stockInstrument(identifier)
	if err != nil {
		return "", err
	}

	return instrument.Name(), nil
}

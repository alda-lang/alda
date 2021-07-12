package code_generator

import (
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type generator struct {
	tokens []parser.Token
}

func (g *generator) generate(scoreUpdates []model.ScoreUpdate) {
	for _, update := range scoreUpdates {
		switch value := update.(type) {
		case model.AttributeUpdate:
			// TODO
		case model.GlobalAttributeUpdate:
			// TODO
		case model.Barline:
			barlineGenerator(value, g)
		case model.Chord:
			chordGenerator(value, g)
		case model.Cram:
			cramGenerator(value, g)
		case model.EventSequence:
			eventSequenceGenerator(value, g)
		case model.LispNil:
			// TODO
		case model.LispList:
			// TODO
		case model.Marker:
			markerGenerator(value, g)
		case model.AtMarker:
			atMarkerGenerator(value, g)
		case model.Note:
			noteGenerator(value, g)
		case model.Rest:
			restGenerator(value, g)
		case model.PartDeclaration:
			partDeclarationGenerator(value, g)
		case model.Repeat:
			repeatGenerator(value, g)
		case model.OnRepetitions:
			onRepetitionsGenerator(value, g)
		case model.VariableDefinition:
			variableDeclarationGenerator(value, g)
		case model.VariableReference:
			variableReferenceGenerator(value, g)
		case model.VoiceMarker:
			voiceMarkerGenerator(value, g)
		case model.VoiceGroupEndMarker:
			voiceGroupEndMarkerGenerator(value, g)
		}
	}
}

func (g *generator) addToken(tokenType parser.TokenType, text string) {
	g.tokens = append(g.tokens, parser.Token{TokenType: tokenType, Text: text})
}

func Generate(scoreUpdates []model.ScoreUpdate) []parser.Token {
	var g generator
	g.generate(scoreUpdates)
	g.addToken(parser.EOF, "")
	return g.tokens
}

func barlineGenerator(_ model.Barline, g *generator) {
	g.addToken(parser.Barline, "|")
}

func chordGenerator(chord model.Chord, g *generator) {
	for index, update := range chord.Events {
		g.generate([]model.ScoreUpdate{update})

		if index < len(chord.Events) - 1 {
			g.addToken(parser.Separator, "/")
		}
	}
}

func cramGenerator(cram model.Cram, g *generator) {
	g.addToken(parser.CramOpen, "{")
	g.generate(cram.Events)
	g.addToken(parser.CramClose, "}")
}

func eventSequenceGenerator(es model.EventSequence, g *generator) {
	g.addToken(parser.EventSeqOpen, "[")
	g.generate(es.Events)
	g.addToken(parser.EventSeqClose, "]")
}

func markerGenerator(marker model.Marker, g *generator) {
	g.addToken(parser.Marker, "%" + marker.Name)
}

func atMarkerGenerator(atMarker model.AtMarker, g *generator) {
	g.addToken(parser.AtMarker, "@" + atMarker.Name)
}

func pitchGenerator(pitch model.PitchIdentifier, g *generator) {
	// PitchIdentifier can be a LetterAndAccidentals or a MidiNoteNumber
	// However, note's are never generated directly with a MidiNoteNumber pitch
	// When parsing Alda itself, MidiNoteNumber is only used with Lisp
	// When building Alda from the MusicXML importer, MidiNoteNumbers are
	// transformed to LetterAndAccidentals during optimization
	// Then it is valid to assume that any pitch here is a LetterAndAccidentals
	if reflect.TypeOf(pitch) == reflect.TypeOf(model.MidiNoteNumber{}) {
		log.Error().Msg(`Alda code generator cannot process note with pitch type MidiNoteNumber`)
	}

	laa := pitch.(model.LetterAndAccidentals)
	g.addToken(parser.NoteLetter, strings.ToLower(laa.NoteLetter.String()))

	for _, accidental := range laa.Accidentals {
		switch accidental {
		case model.Flat:
			g.addToken(parser.Flat, "-")
		case model.Natural:
			g.addToken(parser.Natural, "_")
		case model.Sharp:
			g.addToken(parser.Sharp, "+")
		}
	}
}

func durationGenerator(duration model.Duration, g *generator) {
	for _, component := range duration.Components {
		switch value := component.(type) {
		case model.Barline:
			barlineGenerator(value, g)
		case model.NoteLength:
			noteLength := fmt.Sprintf("%f", value.Denominator) +
				strings.Repeat(".", int(value.Dots))
			g.addToken(parser.NoteLength, noteLength)
		case model.NoteLengthMs:
			noteLengthMs := fmt.Sprintf("%f", value.Quantity) + "ms"
			g.addToken(parser.NoteLengthMs, noteLengthMs)
		case model.NoteLengthBeats:
			// NoteLengthBeats is only generated from Lisp
			log.Error().Msg(`Alda code generator cannot process duration with component type NoteLengthBeats`)
		}
	}
}

func noteGenerator(note model.Note, g *generator) {
	pitchGenerator(note.Pitch, g)
	durationGenerator(note.Duration, g)
	if note.Slurred {
		g.addToken(parser.Tie, "~")
	}
}

func restGenerator(rest model.Rest, g *generator) {
	g.addToken(parser.RestLetter, "r")
	durationGenerator(rest.Duration, g)
}

func partDeclarationGenerator(decl model.PartDeclaration, g *generator) {
	for i, name := range decl.Names {
		if i > 0 {
			g.addToken(parser.Separator, "/")
		}
		g.addToken(parser.Name, name)
	}

	if decl.Alias != "" {
		g.addToken(parser.Alias, "\"" + decl.Alias + "\"")
	}

	g.addToken(parser.Colon, ":")
}

func repeatGenerator(repeat model.Repeat, g *generator) {
	g.generate([]model.ScoreUpdate{repeat.Event})
	g.addToken(parser.Repeat, "*" + strconv.Itoa(int(repeat.Times)))
}

func onRepetitionsGenerator(or model.OnRepetitions, g *generator) {
	g.generate([]model.ScoreUpdate{or.Event})

	repetitionsText := "'"
	for i, repetition := range or.Repetitions {
		if i > 0 {
			repetitionsText += ","
		}
		if repetition.First == repetition.Last {
			repetitionsText += strconv.Itoa(int(repetition.First))
		} else {
			repetitionsText += strconv.Itoa(int(repetition.First))
			repetitionsText += "-"
			repetitionsText += strconv.Itoa(int(repetition.Last))
		}
	}

	g.addToken(parser.Repetitions, repetitionsText)
}

func variableDeclarationGenerator(vd model.VariableDefinition, g *generator) {
	g.addToken(parser.Name, vd.VariableName)
	g.addToken(parser.Equals, "=")
	g.generate(vd.Events)
}

func variableReferenceGenerator(vr model.VariableReference, g *generator) {
	g.addToken(parser.Name, vr.VariableName)
}

func voiceMarkerGenerator(vm model.VoiceMarker, g *generator) {
	g.addToken(
		parser.VoiceMarker,
		"V" + strconv.Itoa(int(vm.VoiceNumber)) + ":",
	)
}

func voiceGroupEndMarkerGenerator(_ model.VoiceGroupEndMarker, g *generator) {
	g.addToken(parser.VoiceMarker, "V0:")
}

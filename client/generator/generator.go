package generator

import (
	"alda.io/client/model"
	"alda.io/client/parser"
	"strconv"
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
			// TODO
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
			// TODO
		case model.Rest:
			// TODO
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

package generator

import (
	"alda.io/client/model"
	"alda.io/client/parser"
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
			// TODO
		case model.EventSequence:
			eventSequenceGenerator(value, g)
		case model.LispNil:
			// TODO
		case model.LispList:
			// TODO
		case model.Marker:
			// TODO
		case model.AtMarker:
			// TODO
		case model.Note:
			// TODO
		case model.Rest:
			// TODO
		case model.PartDeclaration:
			partDeclarationGenerator(value, g)
		case model.Repeat:
			// TODO
		case model.OnRepetitions:
			// TODO
		case model.VariableDefinition:
			// TODO
		case model.VariableReference:
			// TODO
		case model.VoiceMarker:
			// TODO
		case model.VoiceGroupEndMarker:
			// TODO
		}
	}
}

func (g *generator) addToken(tokenType parser.TokenType, text string) {
	g.tokens = append(g.tokens, parser.Token{TokenType: tokenType, Text: text})
}

func Generate(scoreUpdates []model.ScoreUpdate) []parser.Token {
	var g generator
	g.generate(scoreUpdates)
	return g.tokens
}

func barlineGenerator(_ model.Barline, g *generator) {
	g.addToken(parser.Barline, "|")
}

func eventSequenceGenerator(es model.EventSequence, g *generator) {
	g.addToken(parser.EventSeqOpen, "[")
	g.generate(es.Events)
	g.addToken(parser.EventSeqClose, "]")
}

func partDeclarationGenerator(decl model.PartDeclaration, g *generator) {
	for i, name := range decl.Names {
		if i > 0 {
			g.addToken(parser.Separator, "/")
		}
		g.addToken(parser.Name, name)
	}
	g.addToken(parser.Colon, ":")
}

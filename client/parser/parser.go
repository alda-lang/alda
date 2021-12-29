package parser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"alda.io/client/color"
	"alda.io/client/help"
	log "alda.io/client/logging"
	model "alda.io/client/model"
)

type parser struct {
	filename string
	input    []Token
	updates  []model.ScoreUpdate
	current  int
	// When true, source context is _not_ included in parsed tokens. This is
	// useful for testing, e.g. for checking the equality of a list of expected
	// tokens, agnostic of source context like line and column numbers.
	suppressSourceContext bool
}

// A parseOption is a function that customizes a parser instance.
type parseOption func(*parser)

// SuppressSourceContext customizes a parser to ignore source context
func SuppressSourceContext(parser *parser) {
	parser.suppressSourceContext = true
}

func (p *parser) sourceContext(token Token) model.AldaSourceContext {
	if p.suppressSourceContext {
		return model.AldaSourceContext{}
	}

	return token.sourceContext
}

func newParser(filename string, tokens []Token, opts ...parseOption) *parser {
	parser := &parser{
		filename: filename,
		input:    tokens,
		updates:  []model.ScoreUpdate{},
		current:  0,
	}

	for _, opt := range opts {
		opt(parser)
	}

	return parser
}

////////////////////////////////////////////////////////////////////////////////

func (p *parser) peek() Token {
	return p.input[p.current]
}

func (p *parser) previous() Token {
	return p.input[p.current-1]
}

// CAUTION: This will panic if the _current_ token is EOF, because that means
// there is no next token, so the index will be out of bounds.
func (p *parser) next() Token {
	return p.input[p.current+1]
}

func (p *parser) check(types ...TokenType) bool {
	for _, tokenType := range types {
		if p.peek().tokenType == tokenType {
			return true
		}
	}

	return false
}

func (p *parser) advance() Token {
	if !p.check(EOF) {
		p.current++
	}

	return p.previous()
}

func (p *parser) match(types ...TokenType) (Token, bool) {
	for _, tokenType := range types {
		if p.check(tokenType) {
			return p.advance(), true
		}
	}

	return Token{}, false
}

func (p *parser) addUpdate(update model.ScoreUpdate) {
	log.Debug().Str("update", fmt.Sprintf("%#v", update)).Msg("Adding update.")
	p.updates = append(p.updates, update)
}

func (p *parser) errorAtToken(token Token, msg string) *model.AldaSourceError {
	return &model.AldaSourceError{
		Context: token.sourceContext,
		Err:     fmt.Errorf("%s", msg),
	}
}

func (p *parser) unexpectedTokenError(
	token Token, context string,
) *model.AldaSourceError {
	if context != "" {
		context = " " + context
	}

	tokenText := ""
	if len(token.text) > 0 {
		tokenText = fmt.Sprintf(" `%s`", token.text)
	}

	msg := fmt.Sprintf(
		"Unexpected %s%s%s", token.tokenType.String(), tokenText, context,
	)

	return p.errorAtToken(token, msg)
}

func (p *parser) consume(tokenType TokenType, context string) (Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}

	return Token{}, p.unexpectedTokenError(p.peek(), context)
}

func (p *parser) lispForm(context string) (ASTNode, error) {
	if token, matched := p.match(Symbol); matched {
		return ASTNode{
			Type:          LispSymbolNode,
			SourceContext: p.sourceContext(token),
			Literal:       token.text,
		}, nil
	}

	if token, matched := p.match(Number); matched {
		return ASTNode{
			Type:          LispNumberNode,
			SourceContext: p.sourceContext(token),
			Literal:       token.literal,
		}, nil
	}

	if token, matched := p.match(String); matched {
		return ASTNode{
			Type:          LispStringNode,
			SourceContext: p.sourceContext(token),
			Literal:       token.literal,
		}, nil
	}

	if _, matched := p.match(LeftParen); matched {
		return p.lispList()
	}

	return ASTNode{}, p.unexpectedTokenError(p.peek(), context)
}

func (p *parser) lispList() (ASTNode, error) {
	// NB: This assumes the initial LeftParen token was already consumed.
	list := ASTNode{
		SourceContext: p.sourceContext(p.previous()),
		Type:          LispListNode,
	}

	for token := p.peek(); token.tokenType != RightParen; token = p.peek() {
		if _, matched := p.match(EOF); matched {
			return ASTNode{}, p.errorAtToken(token, "unterminated S-expression")
		}

		quoteToken, quoted := p.match(SingleQuote)

		form, err := p.lispForm("in S-expression")
		if err != nil {
			return ASTNode{}, err
		}

		if quoted {
			form = ASTNode{
				Type:          LispQuotedFormNode,
				SourceContext: p.sourceContext(quoteToken),
				Children:      []ASTNode{form},
			}
		}

		list.Children = append(list.Children, form)
	}

	if _, err := p.consume(RightParen, "in S-expression"); err != nil {
		return ASTNode{}, err
	}

	return list, nil
}

func (p *parser) sexp() (ASTNode, error) {
	// NB: This assumes the initial LeftParen token was already consumed.
	list, err := p.lispList()
	if err != nil {
		return ASTNode{}, err
	}

	return p.singleOrRepeated(list), nil
}

func (p *parser) partDeclaration() (ASTNode, error) {
	nameNode := func(nameToken Token) ASTNode {
		return ASTNode{
			Type:          PartNameNode,
			SourceContext: p.sourceContext(nameToken),
			Literal:       nameToken.text,
		}
	}

	// NB: This assumes the initial Name token was already consumed.
	nameToken := p.previous()

	namesNode := ASTNode{
		Type:          PartNamesNode,
		SourceContext: p.sourceContext(nameToken),
		Children:      []ASTNode{nameNode(nameToken)},
	}

	for {
		if _, matched := p.match(Separator); !matched {
			break
		}

		name, err := p.consume(Name, "in part declaration")
		if err != nil {
			return ASTNode{}, err
		}

		namesNode.Children = append(namesNode.Children, nameNode(name))
	}

	partDecl := ASTNode{
		Type:          PartDeclarationNode,
		SourceContext: p.sourceContext(nameToken),
		Children:      []ASTNode{namesNode},
	}

	if alias, matched := p.match(Alias); matched {
		partDecl.Children = append(
			partDecl.Children,
			ASTNode{
				Type:          PartAliasNode,
				SourceContext: p.sourceContext(alias),
				Literal:       alias.literal,
			},
		)
	}

	if _, err := p.consume(Colon, "in part declaration"); err != nil {
		return ASTNode{}, err
	}

	return partDecl, nil
}

func (p *parser) looksLikePartDeclaration() bool {
	current := p.peek().tokenType

	if current == EOF {
		return false
	}

	next := p.next().tokenType

	return current == Name &&
		(next == Alias || next == Separator || next == Colon)
}

func (p *parser) partEvents() (ASTNode, error) {
	partEvents := ASTNode{
		Type:          EventSequenceNode,
		SourceContext: p.sourceContext(p.peek()),
		Children:      []ASTNode{},
	}

	// Keep consuming events until we reach either a part declaration or EOF.
	for {
		if p.check(EOF) || p.looksLikePartDeclaration() {
			break
		}

		event, err := p.innerEvent()
		if err != nil {
			return ASTNode{}, err
		}

		partEvents.Children = append(partEvents.Children, event)
	}

	return partEvents, nil
}

func (p *parser) part() (ASTNode, error) {
	// NB: This assumes the initial Name token was already consumed.
	nameToken := p.previous()

	partDecl, err := p.partDeclaration()
	if err != nil {
		return ASTNode{}, err
	}

	partEvents, err := p.partEvents()
	if err != nil {
		return ASTNode{}, err
	}

	return ASTNode{
		Type:          PartNode,
		SourceContext: p.sourceContext(nameToken),
		Children:      []ASTNode{partDecl, partEvents},
	}, nil
}

// An "implicit part" is an AST node that can contain:
//
// * Initial variable definitions at the top of the file
// * Initial S-expressions at the top of the file, like global attributes
// * Events (see `innerEvent`) without a part definition, in the context of
//   continuing a previous score, e.g. in REPL input.
func (p *parser) implicitPart() (ASTNode, error) {
	partEvents, err := p.partEvents()
	if err != nil {
		return ASTNode{}, err
	}

	return ASTNode{
		Type:          ImplicitPartNode,
		SourceContext: partEvents.SourceContext,
		Children:      []ASTNode{partEvents},
	}, nil
}

func (p *parser) variableDefinition() (ASTNode, error) {
	// NB: This assumes the initial Name token was already consumed.
	nameToken := p.previous()

	nameNode := ASTNode{
		Type:          VariableNameNode,
		SourceContext: p.sourceContext(nameToken),
		Literal:       nameToken.text,
	}

	definitionLine := nameToken.sourceContext.Line

	equalsToken, err := p.consume(Equals, "in variable definition")
	if err != nil {
		return ASTNode{}, err
	}

	if p.check(EOF) || p.peek().sourceContext.Line > definitionLine {
		return ASTNode{}, &model.AldaSourceError{
			Context: nameToken.sourceContext,
			Err: fmt.Errorf(
				"there must be at least one event on the same line as the '='",
			),
		}
	}

	eventsNode := ASTNode{
		Type:          EventSequenceNode,
		SourceContext: p.sourceContext(p.peek()),
		Children:      []ASTNode{},
	}

	for t := p.peek(); t.sourceContext.Line == definitionLine; t = p.peek() {
		if t.tokenType == EOF {
			break
		}

		node, err := p.innerEvent()
		if err != nil {
			return ASTNode{}, err
		}

		eventsNode.Children = append(eventsNode.Children, node)
	}

	definitionNode := ASTNode{
		Type:          VariableDefinitionNode,
		SourceContext: p.sourceContext(equalsToken),
		Children:      []ASTNode{nameNode, eventsNode},
	}

	return definitionNode, nil
}

func (p *parser) singleOrRepeated(node ASTNode) ASTNode {
	if token, matched := p.match(Repeat); matched {
		return ASTNode{
			SourceContext: p.sourceContext(token),
			Type:          RepeatNode,
			Children: []ASTNode{
				node,
				{
					Type:    TimesNode,
					Literal: token.literal,
				},
			},
		}
	}

	return node
}

func (p *parser) variableReference() (ASTNode, error) {
	// NB: This assumes the initial Name token was already consumed.
	nameToken := p.previous()

	reference := ASTNode{
		Type:          VariableReferenceNode,
		SourceContext: p.sourceContext(nameToken),
		Literal:       nameToken.text,
	}

	return p.singleOrRepeated(reference), nil
}

func (p *parser) variableDefinitionOrReference() (ASTNode, error) {
	// NB: This assumes the initial Name token was already consumed.
	if p.check(Equals) {
		return p.variableDefinition()
	}

	return p.variableReference()
}

func (p *parser) partOrVariableDefinition() (ASTNode, error) {
	// NB: This assumes the initial Name token was already consumed.
	if p.check(Equals) {
		return p.variableDefinition()
	}

	return p.part()
}

func (p *parser) octaveSet() (ASTNode, error) {
	// NB: This assumes the OctaveSet token was already consumed.
	token := p.previous()

	return ASTNode{
		Type:          OctaveSetNode,
		SourceContext: p.sourceContext(token),
		Literal:       token.literal,
	}, nil
}

func (p *parser) matchDurationComponent() (Token, bool) {
	return p.match(NoteLength, NoteLengthMs)
}

func (p *parser) durationComponent() ASTNode {
	// NB: This assumes the duration component token was already consumed.
	token := p.previous()

	switch token.tokenType {
	case NoteLength:
		noteLength := token.literal.(noteLength)

		nlNode := ASTNode{
			Type:          NoteLengthNode,
			SourceContext: p.sourceContext(token),
			Children: []ASTNode{
				{
					Type:    DenominatorNode,
					Literal: noteLength.denominator,
				},
			},
		}

		if noteLength.dots > 0 {
			nlNode.Children = append(nlNode.Children, ASTNode{
				Type:    DotsNode,
				Literal: noteLength.dots,
			})
		}

		return nlNode
	case NoteLengthMs:
		return ASTNode{
			Type:          NoteLengthMsNode,
			SourceContext: p.sourceContext(token),
			Literal:       token.literal,
		}
	}

	// We shouldn't get here.
	return ASTNode{}
}

func (p *parser) duration() ASTNode {
	durationNode := func(componentNodes []ASTNode) ASTNode {
		return ASTNode{
			Type:          DurationNode,
			SourceContext: componentNodes[0].SourceContext,
			Children:      componentNodes,
		}
	}

	barlineNode := func(token Token) ASTNode {
		return ASTNode{
			Type:          BarlineNode,
			SourceContext: p.sourceContext(token),
		}
	}

	// NB: This assumes the initial duration component token was already consumed.
	componentNodes := []ASTNode{p.durationComponent()}

	// Repeatedly parse duration components.
	for {
		// Repeatedly parse barlines amongst the duration components.
		for {
			if token, matched := p.match(Barline); matched {
				componentNodes = append(componentNodes, barlineNode(token))
			} else {
				break
			}
		}

		// Take note of the current position. If we encounter a tie that ends up
		// actually being a slur (i.e. it isn't followed by a note length), we can
		// backtrack to this position, return the duration, and let the parser
		// consume the slur as part of e.g. a note.
		beforeTies := p.current

		if _, matched := p.match(Tie); !matched {
			return durationNode(componentNodes)
		}

		// We'll stash any barlines that we encounter here temporarily. We'll add
		// them to the duration components iff we aren't going to backtrack and
		// consume them outside of the duration.
		barlines := []ASTNode{}

		for {
			if token, matched := p.match(Barline); matched {
				barlines = append(barlines, barlineNode(token))
			} else {
				break
			}
		}

		// In some cases, it makes sense to have extraneous ties, e.g. when you're
		// tying a duration across a barline and it feels right to have a tie on
		// either side of the barline. So we'll consume any additional ties here.
		for {
			if _, matched := p.match(Tie); !matched {
				break
			}
		}

		if _, matched := p.matchDurationComponent(); !matched {
			p.current = beforeTies
			return durationNode(componentNodes)
		}

		componentNodes = append(componentNodes, barlines...)
		componentNodes = append(componentNodes, p.durationComponent())
	}
}

func (p *parser) note() (ASTNode, error) {
	// NB: This assumes the initial NoteLetter token was already consumed.
	noteLetterToken := p.previous()

	laaNode := ASTNode{
		Type:          NoteLetterAndAccidentalsNode,
		SourceContext: p.sourceContext(noteLetterToken),
		Children: []ASTNode{
			{
				Type:          NoteLetterNode,
				SourceContext: p.sourceContext(noteLetterToken),
				Literal:       noteLetterToken.literal,
			},
		},
	}

	accidentalNodes := []ASTNode{}

AccidentalsLoop:
	for {
		if token, matched := p.match(Flat); matched {
			accidentalNodes = append(accidentalNodes, ASTNode{
				Type:          FlatNode,
				SourceContext: p.sourceContext(token),
			})
		} else if token, matched := p.match(Natural); matched {
			accidentalNodes = append(accidentalNodes, ASTNode{
				Type:          NaturalNode,
				SourceContext: p.sourceContext(token),
			})
		} else if token, matched := p.match(Sharp); matched {
			accidentalNodes = append(accidentalNodes, ASTNode{
				Type:          SharpNode,
				SourceContext: p.sourceContext(token),
			})
		} else {
			break AccidentalsLoop
		}
	}

	if len(accidentalNodes) > 0 {
		laaNode.Children = append(laaNode.Children, ASTNode{
			Type:          NoteAccidentalsNode,
			SourceContext: accidentalNodes[0].SourceContext,
			Children:      accidentalNodes,
		})
	}

	noteNode := ASTNode{
		Type:          NoteNode,
		SourceContext: p.sourceContext(noteLetterToken),
		Children:      []ASTNode{laaNode},
	}

	if _, matched := p.matchDurationComponent(); matched {
		noteNode.Children = append(noteNode.Children, p.duration())
	}

	if tie, matched := p.match(Tie); matched {
		noteNode.Children = append(noteNode.Children, ASTNode{
			Type:          TieNode,
			SourceContext: p.sourceContext(tie),
		})
	}

	return noteNode, nil
}

func (p *parser) rest() ASTNode {
	// NB: This assumes the initial RestLetter token was already consumed.
	token := p.previous()

	rest := ASTNode{
		Type:          RestNode,
		SourceContext: p.sourceContext(token),
	}

	if _, matched := p.matchDurationComponent(); matched {
		rest.Children = []ASTNode{p.duration()}
	}

	return rest
}

func (p *parser) noteOrRest() (ASTNode, error) {
	// NB: This assumes the initial NoteLetter/RestLetter was already consumed.
	switch letter := p.previous(); letter.tokenType {
	case NoteLetter:
		return p.note()
	case RestLetter:
		return p.rest(), nil
	default:
		return ASTNode{}, p.unexpectedTokenError(letter, "in note/rest")
	}
}

func (p *parser) nodesBetweenNotesInChord() ([]ASTNode, error) {
	updates := []ASTNode{}

	for {
		if token, matched := p.match(OctaveUp); matched {
			updates = append(updates, ASTNode{
				Type:          OctaveUpNode,
				SourceContext: p.sourceContext(token),
			})
		} else if token, matched := p.match(OctaveDown); matched {
			updates = append(updates, ASTNode{
				Type:          OctaveDownNode,
				SourceContext: p.sourceContext(token),
			})
		} else if _, matched := p.match(OctaveSet); matched {
			octaveSetNode, err := p.octaveSet()
			if err != nil {
				return nil, err
			}
			updates = append(updates, octaveSetNode)
		} else if _, matched := p.match(LeftParen); matched {
			sexp, err := p.lispList()
			if err != nil {
				return nil, err
			}
			updates = append(updates, sexp)
		} else {
			return updates, nil
		}
	}
}

// Parses a note or chord. A chord contains multiple chords and rests, not to
// mention attribute changes, so any of those will be parsed too in the process.
func (p *parser) noteRestOrChord() (ASTNode, error) {
	// NB: This assumes the initial NoteLetter/RestLetter was already consumed.

	// The cumulative list of nodes. Depending on whether this is a chord, the
	// nodes will either be emitted as part of the chord, or emitted individually.
	allNodes := []ASTNode{}

	type maybeRepeat struct {
		sourceContext model.AldaSourceContext
		times         int32
	}

	// We're essentially using this as a nil value. Below, we check whether
	// `repeat.times` > 0, which will be true if we don't reassign `repeat`.
	repeat := maybeRepeat{}

	for {
		noteOrRest, err := p.noteOrRest()
		if err != nil {
			return ASTNode{}, err
		}

		if token, matched := p.match(Repeat); matched {
			allNodes = append(allNodes, noteOrRest)
			repeat = maybeRepeat{
				sourceContext: p.sourceContext(token),
				times:         token.literal.(int32),
			}
			break
		}

		// The nodes for just this iteration of the loop
		nodes := []ASTNode{noteOrRest}

		// HACK to work around the complexity that comes with allowing chords to
		// include attribute changes in between the notes. This is all easier if we
		// can make this function return a single node and not a list of nodes (see
		// how we are using it in `innerEvent`.) That's easy enough if the result of
		// the parsing that this function does is a single node, like a note, a
		// rest, or a chord. However, if we start to parse a chord, and we parse
		// something like:
		//
		// * note, (e.g. `a4`)
		// * octave up (`>`)
		// * something besides `/`
		//
		// Then we just want this function to return the note, and we want to
		// backtrack so that the next invocation of `innerEvent` can parse the
		// octave up and all subsequent events separately.
		//
		// Here, we capture the current parser position after we parse each note or
		// rest in the chord, so that we can backtrack to that position if needed.
		backtrackPosition := p.current

		nodesBeforeSeparator, err := p.nodesBetweenNotesInChord()
		if err != nil {
			return ASTNode{}, err
		}

		if _, matched := p.match(Separator); !matched {
			// See HACK comment above.
			p.current = backtrackPosition

			allNodes = append(allNodes, nodes...)

			break
		}

		nodes = append(nodes, nodesBeforeSeparator...)

		nodesAfterSeparator, err := p.nodesBetweenNotesInChord()
		if err != nil {
			return ASTNode{}, err
		}
		nodes = append(nodes, nodesAfterSeparator...)

		allNodes = append(allNodes, nodes...)

		if _, matched := p.match(NoteLetter, RestLetter); !matched {
			return ASTNode{}, p.unexpectedTokenError(p.peek(), "in chord")
		}
	}

	notesCount := 0
	for _, node := range allNodes {
		switch node.Type {
		case NoteNode, RestNode:
			notesCount++
		}
	}

	if notesCount > 1 {
		allNodes = []ASTNode{
			{
				Type:          ChordNode,
				SourceContext: allNodes[0].SourceContext,
				Children:      allNodes,
			},
		}
	}

	if repeat.times > 0 {
		if len(allNodes) != 1 {
			panic(fmt.Sprintf("Expected a single node in %#v", allNodes))
		}

		return ASTNode{
			SourceContext: repeat.sourceContext,
			Type:          RepeatNode,
			Children: []ASTNode{
				allNodes[0],
				{
					Type:    TimesNode,
					Literal: repeat.times,
				},
			},
		}, nil
	}

	return allNodes[0], nil
}

func (p *parser) eventSeq() (ASTNode, error) {
	// NB: This assumes the initial EventSeqOpen token was already consumed.
	eventSeqOpenToken := p.previous()

	eventNodes := []ASTNode{}

	for token := p.peek(); token.tokenType != EventSeqClose; token = p.peek() {
		if _, matched := p.match(EOF); matched {
			return ASTNode{}, p.errorAtToken(token, "unterminated event sequence")
		}

		eventNode, err := p.innerEvent()
		if err != nil {
			return ASTNode{}, err
		}

		if token, matched := p.match(Repetitions); matched {
			repetitionsNode := ASTNode{
				Type:          RepetitionsNode,
				SourceContext: p.sourceContext(token),
				Literal:       token.literal,
			}

			eventNode = ASTNode{
				Type:          OnRepetitionsNode,
				SourceContext: p.sourceContext(token),
				Children:      []ASTNode{eventNode, repetitionsNode},
			}
		}

		eventNodes = append(eventNodes, eventNode)
	}

	if _, err := p.consume(EventSeqClose, "in event sequence"); err != nil {
		return ASTNode{}, err
	}

	eventSeq := ASTNode{
		Type:          EventSequenceNode,
		SourceContext: p.sourceContext(eventSeqOpenToken),
		Children:      eventNodes,
	}

	return p.singleOrRepeated(eventSeq), nil
}

func (p *parser) cram() (ASTNode, error) {
	// NB: This assumes the initial CramOpen token was already consumed.
	cramOpenToken := p.previous()

	allEvents := []ASTNode{}

	for token := p.peek(); token.tokenType != CramClose; token = p.peek() {
		if _, matched := p.match(EOF); matched {
			return ASTNode{}, p.errorAtToken(token, "unterminated cram expression")
		}

		event, err := p.innerEvent()
		if err != nil {
			return ASTNode{}, err
		}
		allEvents = append(allEvents, event)
	}

	if _, err := p.consume(CramClose, "in cram expression"); err != nil {
		return ASTNode{}, err
	}

	eventsNode := ASTNode{
		Type:          EventSequenceNode,
		SourceContext: allEvents[0].SourceContext,
		Children:      allEvents,
	}

	cram := ASTNode{
		Type:          CramNode,
		SourceContext: p.sourceContext(cramOpenToken),
		Children:      []ASTNode{eventsNode},
	}

	if _, matched := p.matchDurationComponent(); matched {
		cram.Children = append(cram.Children, p.duration())
	}

	return p.singleOrRepeated(cram), nil
}

func (p *parser) voiceNumber() (ASTNode, error) {
	// NB: This assumes the VoiceMarker token was already consumed.
	token := p.previous()

	return ASTNode{
		Type:          VoiceNumberNode,
		SourceContext: p.sourceContext(token),
		Literal:       token.literal.(int32),
	}, nil
}

func (p *parser) voice() (ASTNode, error) {
	// NB: This assumes the VoiceMarker token was already consumed.

	voiceNumber, err := p.voiceNumber()
	if err != nil {
		return ASTNode{}, err
	}

	voiceEvents := ASTNode{
		Type:          EventSequenceNode,
		SourceContext: p.sourceContext(p.peek()),
		Children:      []ASTNode{},
	}

	// Keep consuming events until we reach another voice marker (including a
	// voice group end marker), a new part, EOF, or a closing ] or }.
	for {
		if p.check(EOF, VoiceMarker, EventSeqClose, CramClose) ||
			p.looksLikePartDeclaration() {
			break
		}

		event, err := p.innerEvent()
		if err != nil {
			return ASTNode{}, err
		}

		voiceEvents.Children = append(voiceEvents.Children, event)
	}

	return ASTNode{
		Type:          VoiceNode,
		SourceContext: voiceNumber.SourceContext,
		Children:      []ASTNode{voiceNumber, voiceEvents},
	}, nil
}

func (p *parser) voiceGroup() (ASTNode, error) {
	// NB: This assumes the first VoiceMarker token was already consumed.
	firstVoiceMarkerToken := p.previous()

	voiceGroupNode := ASTNode{
		Type:          VoiceGroupNode,
		SourceContext: firstVoiceMarkerToken.sourceContext,
		Children:      []ASTNode{},
	}

	for {
		voice, err := p.voice()
		if err != nil {
			return ASTNode{}, err
		}

		voiceGroupNode.Children = append(voiceGroupNode.Children, voice)

		voiceMarker, matched := p.match(VoiceMarker)
		if !matched {
			break
		}

		if voiceMarker.literal.(int32) == 0 {
			voiceGroupNode.Children = append(voiceGroupNode.Children, ASTNode{
				Type:          VoiceGroupEndMarkerNode,
				SourceContext: p.sourceContext(voiceMarker),
			})

			break
		}
	}

	return voiceGroupNode, nil
}

func (p *parser) innerEvent() (ASTNode, error) {
	if _, matched := p.match(LeftParen); matched {
		return p.sexp()
	}

	if _, matched := p.match(Name); matched {
		return p.variableDefinitionOrReference()
	}

	if _, matched := p.match(OctaveSet); matched {
		return p.octaveSet()
	}

	if token, matched := p.match(OctaveUp); matched {
		return ASTNode{
			Type:          OctaveUpNode,
			SourceContext: p.sourceContext(token),
		}, nil
	}

	if token, matched := p.match(OctaveDown); matched {
		return ASTNode{
			Type:          OctaveDownNode,
			SourceContext: p.sourceContext(token),
		}, nil
	}

	if _, matched := p.match(NoteLetter, RestLetter); matched {
		return p.noteRestOrChord()
	}

	if token, matched := p.match(Barline); matched {
		return ASTNode{
			Type:          BarlineNode,
			SourceContext: p.sourceContext(token),
		}, nil
	}

	if _, matched := p.match(EventSeqOpen); matched {
		return p.eventSeq()
	}

	if _, matched := p.match(CramOpen); matched {
		return p.cram()
	}

	if _, matched := p.match(VoiceMarker); matched {
		return p.voiceGroup()
	}

	if token, matched := p.match(Marker); matched {
		return ASTNode{
			Type:          MarkerNode,
			SourceContext: p.sourceContext(token),
			Literal:       token.literal,
		}, nil
	}

	if token, matched := p.match(AtMarker); matched {
		return ASTNode{
			Type:          AtMarkerNode,
			SourceContext: p.sourceContext(token),
			Literal:       token.literal,
		}, nil
	}

	return ASTNode{}, p.unexpectedTokenError(p.peek(), "in inner events")
}

func (p *parser) topLevel() (ASTNode, error) {
	if p.looksLikePartDeclaration() {
		p.consume(Name, "in part declaration")
		return p.part()
	}

	return p.implicitPart()
}

func (p *parser) parseAST() (ASTNode, error) {
	rootNode := ASTNode{Type: RootNode}

	for t := p.peek(); t.tokenType != EOF; t = p.peek() {
		// fmt.Printf("t: %s\n", t.String())
		node, err := p.topLevel()
		if err != nil {
			return ASTNode{}, err
		}

		rootNode.Children = append(rootNode.Children, node)
	}

	return rootNode, nil
}

// Parse a string of input into a root ASTNode.
func Parse(
	filepath string, input string, opts ...parseOption,
) (ASTNode, error) {
	defer func(start time.Time) {
		if r := recover(); r != nil {
			panic(fmt.Sprintf("Critical error while parsing %s", filepath))
		}

		log.Info().
			Str("filepath", filepath).
			Str("took", fmt.Sprintf("%s", time.Since(start))).
			Msg("Parsed input.")
	}(time.Now())

	tokens, err := Scan(filepath, input)
	if err != nil {
		return ASTNode{}, err
	}

	p := newParser(filepath, tokens, opts...)

	return p.parseAST()
}

// ParseString reads and parses a string of input.
func ParseString(input string) (ASTNode, error) {
	return Parse("", input)
}

// ParseFile reads a file and parses the input.
func ParseFile(filepath string) (ASTNode, error) {
	contents, err := ioutil.ReadFile(filepath)

	if errors.Is(err, os.ErrNotExist) {
		return ASTNode{}, help.UserFacingErrorf(
			`Failed to open %s. The file does not seem to exist.

Please check that you haven't misspelled the file name, etc.`,
			color.Aurora.BrightYellow(filepath),
		)
	}

	if err != nil {
		return ASTNode{}, err
	}

	return Parse(filepath, string(contents))
}

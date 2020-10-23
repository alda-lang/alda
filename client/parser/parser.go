package parser

import (
	"fmt"
	"io/ioutil"
	"time"

	log "alda.io/client/logging"
	model "alda.io/client/model"
)

type parser struct {
	filename string
	input    []Token
	updates  []model.ScoreUpdate
	current  int
}

func newParser(filename string, tokens []Token) *parser {
	return &parser{
		filename: filename,
		input:    tokens,
		updates:  []model.ScoreUpdate{},
		current:  0,
	}
}

func (p *parser) peek() Token {
	return p.input[p.current]
}

func (p *parser) previous() Token {
	return p.input[p.current-1]
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
	if p.peek().tokenType != EOF {
		p.current++
	}

	return p.previous()
}

func (p *parser) match(types ...TokenType) bool {
	for _, tokenType := range types {
		if p.check(tokenType) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *parser) addUpdate(update model.ScoreUpdate) {
	log.Debug().Str("update", fmt.Sprintf("%#v", update)).Msg("Adding update.")
	p.updates = append(p.updates, update)
}

type parseError struct {
	filename string
	token    Token
	msg      string
}

// Should e.token.tokenType be included too?
func (e *parseError) Error() string {
	return fmt.Sprintf(
		"%s:%d:%d %s",
		e.filename, e.token.line, e.token.column, e.msg,
	)
}

func (p *parser) errorAtToken(token Token, msg string) *parseError {
	return &parseError{
		filename: p.filename,
		token:    token,
		msg:      msg,
	}
}

func (p *parser) unexpectedTokenError(token Token, context string) *parseError {
	if context != "" {
		context = " " + context
	}

	msg := fmt.Sprintf(
		"Unexpected %s `%s`%s", token.tokenType.String(), token.text, context,
	)

	return p.errorAtToken(token, msg)
}

func (p *parser) consume(tokenType TokenType, context string) (Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}

	return Token{}, p.unexpectedTokenError(p.peek(), context)
}

func assertSingleUpdate(updates []model.ScoreUpdate) {
	if len(updates) != 1 {
		panic(fmt.Sprintf("Expected a single update in %#v", updates))
	}
}

func (p *parser) lispForm(context string) (model.LispForm, error) {
	switch {
	case p.match(Symbol):
		return model.LispSymbol{Name: p.previous().text}, nil
	case p.match(Number):
		return model.LispNumber{Value: p.previous().literal.(float64)}, nil
	case p.match(String):
		return model.LispString{Value: p.previous().literal.(string)}, nil
	case p.match(LeftParen):
		return p.lispList()
	default:
		return nil, p.unexpectedTokenError(p.peek(), context)
	}
}

func (p *parser) lispList() (model.LispList, error) {
	// NB: This assumes the initial LeftParen token was already consumed.
	list := model.LispList{}

	for token := p.peek(); token.tokenType != RightParen; token = p.peek() {
		if p.match(EOF) {
			return list, p.errorAtToken(token, "Unterminated S-expression.")
		}

		quoted := p.match(SingleQuote)

		form, err := p.lispForm("in S-expression")
		if err != nil {
			return list, err
		}

		if quoted {
			form = model.LispQuotedForm{Form: form}
		}

		list.Elements = append(list.Elements, form)
	}

	if _, err := p.consume(RightParen, "in S-expression"); err != nil {
		return list, err
	}

	return list, nil
}

func (p *parser) sexp() ([]model.ScoreUpdate, error) {
	// NB: This assumes the initial LeftParen token was already consumed.
	list, err := p.lispList()
	if err != nil {
		return nil, err
	}

	return []model.ScoreUpdate{p.singleOrRepeated(list)}, nil
}

func (p *parser) part() ([]model.ScoreUpdate, error) {
	// NB: This assumes the initial Name token was already consumed.
	partDecl := model.PartDeclaration{Names: []string{p.previous().text}}

	for p.match(Separator) {
		name, err := p.consume(Name, "in part declaration")
		if err != nil {
			return nil, err
		}

		partDecl.Names = append(partDecl.Names, name.text)
	}

	if p.match(Alias) {
		partDecl.Alias = p.previous().literal.(string)
	}

	if _, err := p.consume(Colon, "in part declaration"); err != nil {
		return nil, err
	}

	return []model.ScoreUpdate{partDecl}, nil
}

func (p *parser) variableDefinition() ([]model.ScoreUpdate, error) {
	// NB: This assumes the initial Name token was already consumed.
	definition := model.VariableDefinition{VariableName: p.previous().text}
	definitionLine := p.previous().line

	if _, err := p.consume(Equals, "in variable definition"); err != nil {
		return nil, err
	}

	if p.peek().line > definitionLine {
		return nil, fmt.Errorf(
			"There must be at least one event following the '=' on line %d",
			definitionLine,
		)
	}

	for t := p.peek(); t.line == definitionLine; t = p.peek() {
		if t.tokenType == EOF {
			break
		}

		events, err := p.topLevel()
		if err != nil {
			return nil, err
		}
		definition.Events = append(definition.Events, events...)
	}

	return []model.ScoreUpdate{definition}, nil
}

func (p *parser) singleOrRepeated(update model.ScoreUpdate) model.ScoreUpdate {
	if p.match(Repeat) {
		return model.Repeat{Event: update, Times: p.previous().literal.(int32)}
	}

	return update
}

func (p *parser) variableReference() ([]model.ScoreUpdate, error) {
	// NB: This assumes the initial Name token was already consumed.
	reference := model.VariableReference{VariableName: p.previous().text}
	return []model.ScoreUpdate{p.singleOrRepeated(reference)}, nil
}

func (p *parser) partOrVariableOp() ([]model.ScoreUpdate, error) {
	// NB: This assumes the initial Name token was already consumed.
	switch p.peek().tokenType {
	case Equals:
		return p.variableDefinition()
	case Alias, Separator, Colon:
		return p.part()
	default:
		return p.variableReference()
	}
}

func (p *parser) octaveSet() ([]model.ScoreUpdate, error) {
	// NB: This assumes the OctaveSet token was already consumed.
	return []model.ScoreUpdate{
		model.AttributeUpdate{
			PartUpdate: model.OctaveSet{OctaveNumber: p.previous().literal.(int32)},
		},
	}, nil
}

func (p *parser) matchDurationComponent() bool {
	return p.match(NoteLength, NoteLengthMs)
}

func (p *parser) durationComponent() model.DurationComponent {
	// NB: This assumes the duration component token was already consumed.
	token := p.previous()

	switch token.tokenType {
	case NoteLength:
		nl := token.literal.(noteLength)
		return model.NoteLength{
			Denominator: nl.denominator,
			Dots:        nl.dots,
		}
	case NoteLengthMs:
		return model.NoteLengthMs{Quantity: token.literal.(float64)}
	}

	// We shouldn't get here.
	return nil
}

func (p *parser) duration() model.Duration {
	// NB: This assumes the initial duration component token was already consumed.
	duration := model.Duration{
		Components: []model.DurationComponent{p.durationComponent()},
	}

	for {
		for p.match(Barline) {
			duration.Components = append(duration.Components, model.Barline{})
		}

		// Take note of the current position. If we encounter a tie that ends up
		// actually being a slur (i.e. it isn't followed by a note length), we can
		// backtrack to this position, return the duration, and let the parser
		// consume the slur as part of e.g. a note.
		beforeTies := p.current

		// We'll stash any barlines that we encounter here temporarily. We'll add
		// them to the duration components iff we aren't going to backtrack and
		// consume them outside of the duration.
		barlines := []model.DurationComponent{}

		if !p.match(Tie) {
			return duration
		}

		for p.match(Barline) {
			barlines = append(barlines, model.Barline{})
		}

		for p.match(Tie) {
			// In some cases, it makes sense to have extraneous ties, e.g. when you're
			// tying a duration across a barline and it feels right to have a tie on
			// either side of the barline. So we'll consume any additional ties here.
		}

		if !p.matchDurationComponent() {
			p.current = beforeTies
			return duration
		}

		duration.Components = append(duration.Components, barlines...)
		duration.Components = append(duration.Components, p.durationComponent())
	}
}

func (p *parser) note() (model.Note, error) {
	// NB: This assumes the initial NoteLetter token was already consumed.
	noteLetter, err := model.NewNoteLetter(p.previous().literal.(rune))
	if err != nil {
		return model.Note{}, err
	}

	pitch := model.LetterAndAccidentals{NoteLetter: noteLetter}

AccidentalsLoop:
	for {
		switch {
		case p.match(Flat):
			pitch.Accidentals = append(pitch.Accidentals, model.Flat)
		case p.match(Natural):
			pitch.Accidentals = append(pitch.Accidentals, model.Natural)
		case p.match(Sharp):
			pitch.Accidentals = append(pitch.Accidentals, model.Sharp)
		default:
			break AccidentalsLoop
		}
	}

	note := model.Note{Pitch: pitch}

	if p.matchDurationComponent() {
		note.Duration = p.duration()
	}

	if p.match(Tie) {
		note.Slurred = true
	}

	return note, nil
}

func (p *parser) rest() model.Rest {
	// NB: This assumes the initial RestLetter token was already consumed.

	rest := model.Rest{}

	if p.matchDurationComponent() {
		rest.Duration = p.duration()
	}

	return rest
}

func (p *parser) noteOrRest() (model.ScoreUpdate, error) {
	//NB: This assumes the initial NoteLetter/RestLetter was already consumed.
	switch letter := p.previous(); letter.tokenType {
	case NoteLetter:
		return p.note()
	case RestLetter:
		return p.rest(), nil
	default:
		return nil, p.unexpectedTokenError(letter, "in note/rest")
	}
}

func (p *parser) updatesBetweenNotesInChord() ([]model.ScoreUpdate, error) {
	updates := []model.ScoreUpdate{}

	for {
		switch {
		case p.match(OctaveUp):
			updates = append(updates, model.AttributeUpdate{
				PartUpdate: model.OctaveUp{},
			})
		case p.match(OctaveDown):
			updates = append(updates, model.AttributeUpdate{
				PartUpdate: model.OctaveDown{},
			})
		case p.match(OctaveSet):
			octaveSetUpdates, err := p.octaveSet()
			if err != nil {
				return nil, err
			}
			updates = append(updates, octaveSetUpdates...)
		case p.match(LeftParen):
			sexp, err := p.lispList()
			if err != nil {
				return nil, err
			}
			updates = append(updates, sexp)
		default:
			return updates, nil
		}
	}
}

// Parses a note or chord. A chord contains multiple chords and rests, not to
// mention attribute changes, so any of those will be parsed too in the process.
func (p *parser) noteRestOrChord() ([]model.ScoreUpdate, error) {
	// NB: This assumes the initial NoteLetter/RestLetter was already consumed.

	// The cumulative list of updates. Depending on whether this is a chord, the
	// updates will either be emitted as part of the chord, or emitted
	// individually.
	allUpdates := []model.ScoreUpdate{}

	var repeats int32

	for {
		noteOrRest, err := p.noteOrRest()
		if err != nil {
			return nil, err
		}

		if p.match(Repeat) {
			allUpdates = append(allUpdates, noteOrRest)
			repeats = p.previous().literal.(int32)
			break
		}

		// The updates for just this repetition of the loop
		updates := []model.ScoreUpdate{noteOrRest}

		updatesBeforeSeparator, err := p.updatesBetweenNotesInChord()
		if err != nil {
			return nil, err
		}
		updates = append(updates, updatesBeforeSeparator...)

		if !p.match(Separator) {
			allUpdates = append(allUpdates, updates...)
			break
		}

		updatesAfterSeparator, err := p.updatesBetweenNotesInChord()
		if err != nil {
			return nil, err
		}
		updates = append(updates, updatesAfterSeparator...)

		allUpdates = append(allUpdates, updates...)

		if !p.match(NoteLetter, RestLetter) {
			return nil, p.unexpectedTokenError(p.peek(), "in chord")
		}
	}

	notesCount := 0
	for _, update := range allUpdates {
		switch update.(type) {
		case model.Note, model.Rest:
			notesCount++
		}
	}

	if notesCount > 1 {
		allUpdates = []model.ScoreUpdate{model.Chord{Events: allUpdates}}
	}

	if repeats > 0 {
		assertSingleUpdate(allUpdates)
		return []model.ScoreUpdate{
			model.Repeat{Event: allUpdates[0], Times: int32(repeats)},
		}, nil
	}

	return allUpdates, nil
}

func (p *parser) repetitions() ([]model.RepetitionRange, error) {
	// NB: This assumes the Repetitions token was already consumed.
	token := p.previous()

	repetitions := []model.RepetitionRange{}

	for _, er := range token.literal.([]repetitionRange) {
		repetitionRange := model.RepetitionRange{First: er.first, Last: er.last}
		repetitions = append(repetitions, repetitionRange)
	}

	return repetitions, nil
}

func (p *parser) eventSeq() ([]model.ScoreUpdate, error) {
	// NB: This assumes the initial EventSeqOpen token was already consumed.
	allEvents := []model.ScoreUpdate{}

	for token := p.peek(); token.tokenType != EventSeqClose; token = p.peek() {
		if p.match(EOF) {
			return nil, p.errorAtToken(token, "Unterminated event sequence.")
		}

		events, err := p.topLevel()
		if err != nil {
			return nil, err
		}

		if p.match(Repetitions) {
			repetitions, err := p.repetitions()
			if err != nil {
				return nil, err
			}

			lastI := len(events) - 1
			events[lastI] = model.OnRepetitions{
				Repetitions: repetitions, Event: events[lastI],
			}
		}

		allEvents = append(allEvents, events...)
	}

	if _, err := p.consume(EventSeqClose, "in event sequence"); err != nil {
		return nil, err
	}

	eventSeq := model.EventSequence{Events: allEvents}

	return []model.ScoreUpdate{p.singleOrRepeated(eventSeq)}, nil
}

func (p *parser) cram() ([]model.ScoreUpdate, error) {
	// NB: This assumes the initial CramOpen token was already consumed.
	allEvents := []model.ScoreUpdate{}

	for token := p.peek(); token.tokenType != CramClose; token = p.peek() {
		if p.match(EOF) {
			return nil, p.errorAtToken(token, "Unterminated cram expression.")
		}

		events, err := p.topLevel()
		if err != nil {
			return nil, err
		}
		allEvents = append(allEvents, events...)
	}

	if _, err := p.consume(CramClose, "in cram expression"); err != nil {
		return nil, err
	}

	cram := model.Cram{Events: allEvents}

	if p.matchDurationComponent() {
		cram.Duration = p.duration()
	}

	return []model.ScoreUpdate{p.singleOrRepeated(cram)}, nil
}

func (p *parser) voiceMarker() ([]model.ScoreUpdate, error) {
	// NB: This assumes the VoiceMarker token was already consumed.
	voiceNumber := p.previous().literal.(int32)

	if voiceNumber == 0 {
		return []model.ScoreUpdate{model.VoiceGroupEndMarker{}}, nil
	}

	return []model.ScoreUpdate{model.VoiceMarker{VoiceNumber: voiceNumber}}, nil
}

func (p *parser) topLevel() ([]model.ScoreUpdate, error) {
	switch {
	case p.match(LeftParen):
		return p.sexp()
	case p.match(Name):
		return p.partOrVariableOp()
	case p.match(OctaveSet):
		return p.octaveSet()
	case p.match(OctaveUp):
		return []model.ScoreUpdate{
			model.AttributeUpdate{PartUpdate: model.OctaveUp{}},
		}, nil
	case p.match(OctaveDown):
		return []model.ScoreUpdate{
			model.AttributeUpdate{PartUpdate: model.OctaveDown{}},
		}, nil
	case p.match(NoteLetter, RestLetter):
		return p.noteRestOrChord()
	case p.match(Barline):
		return []model.ScoreUpdate{model.Barline{}}, nil
	case p.match(EventSeqOpen):
		return p.eventSeq()
	case p.match(CramOpen):
		return p.cram()
	case p.match(VoiceMarker):
		return p.voiceMarker()
	case p.match(Marker):
		return []model.ScoreUpdate{
			model.Marker{Name: p.previous().literal.(string)},
		}, nil
	case p.match(AtMarker):
		return []model.ScoreUpdate{
			model.AtMarker{Name: p.previous().literal.(string)},
		}, nil
	}

	return nil, p.unexpectedTokenError(p.peek(), "at the top level")
}

// Parse a string of input into a sequence of score updates.
func Parse(filepath string, input string) ([]model.ScoreUpdate, error) {
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
		return nil, err
	}

	p := newParser(filepath, tokens)

	for t := p.peek(); t.tokenType != EOF; t = p.peek() {
		// log.Debug().Str("token", t.String()).Msg("Parsing token.")

		updates, err := p.topLevel()
		if err != nil {
			return nil, err
		}

		for _, update := range updates {
			p.addUpdate(update)
		}
	}

	return p.updates, nil
}

// ParseString reads and parses a string of input.
func ParseString(input string) ([]model.ScoreUpdate, error) {
	return Parse("<no file>", input)
}

// ParseFile reads a file and parses the input.
func ParseFile(filepath string) ([]model.ScoreUpdate, error) {
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return Parse(filepath, string(contents))
}

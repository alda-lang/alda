package parser

import (
	log "alda.io/client/logging"
	"fmt"
	"io/ioutil"
	"strconv"
)

func isDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

func isLetter(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

type scanError struct {
	filename string
	line     int
	column   int
	msg      string
}

func (e *scanError) Error() string {
	return fmt.Sprintf("%s:%d:%d %s", e.filename, e.line, e.column, e.msg)
}

func (s *scanner) errorAtPosition(
	line int, column int, msg string) *scanError {
	return &scanError{
		filename: s.filename,
		line:     line,
		column:   column,
		msg:      msg,
	}
}

func (s *scanner) unexpectedCharError(
	c rune, context string, line int, column int) *scanError {
	if context != "" {
		context = " " + context
	}
	msg := fmt.Sprintf("Unexpected '%c'%s", c, context)
	return s.errorAtPosition(line, column, msg)
}

type TokenType int

const (
	AtMarker TokenType = iota
	Barline
	Colon
	CramClose
	CramOpen
	Endings
	EOF
	Equals
	EventSeqClose
	EventSeqOpen
	Flat
	Integer
	LeftParen
	Marker
	Name
	Natural
	Nickname
	NoteLength
	NoteLengthMs
	NoteLetter
	Number
	OctaveDown
	OctaveSet
	OctaveUp
	RestLetter
	RightParen
	Separator
	Sharp
	SingleQuote
	String
	Symbol
	Tie
	Repeat
	VoiceMarker
)

type Token struct {
	tokenType TokenType
	text      string
	literal   interface{}
	line      int
}

func (tt TokenType) ToString() string {
	switch tt {
	case AtMarker:
		return "AtMarker"
	case Barline:
		return "Barline"
	case Colon:
		return "Colon"
	case CramClose:
		return "CramClose"
	case CramOpen:
		return "CramOpen"
	case Endings:
		return "Endings"
	case EOF:
		return "EOF"
	case Equals:
		return "Equals"
	case EventSeqClose:
		return "EventSeqClose"
	case EventSeqOpen:
		return "EventSeqOpen"
	case Flat:
		return "Flat"
	case Integer:
		return "Integer"
	case LeftParen:
		return "LeftParen"
	case Marker:
		return "Marker"
	case Name:
		return "Name"
	case Natural:
		return "Natural"
	case Nickname:
		return "Nickname"
	case NoteLength:
		return "NoteLength"
	case NoteLengthMs:
		return "NoteLengthMs"
	case NoteLetter:
		return "NoteLetter"
	case Number:
		return "Number"
	case OctaveDown:
		return "OctaveDown"
	case OctaveSet:
		return "OctaveSet"
	case OctaveUp:
		return "OctaveUp"
	case Repeat:
		return "Repeat"
	case RestLetter:
		return "RestLetter"
	case RightParen:
		return "RightParen"
	case Separator:
		return "Separator"
	case Sharp:
		return "Sharp"
	case SingleQuote:
		return "SingleQuote"
	case String:
		return "String"
	case Symbol:
		return "Symbol"
	case Tie:
		return "Tie"
	case VoiceMarker:
		return "VoiceMarker"
	default:
		return fmt.Sprintf("%d (ToString not implemented)", tt)
	}
}

func (t Token) ToString() string {
	return fmt.Sprintf(
		"[line %d] %s | %#q | %#v", t.line, t.tokenType.ToString(), t.text, t.literal,
	)
}

type scanner struct {
	filename  string
	input     []rune
	tokens    []Token
	start     int
	current   int
	line      int
	column    int
	sexpLevel int
}

func newScanner(filename string, input string) *scanner {
	return &scanner{
		filename:  filename,
		input:     []rune(input),
		tokens:    []Token{},
		start:     0,
		current:   0,
		line:      1,
		column:    1,
		sexpLevel: 0,
	}
}

func (s *scanner) reachedEOF() bool {
	return s.current >= len(s.input)
}

func (s *scanner) eofIsNext() bool {
	return s.current+1 >= len(s.input)
}

func (s *scanner) peek() rune {
	if s.current >= len(s.input) {
		return 0
	}

	return s.input[s.current]
}

func (s *scanner) peekNext() rune {
	if s.current+1 >= len(s.input) {
		return 0
	}

	return s.input[s.current+1]
}

func (s *scanner) advance() rune {
	r := s.input[s.current]
	s.current++

	if r == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}

	return r
}

func (s *scanner) match(expected rune) bool {
	if s.current >= len(s.input) {
		return false
	}

	if s.input[s.current] != expected {
		return false
	}

	s.advance()

	return true
}

func (s *scanner) addToken(tokenType TokenType, literal interface{}) {
	text := string(s.input[s.start:s.current])

	token := Token{
		tokenType: tokenType,
		text:      text,
		literal:   literal,
		line:      s.line,
	}

	log.Debug().Str("token", token.ToString()).Msg("Adding token.")
	s.tokens = append(s.tokens, token)
}

func (s *scanner) skipComment() {
	for s.peek() != '\n' && !s.reachedEOF() {
		s.advance()
	}
}

func (s *scanner) parseString() error {
	// NB: This assumes the initial quote was already consumed.

	for s.peek() != '"' && !s.reachedEOF() {
		s.advance()
	}

	if s.reachedEOF() {
		return s.errorAtPosition(s.line, s.column, "Unterminated string")
	}

	// Consume the closing ".
	s.advance()

	// Trim the surrounding quotes.
	contents := s.input[s.start+1 : s.current-1]
	s.addToken(String, string(contents))

	return nil
}

func (s *scanner) consumeWhile(pred func(rune) bool) {
	for c := s.peek(); pred(c); c = s.peek() {
		s.advance()
	}
}

func (s *scanner) consumeDigits() {
	s.consumeWhile(isDigit)
}

func (s *scanner) consumeSpaces() {
	s.consumeWhile(func(c rune) bool { return c == ' ' })
}

// This function is meant to be called after consuming a bunch of digits. It
// reads from the start index up until the current index and parses the result
// as an integer.
func (s *scanner) parseIntegerFrom(startIndex int) int32 {
	integer, _ := strconv.ParseInt(string(s.input[startIndex:s.current]), 10, 32)
	return int32(integer)
}

// This function is meant to be called after consuming a bunch of digits, then
// optionally a period and a bunch more digits. It reads from the start index up
// until the current index and parses the result as a float.
func (s *scanner) parseFloatFrom(startIndex int) float32 {
	number, _ := strconv.ParseFloat(string(s.input[s.start:s.current]), 32)
	return float32(number)
}

type noteLength struct {
	denominator float32
	dots        int32
}

func terminatesNoteLength(c rune) bool {
	return c == ' ' || c == '\n' || c == '/' || c == '~'
}

func (s *scanner) parseNoteLength() {
	// NB: This assumes that the first digit has already been consumed.

	// Consume the rest of the digits.
	s.consumeDigits()

	// Look for a fractional part.
	if s.peek() == '.' && isDigit(s.peekNext()) {
		// Consume the decimal.
		s.advance()
		// Consume digits to the right of the decimal.
		s.consumeDigits()
	}

	number := s.parseFloatFrom(s.start)

	c := s.peek()
	n := s.peekNext()
	if c == 's' && (terminatesNoteLength(n) || s.eofIsNext()) {
		// consume 's'
		s.advance()
		s.addToken(NoteLengthMs, number*1000)
		return
	}

	if c == 'm' && n == 's' {
		// consume 'm' and 's'
		s.advance()
		s.advance()
		s.addToken(NoteLengthMs, number)
		return
	}

	dots := 0

	for c := s.peek(); c == '.'; c = s.peek() {
		dots++
		s.advance()
	}

	s.addToken(NoteLength, noteLength{denominator: number, dots: int32(dots)})
}

func (s *scanner) parseInteger() {
	s.consumeDigits()
	s.addToken(Integer, s.parseIntegerFrom(s.start))
}

func (s *scanner) parseNumber() {
	// Parse numbers before the period.
	s.consumeDigits()

	// Look for a fractional part.
	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance()
	}

	// Parse numbers after the period.
	s.consumeDigits()

	s.addToken(Number, s.parseFloatFrom(s.start))
}

type endingRange struct {
	first int32
	last  int32
}

func (s *scanner) parseEndings() error {
	// NB: This assumes the initial "'" was already consumed.

	ranges := []endingRange{}

	// Parse ending ranges as long as we continue to encounter them.
	for {
		if c := s.peek(); !isDigit(c) {
			return s.unexpectedCharError(c, "in endings", s.line, s.column)
		}

		// Parse the "first" number of the range.
		startNumber := s.current
		s.consumeDigits()
		first := s.parseIntegerFrom(startNumber)
		er := endingRange{first: first}

		// Either parse the "last" number of the range, or make the first number
		// the last number as well, indicating a range of one number, e.g. 3-3.
		if s.match('-') {
			// Make sure a number comes next.
			if c := s.peek(); !isDigit(c) {
				return s.unexpectedCharError(c, "in endings", s.line, s.column)
			}

			startNumber := s.current
			s.consumeDigits()
			er.last = s.parseIntegerFrom(startNumber)
		} else {
			er.last = first
		}

		ranges = append(ranges, er)

		// At this point, there could be a comma, indicating there are more ranges
		// to parse. Only in that case do we continue to loop.
		if !s.match(',') {
			break
		}
	}

	s.addToken(Endings, ranges)
	return nil
}

func (s *scanner) parseRepeat() error {
	// NB: This assumes the initial '*' was already consumed.

	s.consumeSpaces()

	if c := s.peek(); !isDigit(c) {
		return s.unexpectedCharError(c, "in repeat", s.line, s.column)
	}

	startDigits := s.current
	s.consumeDigits()

	if c := s.peek(); c != ' ' && c != '\n' && !s.reachedEOF() {
		return s.unexpectedCharError(c, "in repeat", s.line, s.column)
	}

	digits := s.input[startDigits:s.current]
	times, _ := strconv.ParseInt(string(digits), 10, 32)
	s.addToken(Repeat, int32(times))

	return nil
}

func (s *scanner) parseOctaveSet() error {
	// NB: This assumes the initial 'o' was already consumed.

	s.consumeDigits()

	if c := s.peek(); c != ' ' && c != '\n' && !s.reachedEOF() {
		return s.unexpectedCharError(c, "in octave set", s.line, s.column)
	}

	// Trim the initial 'o'
	digits := s.input[s.start+1 : s.current]
	octaveNumber, _ := strconv.ParseInt(string(digits), 10, 32)
	s.addToken(OctaveSet, int32(octaveNumber))

	return nil
}

func (s *scanner) parseVoiceMarker() error {
	// NB: This assumes the initial 'V' was already consumed.

	s.consumeDigits()

	if c := s.peek(); c != ':' {
		return s.unexpectedCharError(c, "in voice marker", s.line, s.column)
	}

	// Consume the final ':'
	s.advance()

	// Trim the surrounding 'V' and ':'.
	digits := s.input[s.start+1 : s.current-1]
	voiceNumber, _ := strconv.ParseInt(string(digits), 10, 32)
	s.addToken(VoiceMarker, int32(voiceNumber))

	return nil
}

func isValidSymbolChar(c rune) bool {
	switch c {
	case ' ', '\t', '\r', '\n', '(', ')', '[', ']', '{', '}':
		return false
	}

	return true
}

func (s *scanner) parseSymbol() {
	s.consumeWhile(isValidSymbolChar)
	s.addToken(Symbol, nil)
}

func isValidNameChar(c rune) bool {
	if isLetter(c) || isDigit(c) {
		return true
	}

	switch c {
	case '_', '-', '+', '\'', '(', ')', '.':
		return true
	}

	return false
}

func (s *scanner) parseName() {
	s.consumeWhile(isValidNameChar)
	s.addToken(Name, nil)
}

func (s *scanner) parseNickname() error {
	// NB: This assumes the initial double quote was already consumed.

	s.consumeWhile(isValidNameChar)

	if s.reachedEOF() {
		return s.errorAtPosition(s.line, s.column, "Unterminated nickname")
	}

	if c := s.peek(); c != '"' {
		s.unexpectedCharError(c, "in nickname", s.line, s.column)
	}

	// Consume the closing ".
	s.advance()

	// Trim the surrounding quotes.
	contents := s.input[s.start+1 : s.current-1]
	s.addToken(Nickname, string(contents))

	return nil
}

func (s *scanner) parsePrefixedName(
	tokenType TokenType, contextName string) error {
	// NB: This assumes the initial prefix character was already consumed.

	if c := s.peek(); !isValidNameChar(c) {
		return s.unexpectedCharError(c, contextName, s.line, s.column)
	}

	s.consumeWhile(isValidNameChar)

	// Trim the initial prefix character
	s.addToken(tokenType, string(s.input[s.start+1:s.current]))

	return nil
}

func (s *scanner) parseMarker() error {
	return s.parsePrefixedName(Marker, "in marker name")
}

func (s *scanner) parseAtMarker() error {
	return s.parsePrefixedName(AtMarker, "in marker name")
}

func isNoteLetter(c rune) bool {
	return 'a' <= c && c <= 'g'
}

func isRestLetter(c rune) bool {
	return c == 'r'
}

func isOctaveLetter(c rune) bool {
	return c == 'o'
}

func isVoiceLetter(c rune) bool {
	return c == 'V'
}

func followsNoteLetter(c rune) bool {
	if isDigit(c) {
		return true
	}

	switch c {
	case '#', ' ', '\n', '+', '-', '_', '/', '~', '*', '\'', '}', ']', '<', '>':
		return true
	}

	return false
}

func followsRestLetter(c rune) bool {
	if isDigit(c) {
		return true
	}

	switch c {
	case '#', ' ', '\n', '/', '~', '*', '\'', '}', ']', '<', '>':
		return true
	}

	return false
}

func (s *scanner) scanToken() error {
	prevLine := s.line
	prevColumn := s.column

	c := s.advance()

	switch c {
	case ' ', '\r', '\n', '\t':
		// skip whitespace
		return nil
	case '#':
		s.skipComment()
		return nil
	case '(':
		s.sexpLevel++
		s.addToken(LeftParen, nil)
		return nil
	case ')':
		s.sexpLevel--
		s.addToken(RightParen, nil)
		return nil
	}

	// If we're inside of parentheses, we are in "lisp mode," which has different
	// syntax rules.
	if s.sexpLevel > 0 {
		var err error

		switch c {
		case '\'':
			s.addToken(SingleQuote, nil)
		case '"':
			err = s.parseString()
		default:
			switch {
			case isDigit(c):
				s.parseNumber()
			case isValidSymbolChar(c):
				s.parseSymbol()
			default:
				return s.unexpectedCharError(c, "in S-expression", prevLine, prevColumn)
			}
		}
		return err
	}

	var err error

	switch c {
	case '{':
		s.addToken(CramOpen, nil)
	case '}':
		s.addToken(CramClose, nil)
	case '[':
		s.addToken(EventSeqOpen, nil)
	case ']':
		s.addToken(EventSeqClose, nil)
	case '-':
		s.addToken(Flat, nil)
	case '+':
		s.addToken(Sharp, nil)
	case '_':
		s.addToken(Natural, nil)
	case '/':
		s.addToken(Separator, nil)
	case '<':
		s.addToken(OctaveDown, nil)
	case '>':
		s.addToken(OctaveUp, nil)
	case ':':
		s.addToken(Colon, nil)
	case '~':
		s.addToken(Tie, nil)
	case '|':
		s.addToken(Barline, nil)
	case '=':
		s.addToken(Equals, nil)
	case '\'':
		err = s.parseEndings()
	case '*':
		err = s.parseRepeat()
	case '"':
		err = s.parseNickname()
	case '%':
		err = s.parseMarker()
	case '@':
		err = s.parseAtMarker()
	default:
		switch {
		case isDigit(c):
			s.parseNoteLength()
		case isLetter(c):
			n := s.peek()
			switch {
			case isLetter(n):
				s.parseName()
			case isNoteLetter(c) && (followsNoteLetter(n) || s.reachedEOF()):
				s.addToken(NoteLetter, c)
			case isRestLetter(c) && (followsRestLetter(n) || s.reachedEOF()):
				s.addToken(RestLetter, nil)
			case isOctaveLetter(c) && (isDigit(n) || s.reachedEOF()):
				err = s.parseOctaveSet()
			case isVoiceLetter(c) && isDigit(n):
				err = s.parseVoiceMarker()
			default:
				return s.unexpectedCharError(n, "in note/rest/name", s.line, s.column)
			}
		default:
			return s.unexpectedCharError(c, "at the top level", prevLine, prevColumn)
		}
	}

	return err
}

// Scan an input string and return a list of tokens.
//
// The `filename` argument is included in the error message in the event of a
// parse error.
func Scan(filename string, input string) ([]Token, error) {
	s := newScanner(filename, input)
	for !s.reachedEOF() {
		// We are at the beginning of the next lexeme.
		s.start = s.current
		// log.Debug().
		// 	Int("line", s.line).
		// 	Int("column", s.column).
		// 	Int("sexpLevel", s.sexpLevel).
		// 	Str("atCharacter", string(s.peek())).
		// 	Msg("Scanning token.")
		// Scan the next token.
		if err := s.scanToken(); err != nil {
			return nil, err
		}
	}

	s.tokens = append(s.tokens, Token{
		tokenType: EOF,
		text:      "",
		literal:   nil,
		line:      s.line,
	})

	return s.tokens, nil
}

// ScanFile reads a file, scans it, and returns a list of tokens.
func ScanFile(filepath string) ([]Token, error) {
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	return Scan(filepath, string(contents))
}

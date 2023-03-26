package parser

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type varDefState int

const (
	None     = iota // not currently in a var def
	Defining        // currently formatting a var def (cannot go to new line)
	LastNode        // last node of a var def, treated specially
)

type formatter struct {
	softWrapLen int         // configured line length to soft wrap formatting
	indentText  string      // configured indent string (i.e. spaces vs tabs)
	varDef      varDefState // state to handle formatting variable definitions
	indentLevel int         // state for indentation level
	texts       []string    // buffer of "tokens" for the ongoing formatted line
	out         io.Writer
}

type formatterOption func(*formatter)

func newFormatter(out io.Writer, opts ...formatterOption) *formatter {
	formatter := &formatter{
		// TODO: make these configurable from cli
		softWrapLen: 80,
		indentText:  "    ",
		varDef:      None,
		indentLevel: 0,
		texts:       []string{},
		out:         out,
	}

	for _, opt := range opts {
		opt(formatter)
	}

	return formatter
}

// line constructs and returns the current line being formatted.
func (f *formatter) line() string {
	text := strings.Join(f.texts, " ")
	if len(text) == 0 {
		return text
	} else {
		indent := strings.Repeat(f.indentText, f.indentLevel)
		return indent + text
	}
}

// emptyLine writes an empty line
func (f *formatter) emptyLine() {
	if f.varDef == None {
		f.flush()
		f.out.Write([]byte("\n"))
	}
}

// flush flushes out the current line to the output.
func (f *formatter) flush() {
	if len(f.texts) > 0 && f.varDef == None {
		f.out.Write([]byte(f.line() + "\n"))
		f.texts = []string{}
	}
}

// indent increments the indentation level of subsequent formatting.
func (f *formatter) indent() {
	switch f.varDef {
	case LastNode:
		// During the last node of a var def, we are allowed to effectively exit
		// the var def and continue formatting on the next line.
		f.varDef = None
		fallthrough
	case None:
		f.flush()
		f.indentLevel++
	}
}

// unindent decrements the indentation level of subsequent formatting.
// A corresponding unindent should always be called after calling indent.
func (f *formatter) unindent() {
	if f.varDef == None {
		f.flush()
		f.indentLevel--
	}
}

// write formats text to the output with indentation, wrapping, and spacing.
// Each "text" is an unwrappable token, i.e. wrapping only happens between text.
func (f *formatter) write(text string) {
	f.texts = append(f.texts, text)
	if len(f.line()) > f.softWrapLen && f.varDef == None {
		f.texts = f.texts[0 : len(f.texts)-1]
		f.flush()
		f.texts = append(f.texts, text)
	}
}

// formatWithDuration handles duration formatting.
// Durations are formatted with possible text directly pre/post (no spaces),
// i.e. note pitches preceding durations.
// All durations are treated as a single unwrappable text with the exception of
// barlines which cause a duration to be split into separate texts.
func (f *formatter) formatWithDuration(
	pre string, duration ASTNode, post string,
) error {
	text := strings.Builder{}
	text.WriteString(pre)
	shouldTie := false

	for i, child := range duration.Children {
		switch child.Type {

		default:
			return fmt.Errorf(
				"unexpected DurationNode %#v during formatting", child,
			)

		case BarlineNode:
			if i == len(duration.Children)-1 {
				// The final duration is a barline
				// We write out any post text before the barline for clarity
				text.WriteString(post)
			}

			// Barlines in a duration split formatting into separate texts
			if text.Len() > 0 {
				f.write(text.String())
			}
			f.write("|")

			text.Reset()

		case NoteLengthMsNode:
			if shouldTie {
				text.WriteString("~")
			}

			// NoteLengthMs is parsed to us as ms, but can be coded using s/ms
			// We will try our best to use seconds first
			totalMs := child.Literal.(float64)
			s := int(totalMs) / 1000
			ms := totalMs - float64(s*1000)

			if s > 0 && ms > 0 {
				text.WriteString(fmt.Sprintf(
					"%ds~%sms",
					s,
					strconv.FormatFloat(ms, 'f', -1, 64),
				))
			} else if s > 0 {
				text.WriteString(fmt.Sprintf("%ds", s))
			} else {
				text.WriteString(fmt.Sprintf(
					"%sms",
					strconv.FormatFloat(ms, 'f', -1, 64),
				))
			}

			shouldTie = true

		case NoteLengthNode:
			if shouldTie {
				text.WriteString("~")
			}

			if err := child.expectNChildren(1, 2); err != nil {
				return err
			}

			denom, err := child.Children[0].expectNodeType(DenominatorNode)
			if err != nil {
				return err
			}

			numDots := 0
			if len(child.Children) > 1 {
				dotsNode, err := child.Children[1].expectNodeType(DotsNode)
				if err != nil {
					return err
				}

				numDots = int(dotsNode.Literal.(int32))
			}

			text.WriteString(fmt.Sprintf(
				"%s%s",
				strconv.FormatFloat(denom.Literal.(float64), 'f', -1, 64),
				strings.Repeat(".", numDots),
			))

			shouldTie = true

		}
	}

	if text.Len() > 0 {
		text.WriteString(post)
		f.write(text.String())
	}

	return nil
}

// formatInnerEvents handles formatting of inner events within parts.
func (f *formatter) formatInnerEvents(nodes ...ASTNode) error {
	for _, node := range nodes {
		switch node.Type {

		default:
			return fmt.Errorf(
				"unexpected ASTNode Type %#v during formatting", node.Type,
			)

		case AtMarkerNode:
			f.write(fmt.Sprintf("@%s", node.Literal.(string)))

		case BarlineNode:
			f.write("|")

		case ChordNode:
			// We make each note + each separator individual texts to formatInnerEvents
			// Meaning extra spaces padding separators + chords can be wrapped
			// This is to avoid additional complexity in the formatter
			// We can change this by creating a new helper function for
			// inner-chord nodes that returns a []string of texts
			// Would have to handle the fact that barlines make multiple writes

			if err := node.expectChildren(); err != nil {
				return err
			}

			// Within a chord, there can be additional nodes between notes
			// We formatInnerEvents all of these after the separator for readability as
			// they apply to the subsequent note
			lastNoteOrRest := 0
			for i, child := range node.Children {
				if child.Type == NoteNode || child.Type == RestNode {
					lastNoteOrRest = i
				}
			}

			for i, child := range node.Children {
				err := f.formatInnerEvents(child)
				if err != nil {
					return err
				}

				if child.Type == NoteNode || child.Type == RestNode {
					if i < lastNoteOrRest {
						f.write("/")
					}
				}
			}

		case CramNode:
			if err := node.expectNChildren(1, 2); err != nil {
				return err
			}

			events, err := node.Children[0].expectNodeType(EventSequenceNode)
			if err != nil {
				return err
			}

			f.write("{")

			err = f.formatInnerEvents(events.Children...)
			if err != nil {
				return err
			}

			if len(node.Children) > 1 {
				duration, err := node.Children[1].expectNodeType(DurationNode)
				if err != nil {
					return err
				}

				err = f.formatWithDuration("}", duration, "")
				if err != nil {
					return err
				}
			} else {
				f.write("}")
			}

		case EventSequenceNode:
			// Always try to indent the children of standalone event sequences
			// (i.e. those not used as part of a separate node such as cram)
			f.flush()
			f.write("[")
			f.indent()

			err := f.formatInnerEvents(node.Children...)
			if err != nil {
				return err
			}

			f.unindent()
			f.write("]")

		case LispListNode:
			var lispString func(ASTNode) (string, error)
			lispString = func(lisp ASTNode) (string, error) {
				switch lisp.Type {

				default:
					return "", fmt.Errorf(
						"unexpected LispLispNode %#v during formatting", lisp,
					)

				case LispListNode:
					texts := []string{}

					for _, child := range lisp.Children {
						text, err := lispString(child)
						if err != nil {
							return "", err
						}

						texts = append(texts, text)
					}

					return fmt.Sprintf("(%s)", strings.Join(texts, " ")), nil

				case LispNumberNode:
					return strconv.FormatFloat(
						lisp.Literal.(float64), 'f', -1, 64,
					), nil

				case LispQuotedFormNode:
					form, err := lispString(lisp.Children[0])
					if err != nil {
						return "", err
					}

					return fmt.Sprintf("'%s", form), nil

				case LispStringNode:
					return fmt.Sprintf("\"%s\"", lisp.Literal.(string)), nil

				case LispSymbolNode:
					return lisp.Literal.(string), nil

				}
			}

			text, err := lispString(node)
			if err != nil {
				return err
			}

			// Lisp lists are generally short
			// We write them as a single unwrappable text for readability
			f.write(text)

		case MarkerNode:
			f.write(fmt.Sprintf("%%%s", node.Literal.(string)))

		case NoteNode:
			if err := node.expectNChildren(1, 2, 3); err != nil {
				return err
			}

			laa, err := node.Children[0].expectNodeType(
				NoteLetterAndAccidentalsNode,
			)
			if err != nil {
				return err
			}

			if err := laa.expectChildren(); err != nil {
				return err
			}

			letter, err := laa.Children[0].expectNodeType(NoteLetterNode)
			if err != nil {
				return err
			}

			pitchText := strings.Builder{}
			pitchText.WriteRune(letter.Literal.(rune))

			if len(laa.Children) > 1 {
				accidentals, err := laa.Children[1].expectNodeType(
					NoteAccidentalsNode,
				)
				if err != nil {
					return err
				}

				for _, child := range accidentals.Children {
					switch child.Type {
					default:
						return fmt.Errorf(
							"unexpected NoteAccidentalsNode %#v during formatting",
							child,
						)
					case FlatNode:
						pitchText.WriteString("-")
					case NaturalNode:
						pitchText.WriteString("_")
					case SharpNode:
						pitchText.WriteString("+")
					}
				}
			}

			slurText := ""
			for _, child := range node.Children[1:] {
				if child.Type == TieNode {
					slurText = "~"
				}
			}

			if len(node.Children) > 1 && node.Children[1].Type == DurationNode {
				err = f.formatWithDuration(
					pitchText.String(), node.Children[1], slurText,
				)
				if err != nil {
					return err
				}
			} else {
				f.write(fmt.Sprintf("%s%s", pitchText.String(), slurText))
			}

		case OctaveDownNode:
			f.write("<")

		case OctaveSetNode:
			f.write(fmt.Sprintf("o%d", node.Literal.(int32)))

		case OctaveUpNode:
			f.write(">")

		case RepeatNode:
			if err := node.expectNChildren(2); err != nil {
				return err
			}

			err := f.formatInnerEvents(node.Children[0])
			if err != nil {
				return err
			}

			times, err := node.Children[1].expectNodeType(TimesNode)
			if err != nil {
				return err
			}

			f.write(fmt.Sprintf("*%d", times.Literal.(int32)))

		case OnRepetitionsNode:
			if err := node.expectNChildren(2); err != nil {
				return err
			}

			err := f.formatInnerEvents(node.Children[0])
			if err != nil {
				return err
			}

			repetitions, err := node.Children[1].expectNodeType(RepetitionsNode)
			if err != nil {
				return err
			}

			ranges := []string{}
			for _, child := range repetitions.Children {
				rr, err := child.expectNodeType(RepetitionRangeNode)
				if err != nil {
					return err
				}

				if err := rr.expectNChildren(2); err != nil {
					return err
				}

				fr, err := rr.Children[0].expectNodeType(FirstRepetitionNode)
				if err != nil {
					return err
				}

				lr, err := rr.Children[1].expectNodeType(LastRepetitionNode)
				if err != nil {
					return err
				}

				frNum := fr.Literal.(int32)
				lrNum := lr.Literal.(int32)

				if frNum == lrNum {
					ranges = append(ranges,
						fmt.Sprintf("%d", frNum),
					)
				} else {
					ranges = append(ranges,
						fmt.Sprintf("%d-%d", frNum, lrNum),
					)
				}
			}
			f.write(fmt.Sprintf("'%s", strings.Join(ranges, ",")))

		case RestNode:
			if len(node.Children) > 0 {
				duration, err := node.Children[0].expectNodeType(DurationNode)
				if err != nil {
					return err
				}

				err = f.formatWithDuration("r", duration, "")
				if err != nil {
					return err
				}
			} else {
				f.write("r")
			}

		case VariableDefinitionNode:
			// Variable definitions are incredibly tricky to formatInnerEvents because
			// formatted text must be on the same line as the variable name.
			// We handle this by maintaining varDefState:
			// - While "None", behaviour is normal.
			// - While "Defining", we never flush/wrap to a new line.
			// - While "LastNode", we allow event sequences (including repeats)
			// 	 and voice groups to indent and continue on new lines.
			// 	 This is complicated, but any alternative I tried was worse.

			f.flush()
			f.varDef = Defining

			if err := node.expectNChildren(2); err != nil {
				return err
			}

			name, err := node.Children[0].expectNodeType(VariableNameNode)
			if err != nil {
				return err
			}

			f.write(fmt.Sprintf("%s =", name.Literal.(string)))

			events, err := node.Children[1].expectNodeType(EventSequenceNode)
			if err != nil {
				return err
			}

			if len(events.Children) > 0 {
				lastIndex := len(events.Children) - 1

				err = f.formatInnerEvents(events.Children[:lastIndex]...)
				if err != nil {
					return err
				}

				f.varDef = LastNode

				err = f.formatInnerEvents(events.Children[lastIndex])
				if err != nil {
					return err
				}
			}

			f.varDef = None
			f.flush()

		case VariableReferenceNode:
			f.write(node.Literal.(string))

		case VoiceGroupEndMarkerNode:
			f.write("V0:")

		case VoiceGroupNode:
			f.flush()
			err := f.formatInnerEvents(node.Children...)
			if err != nil {
				return err
			}

		case VoiceNode:
			if err := node.expectNChildren(2); err != nil {
				return err
			}

			voiceNumber, err := node.Children[0].expectNodeType(VoiceNumberNode)
			if err != nil {
				return err
			}

			f.write(fmt.Sprintf("V%d:", voiceNumber.Literal.(int32)))

			f.indent()

			events, err := node.Children[1].expectNodeType(EventSequenceNode)
			if err != nil {
				return err
			}

			err = f.formatInnerEvents(events.Children...)
			if err != nil {
				return err
			}

			f.unindent()

		}
	}

	return nil
}

// formatTopLevel handles formatting for the RootNode and parts.
func (f *formatter) formatTopLevel(root ASTNode) error {
	for i, part := range root.Children {
		switch part.Type {

		case ImplicitPartNode:
			if err := part.expectNChildren(1); err != nil {
				return err
			}

			events, err := part.Children[0].expectNodeType(EventSequenceNode)
			if err != nil {
				return err
			}

			err = f.formatInnerEvents(events.Children...)
			if err != nil {
				return err
			}

		case PartNode:
			if err := part.expectNChildren(2); err != nil {
				return err
			}

			// Part declaration
			decl, err := part.Children[0].expectNodeType(PartDeclarationNode)
			if err != nil {
				return err
			}

			if err := decl.expectNChildren(1, 2); err != nil {
				return err
			}

			partNames, err := decl.Children[0].expectNodeType(PartNamesNode)
			if err != nil {
				return err
			}

			if err := partNames.expectChildren(); err != nil {
				return err
			}

			names := []string{}
			for _, child := range partNames.Children {
				partNameNode, err := child.expectNodeType(PartNameNode)
				if err != nil {
					return err
				}

				names = append(names, partNameNode.Literal.(string))
			}
			namesText := strings.Join(names, "/")

			if len(decl.Children) > 1 {
				partAlias, err := decl.Children[1].expectNodeType(
					PartAliasNode,
				)
				if err != nil {
					return err
				}

				f.write(fmt.Sprintf(
					"%s \"%s\":",
					namesText,
					partAlias.Literal.(string),
				))
			} else {
				f.write(fmt.Sprintf(
					"%s:",
					namesText,
				))
			}

			// Part events
			f.indent()

			events, err := part.Children[1].expectNodeType(EventSequenceNode)
			if err != nil {
				return err
			}

			err = f.formatInnerEvents(events.Children...)
			if err != nil {
				return err
			}

			f.unindent()

		}

		f.flush()
		if i+1 < len(root.Children) {
			f.emptyLine()
		}
	}

	return nil
}

// FormatASTToCode performs rudimentary output formatting of Alda code including
// handling basic spacing, indentation, and line wrapping.
func FormatASTToCode(
	root ASTNode, out io.Writer, opts ...formatterOption,
) error {
	// Write to temp buffer instead of directly to file in case of error
	temp := bytes.Buffer{}
	f := newFormatter(&temp, opts...)
	err := f.formatTopLevel(root)
	if err != nil {
		return err
	}
	_, err = out.Write(temp.Bytes())
	return err
}

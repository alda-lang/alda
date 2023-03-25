package parser

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type formatter struct {
	softWrap int	// soft wrap lines
	shouldWrap bool
	indentLevel int	// indent level
	current []string
	out io.Writer
}

type formatterOption func(*formatter)

func newFormatter(out io.Writer, opts ...formatterOption) *formatter {
	formatter := &formatter{
		softWrap: 120,
		shouldWrap: true,
		indentLevel: 0,
		current: []string{},
		out: out,
	}

	for _, opt := range opts {
		opt(formatter)
	}

	return formatter
}

func (f *formatter) currentLine() string {
	indent := strings.Repeat("    ", f.indentLevel)
	text := strings.Join(f.current, " ")
	return strings.TrimSpace(indent + text)
}

func (f *formatter) flush() {
	if len(f.current) > 0 {
		f.out.Write([]byte(f.currentLine() + "\n"))
		f.current = []string{}
	}
}

func (f *formatter) emptyLine() {
	f.flush()
	f.out.Write([]byte("\n"))
}

func (f *formatter) indent() {
	f.flush()
	f.indentLevel++
}

func (f *formatter) unindent() {
	f.flush()
	f.indentLevel--
}

// We write an "atomic" / "unwrappable" text
func (f *formatter) write(text string) {
	f.current = append(f.current, text)
	if f.shouldWrap && len(f.currentLine()) > f.softWrap {
		f.current = f.current[0:len(f.current) - 1]
		f.flush()
		f.current = append(f.current, text)
	}
}

// formatWithDuration handles durations, which we format directly after text for readability
// We treat all duration as a single text (unwrappable) for readability
// 	EXCEPTION: barlines! then we split
// Ah make this a separate function that takes in the initial string
// That way we can split write separate texts by barlines
func (f *formatter) formatWithDuration(pre string, duration ASTNode, post string) error {
	text := strings.Builder{}
	text.WriteString(pre)
	shouldTie := false

	for i, child := range duration.Children {
		switch child.Type {

		default:
			return fmt.Errorf("unknown duration type") // TODO

		case BarlineNode:
			if i == len(duration.Children) - 1 {
				// The final duration is a barline
				// We will write out any post before the barline for clarity
				text.WriteString(post)
			}
			// We split separate wrappable if there's a barline
			if text.Len() > 0 {
				f.write(text.String())
			}
			f.write("|")
			text.Reset()
			shouldTie = false

		case NoteLengthMsNode:
			if shouldTie {
				text.WriteString("~")	// TODO: do this after any barlines - this is right
			}
			text.WriteString(child.Literal.(string))
			shouldTie = true

		case NoteLengthNode:
			if shouldTie {
				text.WriteString("~")
			}

			if err := child.expectChildren(); err != nil {
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

				numDots = dotsNode.Literal.(int)
			}

			text.WriteString(fmt.Sprintf("%f%s",
				denom.Literal.(float64),
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

// format formats individual notes
func (f *formatter) format(nodes ...ASTNode) error {
	for _, node := range nodes {
		switch node.Type {

		default:
			return fmt.Errorf("unexpected ASTNode type")

		case BarlineNode:
			f.write("|")

		case OctaveSetNode:
			f.write(fmt.Sprintf("o%d", node.Literal.(int32)))

		case OctaveUpNode:
			f.write(">")

		case OctaveDownNode:
			f.write("<")

		case AtMarkerNode:
			f.write(fmt.Sprintf("@%s", node.Literal.(string)))

		case MarkerNode:
			f.write(fmt.Sprintf("%%%s", node.Literal.(string)))

		case VariableReferenceNode:
			f.write(node.Literal.(string))

		// We consider entire notes and rests (+ their durations) as atomic
		// This is for purpose of readability
		// TODO: Barlines, however, can separate a note into two+ atomic texts
		case NoteNode:
			if err := node.expectNChildren(1, 2, 3); err != nil {
				return err
			}

			laa, err := node.Children[0].expectNodeType(NoteLetterAndAccidentalsNode)
			if err != nil {
				return err
			}

			if err := laa.expectChildren(); err != nil {
				return err
			}

			letter, err := laa.Children[0].expectNodeType(NoteLetterNode)

			pitchText := strings.Builder{}
			pitchText.WriteRune(letter.Literal.(rune))

			if len(letter.Children) > 1 {
				accidentals, err := laa.Children[1].expectNodeType(NoteAccidentalsNode)
				if err != nil {
					return err
				}

				for _, child := range accidentals.Children {
					switch child.Type {
					default:
						return fmt.Errorf("unknown accidentals type")
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
			if len(node.Children) > 2 {
				_, err := node.Children[2].expectNodeType(TieNode)
				if err != nil {
					return err
				}
				slurText = "~"
			}

			if len(node.Children) > 1 {
				duration, err := node.Children[1].expectNodeType(DurationNode)
				if err != nil {
					return err
				}

				err = f.formatWithDuration(pitchText.String(), duration, slurText)
				if err != nil {
					return err
				}
			} else {
				f.write(fmt.Sprintf("%s%s", pitchText.String(), slurText))
			}

		case RestNode:
			if len(node.Children) > 0 {
				durationNode, err := node.Children[0].expectNodeType(DurationNode)
				if err != nil {
					return err
				}

				err = f.formatWithDuration("r", durationNode, "")
				if err != nil {
					return err
				}

			}

		// For chords, we make separators pad by spaces
		// TODO: We could change this to not
		// 	Code would have to have separate helper function for inner-chord nodes
		// 	This function would not format directly, instead return []string
		// 	Problem here is notes in chords can have barlines...
		// 	This gets too complicated, better to leave it
		case ChordNode:
			if err := node.expectChildren(); err != nil {
				return err
			}

			// We format ensuring all note modifiers happen after separator
			// i.e. prefer c / >d over c> / d for clarity
			lastIndexWithDuration := 0
			for i, child := range node.Children {
				if child.Type == NoteNode || child.Type == RestNode {
					lastIndexWithDuration = i
				}
			}

			for i, child := range node.Children {
				err := f.format(child)
				if err != nil {
					return err
				}

				if i < lastIndexWithDuration && (child.Type == NoteNode || child.Type == RestNode) {
					// We inject separators "/" after non-last notes/rests
					f.write("/")
				}
			}

		// For Lisp, with the current state of Alda
		// Realistically it's very hard to get very long texts
		// So we keep it atomic for readability
		case LispListNode:
			var lispToString func(ASTNode) (string, error)
			lispToString = func(lisp ASTNode) (string, error) {
				switch lisp.Type {

				default:
					return "", fmt.Errorf("unexpected lisp type")

				case LispListNode:
					forms := []string{}
					for _, child := range lisp.Children {
						form, err := lispToString(child)
						if err != nil {
							return "", err
						}
						forms = append(forms, form)
					}
					return fmt.Sprintf("(%s)", strings.Join(forms, " ")), nil

				case LispNumberNode:
					return fmt.Sprintf("%d", lisp.Literal.(int32)), nil

				case LispQuotedFormNode:
					form, err := lispToString(lisp.Children[0])
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

			text, err := lispToString(node)
			if err != nil {
				return err
			}

			f.write(text)

		case VariableDefinitionNode:
			f.flush()

			if err := node.expectNChildren(2); err != nil {
				return err
			}

			variableName, err := node.Children[0].expectNodeType(VariableNameNode)
			if err != nil {
				return err
			}

			// Variable definitions must be on the same line
			// We introduce this special flag to handle this instead of auto-wrapping
			f.shouldWrap = false
			f.write(fmt.Sprintf("%s =", variableName.Literal.(string)))

			events, err := node.Children[1].expectNodeType(EventSequenceNode)
			if err != nil {
				return err
			}

			if len(events.Children) == 1 && events.Children[0].Type == EventSequenceNode {
				f.write("[")
				f.indent()
				f.shouldWrap = true

				err = f.format(events.Children[0].Children...)
				if err != nil {
					return err
				}

				f.unindent()
				f.write("]")
			} else {
				err = f.format(events.Children...)
				if err != nil {
					return err
				}
			}

			f.shouldWrap = true
			f.flush()

		case EventSequenceNode:
			// We always indent individual event sequences
			f.flush()
			f.write("[")
			f.indent()

			err := f.format(node.Children...)
			if err != nil {
				return err
			}

			f.unindent()
			f.write("]")
			f.flush()

		case CramNode:
			if err := node.expectNChildren(1, 2); err != nil {
				return err
			}

			events, err := node.Children[0].expectNodeType(EventSequenceNode)
			if err != nil {
				return err
			}

			f.write("{")

			err = f.format(events.Children...)
			if err != nil {
				return err
			}

			if len(node.Children) > 1 {
				durationNode, err := node.Children[1].expectNodeType(DurationNode)
				if err != nil {
					return err
				}

				err = f.formatWithDuration("}", durationNode, "")
				if err != nil {
					return err
				}
			} else {
				f.write("}")
			}

		case RepeatNode:
			if err := node.expectNChildren(2); err != nil {
				return err
			}

			err := f.format(node.Children[0])
			if err != nil {
				return err
			}

			times, err := node.Children[1].expectNodeType(TimesNode)
			if err != nil {
				return err
			}

			f.write(fmt.Sprintf("*%d", times.Literal.(int32)))

		case RepetitionsNode:
			if err := node.expectNChildren(2); err != nil {
				return err
			}

			err := f.format(node.Children[0])
			if err != nil {
				return err
			}

			repetitions, err := node.Children[1].expectNodeType(RepetitionsNode)
			if err != nil {
				return err
			}

			texts := []string{}
			for _, rrNode := range repetitions.Children {
				if err := rrNode.expectNChildren(2); err != nil {
					return err
				}

				frNode, err := rrNode.Children[0].expectNodeType(FirstRepetitionNode)
				if err != nil {
					return err
				}

				lrNode, err := rrNode.Children[1].expectNodeType(LastRepetitionNode)
				if err != nil {
					return err
				}

				fr := frNode.Literal.(int32)
				lr := lrNode.Literal.(int32)

				if fr == lr {
					texts = append(texts, string(fr))
				} else {
					texts = append(texts, fmt.Sprintf("%d-%d", fr, lr))
				}
			}
			f.write(fmt.Sprintf("'%s", strings.Join(texts, ",")))


		case VoiceGroupNode:
			return f.format(node.Children...)

		case VoiceNode:
			if err := node.expectNChildren(2); err != nil {
				return err
			}

			voiceNumber, err := node.Children[0].expectNodeType(VoiceNumberNode)
			if err != nil {
				return err
			}

			f.write(fmt.Sprintf("V%d:", voiceNumber.Literal.(int32)))

			voiceEvents, err := node.Children[1].expectNodeType(EventSequenceNode)
			if err != nil {
				return err
			}

			f.indent()

			err = f.format(voiceEvents.Children...)
			if err != nil {
				return err
			}

			f.unindent()

		case VoiceGroupEndMarkerNode:
			f.write("V0:")
			// Let remaining events continue on the same line
			// In subsequent lines, we remain unindented
			// Indentation behaviour is up to us

		}
	}

	f.flush()
	return nil
}

// formatAST is the overall format for root node
func (f *formatter) formatAST(root ASTNode) error {
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

			err = f.format(events.Children...)
			if err != nil {
				return err
			}

		case PartNode:
			if err := part.expectNChildren(2); err != nil {
				return err
			}

			// Part declaration
			partDecl, err := part.Children[0].expectNodeType(PartDeclarationNode)
			if err != nil {
				return err
			}

			if err := partDecl.expectNChildren(1, 2); err != nil {
				return err
			}

			partNames, err := partDecl.Children[0].expectNodeType(PartNamesNode)
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

			if len(partNames.Children) > 1 {
				partAlias, err := partNames.Children[1].expectNodeType(PartAliasNode)
				if err != nil {
					return err
				}
				f.write(fmt.Sprintf("%s \"%s\":", namesText, partAlias.Literal.(string)))
			} else {
				f.write(fmt.Sprintf("%s:", namesText))
			}

			// Part events
			f.indent()

			events, err := part.Children[1].expectNodeType(EventSequenceNode)
			if err != nil {
				return err
			}

			err = f.format(events.Children...)
			if err != nil {
				return err
			}

			f.unindent()

		}

		if i+1 < len(root.Children) {
			f.emptyLine()
		}
	}

	return nil
}

// FormatASTToCode performs rudimentary output formatting of Alda code
func FormatASTToCode(root ASTNode, out io.Writer, opts ...formatterOption) error {
	// Write to temp buffer instead of directly to file in case of error
	temp := bytes.Buffer{}
	f := newFormatter(&temp, opts...)
	err := f.formatAST(root)
	if err != nil {
		return err
	}
	_, err = out.Write(temp.Bytes())
	return err
}

/*
General note for Dave

I tried to balance simplicity, effectiveness, and cost of development
Keeping in mind the primary focus of this is to get MusicXML import working
Sacrificing a bit of the quality and idiomatic nature of the generated Alda


 */
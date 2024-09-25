package parser

import (
	"alda.io/client/model"
	"fmt"
)

// withDuration returns the same ASTNode with an added DurationNode child if the
// provided model.Duration is non-empty.
func withDuration(node ASTNode, duration model.Duration) (ASTNode, error) {
	if len(duration.Components) == 0 {
		return node, nil
	}

	durationNode := ASTNode{Type: DurationNode}

	for _, component := range duration.Components {
		switch c := component.(type) {

		case model.Barline:
			durationNode.Children = append(durationNode.Children, ASTNode{
				Type: BarlineNode,
			})

		case model.NoteLength:
			noteLength := ASTNode{Type: NoteLengthNode, Children: []ASTNode{{
				Type:    DenominatorNode,
				Literal: c.Denominator,
			}}}
			if c.Dots > 0 {
				noteLength.Children = append(noteLength.Children, ASTNode{
					Type:    DotsNode,
					Literal: c.Dots,
				})
			}
			durationNode.Children = append(durationNode.Children, noteLength)

		case model.NoteLengthMs:
			durationNode.Children = append(durationNode.Children, ASTNode{
				Type:    NoteLengthMsNode,
				Literal: c.Quantity,
			})

		case model.NoteLengthSeconds:
			durationNode.Children = append(durationNode.Children, ASTNode{
				Type:    NoteLengthSecondsNode,
				Literal: c.Quantity,
			})

		default:
			// model.NoteLengthBeats - only generated within lisp
			return ASTNode{}, fmt.Errorf(
				"unexpected DurationComponent during AST generation: %#v", c,
			)

		}
	}

	node.Children = append(node.Children, durationNode)
	return node, nil
}

// mapIsolatedUpdate maps a single isolated model.ScoreUpdate to ASTNode.
// Holistic updates that require "re-construction" are handled upstream:
//  1. Parts in mapTopLevel.
//  2. model.VoiceMarker, model.VoiceGroupEndMarker in mapInnerEvents.
func mapIsolatedUpdate(scoreUpdate model.ScoreUpdate) (ASTNode, error) {
	switch update := scoreUpdate.(type) {

	case model.AtMarker:
		return ASTNode{Type: AtMarkerNode, Literal: update.Name}, nil

	case model.AttributeUpdate:
		switch pu := update.PartUpdate.(type) {

		// The following have direct ASTNode representations
		case model.OctaveSet:
			return ASTNode{Type: OctaveSetNode, Literal: pu.OctaveNumber}, nil

		case model.OctaveUp:
			return ASTNode{Type: OctaveUpNode}, nil

		case model.OctaveDown:
			return ASTNode{Type: OctaveDownNode}, nil

		// Most part updates must be formatted via lisp.
		// We handle the subset that can be generated via MusicXML import.
		// TODO: handle generating all possible part updates into lisp.
		case model.DynamicMarking:
			return ASTNode{Type: LispListNode, Children: []ASTNode{{
				Type:    LispSymbolNode,
				Literal: pu.Marking,
			}}}, nil

		case model.KeySignatureSet:
			// Note: we arbitrarily select one of multiple lisp names.
			// This is ok for now, but would make generated ASTs different from
			// parsed ones (different LispSymbolNode Literal) if ASTNode.Updates
			// ever directly outputs evaluated lisp.
			return ASTNode{Type: LispListNode, Children: []ASTNode{
				{
					Type:    LispSymbolNode,
					Literal: "key-signature",
				},
				{
					Type:    LispStringNode,
					Literal: pu.KeySignature.String(),
				},
			}}, nil

		case model.TranspositionSet:
			return ASTNode{Type: LispListNode, Children: []ASTNode{
				{
					Type:    LispSymbolNode,
					Literal: "transpose",
				},
				{
					Type:    LispNumberNode,
					Literal: pu.Semitones,
				},
			}}, nil

		default:
			return ASTNode{}, fmt.Errorf(
				"unexpected PartUpdate type during AST generation: %#v", pu,
			)

		}

	case model.Barline:
		return ASTNode{Type: BarlineNode}, nil

	case model.Chord:
		children, err := mapInnerEvents(update.Events)
		if err != nil {
			return ASTNode{}, err
		}
		return ASTNode{Type: ChordNode, Children: children}, nil

	case model.Cram:
		children, err := mapInnerEvents(update.Events)
		if err != nil {
			return ASTNode{}, err
		}
		cram := ASTNode{Type: CramNode, Children: []ASTNode{{
			Type:     EventSequenceNode,
			Children: children,
		}}}
		return withDuration(cram, update.Duration)

	case model.EventSequence:
		children, err := mapInnerEvents(update.Events)
		if err != nil {
			return ASTNode{}, err
		}

		if children == nil {
			children = []ASTNode{}
		}
		return ASTNode{Type: EventSequenceNode, Children: children}, nil

	case model.LispList:
		var lispFormToNode func(model.LispForm) (ASTNode, error)
		lispFormToNode = func(lispForm model.LispForm) (ASTNode, error) {
			switch l := lispForm.(type) {

			case model.LispList:
				lispList := ASTNode{Type: LispListNode}
				for _, element := range l.Elements {
					node, err := lispFormToNode(element)
					if err != nil {
						return ASTNode{}, err
					}
					lispList.Children = append(lispList.Children, node)
				}
				return lispList, nil

			case model.LispNumber:
				return ASTNode{Type: LispNumberNode, Literal: l.Value}, nil

			case model.LispQuotedForm:
				node, err := lispFormToNode(l.Form)
				if err != nil {
					return ASTNode{}, err
				}
				return ASTNode{
					Type:     LispQuotedFormNode,
					Children: []ASTNode{node},
				}, nil

			case model.LispString:
				return ASTNode{Type: LispStringNode, Literal: l.Value}, nil

			case model.LispSymbol:
				return ASTNode{Type: LispSymbolNode, Literal: l.Name}, nil

			default:
				return ASTNode{}, fmt.Errorf(
					"unexpected LispForm type during AST gen: %#v", l,
				)

			}
		}
		return lispFormToNode(update)

	case model.Marker:
		return ASTNode{Type: MarkerNode, Literal: update.Name}, nil

	case model.Note:
		note := ASTNode{Type: NoteNode}

		switch pitch := update.Pitch.(type) {

		case model.LetterAndAccidentals:
			laa := ASTNode{Type: NoteLetterAndAccidentalsNode}

			laa.Children = append(laa.Children, ASTNode{
				Type:    NoteLetterNode,
				Literal: rune(pitch.NoteLetter + 'a'),
			})

			if len(pitch.Accidentals) > 0 {
				var acc []ASTNode
				for _, accidental := range pitch.Accidentals {
					switch accidental {
					case model.Flat:
						acc = append(acc, ASTNode{Type: FlatNode})
					case model.Natural:
						acc = append(acc, ASTNode{Type: NaturalNode})
					case model.Sharp:
						acc = append(acc, ASTNode{Type: SharpNode})
					}
				}
				laa.Children = append(laa.Children, ASTNode{
					Type:     NoteAccidentalsNode,
					Children: acc,
				})
			}

			note.Children = append(note.Children, laa)

		default:
			// model.MidiNoteNumber - never parsed from Alda code, only used:
			// 	1. with model.LispPitch (unnecessary to handle here)
			// 	2. by MusicXML importer (removed during optimization)
			return ASTNode{}, fmt.Errorf(
				"unexpected PitchIdentifier type during AST gen: %#v", pitch,
			)

		}

		note, err := withDuration(note, update.Duration)
		if err != nil {
			return ASTNode{}, err
		}

		if update.Slurred {
			note.Children = append(note.Children, ASTNode{Type: TieNode})
		}

		return note, nil

	case model.OnRepetitions:
		node, err := mapIsolatedUpdate(update.Event)
		if err != nil {
			return ASTNode{}, err
		}

		repetitions := ASTNode{Type: RepetitionsNode}
		for _, rep := range update.Repetitions {
			repetitions.Children = append(repetitions.Children, ASTNode{
				Type: RepetitionRangeNode,
				Children: []ASTNode{
					{Type: FirstRepetitionNode, Literal: rep.First},
					{Type: LastRepetitionNode, Literal: rep.Last},
				},
			})
		}

		return ASTNode{
			Type:     OnRepetitionsNode,
			Children: []ASTNode{node, repetitions},
		}, nil

	case model.Repeat:
		node, err := mapIsolatedUpdate(update.Event)
		if err != nil {
			return ASTNode{}, err
		}
		return ASTNode{Type: RepeatNode, Children: []ASTNode{
			node,
			{Type: TimesNode, Literal: update.Times},
		}}, nil

	case model.Rest:
		return withDuration(ASTNode{Type: RestNode}, update.Duration)

	case model.VariableDefinition:
		children, err := mapInnerEvents(update.Events)
		if err != nil {
			return ASTNode{}, err
		}
		return ASTNode{Type: VariableDefinitionNode, Children: []ASTNode{
			{Type: VariableNameNode, Literal: update.VariableName},
			{Type: EventSequenceNode, Children: children},
		}}, nil

	case model.VariableReference:
		return ASTNode{
			Type:    VariableReferenceNode,
			Literal: update.VariableName,
		}, nil

	default:
		// PartDeclaration, VoiceMarker, VoiceGroupEndMarker - handled upstream
		// GlobalAttributeUpdate - only generated within Lisp
		// LispNil - should never exist here
		return ASTNode{}, fmt.Errorf(
			"unexpected ScoreUpdate type during AST gen: %#v", update,
		)

	}
}

// mapInnerEvents maps a group of inner updates to a group of ASTNode's.
// Parts are ignored and handled upstream in mapTopLevel.
// Voices are handled by identifying model.VoiceMarker then generating
// subsequent nodes within an overarching VoiceGroupNode.
func mapInnerEvents(updates []model.ScoreUpdate) ([]ASTNode, error) {
	dummyNode := ASTNode{Children: []ASTNode{}}

	var currentVoiceGroup *ASTNode
	currentNode := &dummyNode

	for _, update := range updates {
		switch u := update.(type) {

		case model.VoiceMarker:
			if currentVoiceGroup == nil {
				currentVoiceGroup = &ASTNode{Type: VoiceGroupNode}
			}
			currentVoiceGroup.Children = append(currentVoiceGroup.Children,
				ASTNode{Type: VoiceNode, Children: []ASTNode{
					{Type: VoiceNumberNode, Literal: u.VoiceNumber},
					{Type: EventSequenceNode},
				}})
			currIndex := len(currentVoiceGroup.Children) - 1
			currentNode = &currentVoiceGroup.Children[currIndex].Children[1]

		case model.VoiceGroupEndMarker:
			currentVoiceGroup.Children = append(currentVoiceGroup.Children,
				ASTNode{Type: VoiceGroupEndMarkerNode},
			)
			dummyNode.Children = append(dummyNode.Children, *currentVoiceGroup)
			currentVoiceGroup = nil
			currentNode = &dummyNode

		default:
			node, err := mapIsolatedUpdate(u)
			if err != nil {
				return []ASTNode{}, err
			}
			currentNode.Children = append(currentNode.Children, node)

		}
	}

	if currentVoiceGroup != nil {
		dummyNode.Children = append(dummyNode.Children, *currentVoiceGroup)
	}

	return dummyNode.Children, nil
}

// mapTopLevel maps a top level group of updates to a RootNode.
// mapTopLevel handles PartNode and ImplicitPartNode, then calls mapInnerEvents.
func mapTopLevel(updates []model.ScoreUpdate) (ASTNode, error) {
	root := ASTNode{Type: RootNode}

	// Construct PartNodes by iterating backwards
	currentPartEndIndex := len(updates)
	for i := len(updates) - 1; i >= 0; i-- {
		if partDeclarationUpdate, ok := updates[i].(model.PartDeclaration); ok {
			partDecl := ASTNode{Type: PartDeclarationNode}

			partNames := ASTNode{Type: PartNamesNode}
			for _, name := range partDeclarationUpdate.Names {
				partNames.Children = append(partNames.Children, ASTNode{
					Type:    PartNameNode,
					Literal: name,
				})
			}
			partDecl.Children = append(partDecl.Children, partNames)

			if len(partDeclarationUpdate.Alias) > 0 {
				partDecl.Children = append(partDecl.Children, ASTNode{
					Type:    PartAliasNode,
					Literal: partDeclarationUpdate.Alias,
				})
			}

			events, err := mapInnerEvents(
				updates[i+1 : currentPartEndIndex],
			)
			if err != nil {
				return ASTNode{}, err
			}

			part := ASTNode{Type: PartNode, Children: []ASTNode{
				partDecl,
				{Type: EventSequenceNode, Children: events},
			}}
			root.Children = append([]ASTNode{part}, root.Children...)
			currentPartEndIndex = i
		}
	}

	// Any remaining prefix is put into an ImplicitPartNode
	if currentPartEndIndex > 0 {
		nodes, err := mapInnerEvents(updates[0:currentPartEndIndex])
		if err != nil {
			return ASTNode{}, err
		}

		implicitPart := ASTNode{Type: ImplicitPartNode, Children: []ASTNode{{
			Type:     EventSequenceNode,
			Children: nodes,
		}}}
		root.Children = append([]ASTNode{implicitPart}, root.Children...)
	}

	return root, nil
}

// GenerateASTFromScoreUpdates generates an ASTNode from []model.ScoreUpdate.
// This is a direct inverse of ASTNode.Updates with the exception of
// model.AldaSourceContext which is currently ignored because:
//  1. ASTNode.Updates is lossy and drops model.AldaSourceContext converting
//     DurationNode -> model.Duration. This is the only lost info and can be
//     remedied by adding model.AldaSourceContext to model.DurationComponent.
//  2. ASTNode generation currently doesn't require model.AldaSourceContext.
//     It can always be obtained from the original Alda file.
//     The current use case is MusicXML import, which generates
//     model.ScoreUpdate's without model.AldaSourceContext.
func GenerateASTFromScoreUpdates(updates []model.ScoreUpdate) (ASTNode, error) {
	return mapTopLevel(updates)
}

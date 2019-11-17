package model

import (
	"fmt"
	"strings"

	log "alda.io/client/logging"
	"github.com/mohae/deepcopy"
)

// A PartDeclaration sets the current instruments of the score, creating them if
// necessary.
type PartDeclaration struct {
	Names []string
	Alias string
}

// A Part is a single instance of an instrument used within a score.
//
// A score can include multiple instances of the same type of instrument.
type Part struct {
	Name              string
	StockInstrument   Instrument
	TempoRole         TempoRole
	Tempo             float32
	KeySignature      KeySignature
	Transposition     int32
	ReferencePitch    float32
	CurrentOffset     OffsetMs
	LastOffset        OffsetMs
	Octave            int32
	Volume            float32
	TrackVolume       float32
	Panning           float32
	Quantization      float32
	Duration          Duration
	TimeScale         float32
	CurrentRepetition int32
	// A snapshot copy of the part at the point in time when a voice group starts.
	// This is used as a template for each new voice.
	voiceTemplate *Part
	// We stash this here so that clones can retain a reference to the original.
	origin *Part
	// A record of the clones created, one per voice.
	voices *Voices
	// A reference to the score to which the part belongs.
	score *Score
}

// Clone returns a copy of a part.
func (part *Part) Clone() *Part {
	// mohae/deepcopy doesn't copy private fields.
	//
	// Some fields of Part are deliberately private because if we make them
	// public, deepcopy recurses infinitely through the part copies until the
	// stack overflows.
	clone := deepcopy.Copy(part).(*Part)

	// Instead, we manually copy the fields here.
	clone.origin = part.origin
	clone.voiceTemplate = part.voiceTemplate
	clone.voices = part.voices
	clone.score = part.score

	return clone
}

// NewPart returns a new part in the score.
func (score *Score) NewPart(name string) (*Part, error) {
	stock, err := stockInstrument(name)
	if err != nil {
		return nil, err
	}

	// NB: In Alda v1, the implementation of `new-part` also included:
	//
	// * a randomly generated ID
	// * base initial attribute values (*initial-attr-vals*)
	// * Initial values specific to the stock instrument (`initial-vals`)
	// * Additional attribute values optionally passed into `new-part`.
	//
	// I'm not sure how much of this actually needs to be ported. I think I might
	// be able to get some mileage out of using a part's pointer as its ID, for
	// example.
	part := &Part{
		Name:            name,
		StockInstrument: stock,
		CurrentOffset:   0,
		LastOffset:      -1,
		Octave:          4,
		Tempo:           120,
		Volume:          1.0,
		TrackVolume:     100.0 / 127,
		Panning:         0.5,
		Quantization:    0.9,
		Duration: Duration{
			Components: []DurationComponent{NoteLength{Denominator: 4}},
		},
		TimeScale:      1.0,
		KeySignature:   KeySignature{},
		Transposition:  0,
		ReferencePitch: 440.0,
		voices:         NewVoices(),
		score:          score,
	}

	part.origin = part

	return part, nil
}

// SetAlias defines an alias that refers to 1 more parts.
func (score *Score) SetAlias(alias string, parts []*Part) {
	log.Debug().
		Str("alias", alias).
		Interface("parts", parts).
		Msg("Adding alias.")

	score.Aliases[alias] = parts
}

// NamedParts returns a list of Parts included in the score that correspond to
// the provided `alias`, or nil if there are no such parts.
func (score *Score) NamedParts(alias string) []*Part {
	return score.Aliases[alias]
}

// UnnamedParts returns the list of Parts in the score that are not included in
// any alias, and that are instances of the stock instrument identified by
// `name`.
func (score *Score) UnnamedParts(name string) []*Part {
	stock := "N/A"
	if stockInstrument, err := stockInstrumentName(name); err == nil {
		stock = stockInstrument
	}

	results := []*Part{}

	for _, part := range score.Parts {
		isNamedPart := false
		for _, namedParts := range score.Aliases {
			for _, namedPart := range namedParts {
				if namedPart == part {
					isNamedPart = true
				}
			}
		}

		if !isNamedPart && part.StockInstrument.Name() == stock {
			results = append(results, part)
		}
	}

	return results
}

// AliasedStockInstruments returns the list of Parts in the score that have a
// dedicated alias (e.g. 'piano "foo"'), and that are instances of the stock
// instrument identified by `name`.
func (score *Score) AliasedStockInstruments(name string) []*Part {
	stock := "N/A"
	if stockInstrument, err := stockInstrumentName(name); err == nil {
		stock = stockInstrument
	}

	results := []*Part{}

	for _, namedParts := range score.Aliases {
		if len(namedParts) == 1 {
			part := namedParts[0]
			if part.StockInstrument.Name() == stock {
				results = append(results, part)
			}
		}
	}

	return results
}

// AliasesFor returns the list of aliases in the score that correspond to a
// single part, the one provided.
func (score *Score) AliasesFor(part *Part) []string {
	results := []string{}

	for alias, parts := range score.Aliases {
		if len(parts) == 1 && parts[0] == part {
			results = append(results, alias)
		}
	}

	return results
}

// The PartUpdate interface defines how something updates a part.
type PartUpdate interface {
	updatePart(*Part)
}

// Once an alias is defined for a group, its individual parts can be accessed by
// using the "." (dot) operator.
func dotAccess(name string) bool {
	return strings.ContainsRune(name, '.')
}

func determineParts(decl PartDeclaration, score *Score) ([]*Part, error) {
	// e.g. foo, foo "bar"
	if len(decl.Names) == 1 {
		name := decl.Names[0]
		namedParts := score.NamedParts(name)
		unnamedParts := score.UnnamedParts(name)
		partsForAlias := score.NamedParts(decl.Alias)
		aliasedStockInstruments := score.AliasedStockInstruments(name)

		// If there is an alias, then `name` is expected to be the name of a stock
		// instrument, not the alias of an existing part.
		if decl.Alias != "" && len(namedParts) > 0 {
			return nil, fmt.Errorf(
				"Can't assign alias \"%s\" to existing instance \"%s\"",
				decl.Alias,
				name,
			)
		}

		// Can't redefine an existing alias.
		if decl.Alias != "" && len(partsForAlias) > 0 {
			return nil, fmt.Errorf(
				"The alias \"%s\" has already been assigned to another part/group",
				decl.Alias,
			)
		}

		// Can't use both unnamed and named instances of the same instrument.
		if (decl.Alias != "" && len(unnamedParts) > 0) ||
			(decl.Alias == "" && len(aliasedStockInstruments) > 0 && !dotAccess(name)) {
			return nil, fmt.Errorf(
				"ambiguous instrument reference \"%s\": can't use both unnamed and "+
					"named instances of the same instrument in a score",
				decl.Alias,
			)
		}

		// Always create a new part if there is an alias.
		if decl.Alias != "" {
			part, err := score.NewPart(name)
			if err != nil {
				return nil, err
			}
			return []*Part{part}, nil
		}

		if len(namedParts) > 0 {
			return namedParts, nil
		}

		if len(unnamedParts) > 0 {
			return unnamedParts, nil
		}

		part, err := score.NewPart(name)
		if err != nil {
			return nil, err
		}
		return []*Part{part}, nil
	}

	// Guard against duplicate names, e.g. piano/piano, foo/foo.
	seen := map[string]bool{}
	for _, name := range decl.Names {
		if seen[name] {
			return nil, fmt.Errorf("Name included multiple times in group: %s", name)
		}
		seen[name] = true
	}

	// If we've gotten this far, there are multiple names, e.g.:
	// foo/bar, foo/bar "baz", foo/bar/baz

	namedParts := []*Part{}
	stockParts := []*Part{}

	for _, name := range decl.Names {
		named := score.NamedParts(name)
		unnamed := score.UnnamedParts(name)
		if len(named) > 0 {
			namedParts = append(namedParts, named...)
		} else if len(unnamed) > 0 {
			stockParts = append(stockParts, unnamed...)
		} else {
			part, err := score.NewPart(name)
			if err != nil {
				return nil, err
			}
			stockParts = append(stockParts, part)
		}
	}

	// Can't use both named and stock instruments in a group.
	if len(namedParts) > 0 && len(stockParts) > 0 {
		return nil, fmt.Errorf(
			"Invalid group \"%s\": can't use both stock instruments and named parts",
			strings.Join(decl.Names, "/"),
		)
	}

	// We use a "set" here instead of a slice because it's possible to refer to
	// two existing named groups where the parts covered by each group overlap.
	//
	// For example:
	//
	// piano "foo":
	// trumpet "bar":
	// bassoon "baz":
	// foo/bar "group1":
	// foo/baz "group2":
	// group1/group2 "groups1and2":
	//
	// In this contrived example, `groups1and2` refers to 3 parts, by way of
	// referring to 2 groups of 2 parts that have 1 part in common.
	//
	// By using a set here, we ensure that we don't end up with duplicate parts in
	// this scenario.
	parts := map[*Part]bool{}

	// Always create new parts when creating a named group consisting of stock
	// instruments.
	if decl.Alias != "" && len(stockParts) > 0 {
		for _, name := range decl.Names {
			part, err := score.NewPart(name)
			if err != nil {
				return nil, err
			}
			parts[part] = true
		}

		// Convert set to slice.
		result := []*Part{}
		for part := range parts {
			result = append(result, part)
		}
		return result, nil
	}

	for _, part := range namedParts {
		parts[part] = true
	}

	for _, part := range stockParts {
		parts[part] = true
	}

	// Convert set to slice.
	result := []*Part{}
	for part := range parts {
		result = append(result, part)
	}
	return result, nil
}

// UpdateScore implements ScoreUpdate.UpdateScore by setting the current
// ("active") parts.
//
// When a part is declared, the associated instruments become active, meaning
// that subsequent events (notes, etc.) will be applied to those instruments. A
// part can consist of multiple instruments, can refer to instruments using
// aliases, and can assign an alias to the set of instruments being referenced.
//
// When a reference is made to instrument instances that don't exist yet, the
// appropriate instances are initialized and added to the score.
func (decl PartDeclaration) UpdateScore(score *Score) error {
	// The beginning of a new part (or resumption of an existing part) implicitly
	// ends a voice group in the preceding part if there is one.
	if err := (VoiceGroupEndMarker{}.UpdateScore(score)); err != nil {
		return err
	}

	parts, err := determineParts(decl, score)
	if err != nil {
		return err
	}

	// If this is the first time we're adding an instrument part to the score,
	// then we designate the first part as being the tempo "master."
	if len(score.Parts) == 0 {
		parts[0].TempoRole = TempoRoleMaster
	}

	for _, part := range parts {
		alreadyInScore := false
		for _, existingPart := range score.Parts {
			if existingPart == part {
				alreadyInScore = true
			}
		}

		if !alreadyInScore {
			score.Parts = append(score.Parts, part)
		}
	}

	if decl.Alias != "" {
		score.SetAlias(decl.Alias, parts)

		for _, part := range parts {
			score.SetAlias(decl.Alias+"."+part.Name, []*Part{part})

			for _, alias := range score.AliasesFor(part) {
				score.SetAlias(decl.Alias+"."+alias, []*Part{part})
			}
		}
	}

	score.CurrentParts = parts

	score.ApplyGlobalAttributes()

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since a part
// declaration is conceptually instantaneous.
func (decl PartDeclaration) DurationMs(part *Part) float32 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (decl PartDeclaration) VariableValue(score *Score) (ScoreUpdate, error) {
	return nil, fmt.Errorf(
		"a part declaration cannot be part of a variable definition",
	)
}

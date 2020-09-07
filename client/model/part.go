package model

import (
	"fmt"
	"strings"

	"alda.io/client/json"
	log "alda.io/client/logging"
	"github.com/mohae/deepcopy"
)

// A PartDeclaration sets the current instruments of the score, creating them if
// necessary.
type PartDeclaration struct {
	Names []string
	Alias string
}

// JSON implements RepresentableAsJSON.JSON.
func (decl PartDeclaration) JSON() *json.Container {
	object := json.Object(
		"type", "part-declaration",
		"value", json.Object("names", decl.Names),
	)

	if decl.Alias != "" {
		object.Set(decl.Alias, "alias")
	}

	return object
}

// A Part is a single instance of an instrument used within a score.
//
// A score can include multiple instances of the same type of instrument.
type Part struct {
	Name            string
	StockInstrument Instrument
	TempoRole       TempoRole
	Tempo           float64
	KeySignature    KeySignature
	Transposition   int32
	ReferencePitch  float64
	CurrentOffset   float64
	LastOffset      float64
	Octave          int32
	Volume          float64
	TrackVolume     float64
	Panning         float64
	Quantization    float64
	Duration        Duration
	TimeScale       float64
	// Used for conditionally playing or not playing an event based on how many
	// times through a repeated sequence the part has played so far.
	//
	// See repetitions.go.
	currentRepetition int32
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

// ID returns a unique identifier to the part.
func (part *Part) ID() string {
	return fmt.Sprintf("%p", part)
}

// JSON implements RepresentableAsJSON.JSON.
func (part *Part) JSON() *json.Container {
	return json.Object(
		"name", part.Name,
		"stock-instrument", part.StockInstrument.Name(),
		"tempo-role", part.TempoRole.String(),
		"tempo", part.Tempo,
		"key-signature", part.KeySignature.JSON(),
		"transposition", part.Transposition,
		"reference-pitch", part.ReferencePitch,
		"current-offset", part.CurrentOffset,
		"last-offset", part.LastOffset,
		"octave", part.Octave,
		"volume", part.Volume,
		"track-volume", part.TrackVolume,
		"panning", part.Panning,
		"quantization", part.Quantization,
		"duration", part.Duration.JSON(),
		"time-scale", part.TimeScale,
	)
}

// Clone returns a copy of a part.
func (part *Part) Clone() *Part {
	// mohae/deepcopy doesn't copy private fields.
	//
	// Some fields of Part are deliberately private because if we make them
	// public, deepcopy recurses infinitely through the part copies until the
	// stack overflows. (Some of these are also private just for logical reasons,
	// e.g. implementation details.)
	clone := deepcopy.Copy(part).(*Part)

	// Instead, we manually copy the fields here.
	clone.currentRepetition = part.currentRepetition
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
	json.RepresentableAsJSON

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
		aliasesForPart := map[*Part][]string{}
		for _, part := range parts {
			aliasesForPart[part] = score.AliasesFor(part)
		}

		score.SetAlias(decl.Alias, parts)

		for _, part := range parts {
			score.SetAlias(decl.Alias+"."+part.Name, []*Part{part})

			for _, alias := range aliasesForPart[part] {
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
func (decl PartDeclaration) DurationMs(part *Part) float64 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (decl PartDeclaration) VariableValue(score *Score) (ScoreUpdate, error) {
	return nil, fmt.Errorf(
		"a part declaration cannot be part of a variable definition",
	)
}

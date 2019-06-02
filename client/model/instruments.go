package model

import "fmt"

// An Instrument is a template for a Part.
type Instrument struct {
	Name string
}

// TODO define stock instruments; place each one into the map at least one time,
// once for the proper name and once for each instrument alias
//
// This is a placeholder value for now.
var stockInstruments = map[string]Instrument{
	"accordion": Instrument{"accordion"},
	"bassoon":   Instrument{"bassoon"},
	"clarinet":  Instrument{"clarinet"},
	"flute":     Instrument{"flute"},
	"guitar":    Instrument{"guitar"},
	"piano":     Instrument{"piano"},
	"trombone":  Instrument{"trombone"},
	"trumpet":   Instrument{"trumpet"},
}

// stockInstrument returns a stock instrument, given an identifier which is the
// name or alias of a stock instrument.
//
// Returns an error if the identifier is not recognized as the name or alias of
// a stock instrument.
func stockInstrument(identifier string) (Instrument, error) {
	instrument, hit := stockInstruments[identifier]

	if !hit {
		return Instrument{}, fmt.Errorf("Unrecognized instrument: %s", identifier)
	}

	return instrument, nil
}

// stockInstrumentName returns the name of a stock instrument, given an
// identifier which is the name or alias of a stock instrument.
//
// Returns an error if the identifier is not recognized as the name or alias of
// a stock instrument.
func stockInstrumentName(identifier string) (string, error) {
	instrument, err := stockInstrument(identifier)
	if err != nil {
		return "", err
	}

	return instrument.Name, nil
}

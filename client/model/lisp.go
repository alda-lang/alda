package model

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	log "alda.io/client/logging"
)

// Alda includes a minimal Lisp implementation as a subset of the language, in
// order to facilitate adding new features to the language without accumulating
// syntax.
//
// The Lisp might also facilitate some sort of metaprogramming, but the
// preferred approach is to drive Alda from a Turing-complete programming
// language like Clojure. (See: https://github.com/daveyarwood/alda-clj)

type LispForm interface {
	TypeString() string
	Eval() (LispForm, error)
}

type Operator interface {
	Operate([]LispForm) (LispForm, error)
}

type FunctionSignature struct {
	ArgumentTypes  []LispForm
	Implementation func(...LispForm) (LispForm, error)
}

type LispFunction struct {
	Name       string
	Signatures []FunctionSignature
}

func (LispFunction) TypeString() string {
	return "function"
}

func (f LispFunction) Eval() (LispForm, error) {
	return f, nil
}

func argumentsMatchSignature(
	arguments []LispForm, signature FunctionSignature,
) bool {
	if len(signature.ArgumentTypes) != len(arguments) {
		return false
	}

	for i := 0; i < len(arguments); i++ {
		if reflect.TypeOf(arguments[i]) !=
			reflect.TypeOf(signature.ArgumentTypes[i]) {
			return false
		}
	}

	return true
}

func argumentTypesLine(argumentTypes []LispForm) string {
	if len(argumentTypes) == 0 {
		return "(no arguments)"
	}

	result := []string{}

	for _, argumentType := range argumentTypes {
		result = append(result, argumentType.TypeString())
	}

	return strings.Join(result, ", ")
}

func signatureLines(signatures []FunctionSignature) string {
	lines := []string{}

	for _, signature := range signatures {
		lines = append(lines, argumentTypesLine(signature.ArgumentTypes))

	}

	return strings.Join(lines, "    OR\n")
}

func (f LispFunction) Operate(arguments []LispForm) (LispForm, error) {
	for _, signature := range f.Signatures {
		if argumentsMatchSignature(arguments, signature) {
			return signature.Implementation(arguments...)
		}
	}

	return nil, fmt.Errorf(
		`Provided arguments do not match the signature of %s

Expected:
%s

Got:
%s`,
		"`"+f.Name+"`",
		signatureLines(f.Signatures),
		argumentTypesLine(arguments),
	)
}

var environment = map[string]LispForm{}

type attributeFunctionSignature struct {
	argumentTypes  []LispForm
	implementation func(...LispForm) (PartUpdate, error)
}

func defattribute(names []string, signatures ...attributeFunctionSignature) {
	type defattributeImpl struct {
		name        string
		scoreUpdate func(PartUpdate) ScoreUpdate
	}

	for _, name := range names {
		for _, impl := range []defattributeImpl{
			{
				name: name,
				scoreUpdate: func(partUpdate PartUpdate) ScoreUpdate {
					return AttributeUpdate{PartUpdate: partUpdate}
				},
			},
			{
				name: name + "!",
				scoreUpdate: func(partUpdate PartUpdate) ScoreUpdate {
					return GlobalAttributeUpdate{PartUpdate: partUpdate}
				},
			},
		} {
			functionSignatures := []FunctionSignature{}

			for _, _signature := range signatures {
				// We have to do this in order to close over `signature` in the function
				// below. This is because closures don't work properly in Go.
				//
				// ref: https://www.calhoun.io/gotchas-and-common-mistakes-with-closures-in-go/
				signature := _signature

				functionSignatures = append(functionSignatures, FunctionSignature{
					ArgumentTypes: signature.argumentTypes,
					Implementation: func(args ...LispForm) (LispForm, error) {
						partUpdate, err := signature.implementation(args...)
						if err != nil {
							return nil, err
						}
						return LispScoreUpdate{
							ScoreUpdate: impl.scoreUpdate(partUpdate),
						}, nil
					},
				})
			}

			environment[impl.name] = LispFunction{
				Name: impl.name, Signatures: functionSignatures,
			}
		}
	}
}

func positiveNumber(form LispForm) (float32, error) {
	value := form.(LispNumber).Value

	if value < 1 {
		return 0, fmt.Errorf("Expected positive number, got %f", value)
	}

	return value, nil
}

func nonNegativeNumber(form LispForm) (float32, error) {
	value := form.(LispNumber).Value

	if value < 0 {
		return 0, fmt.Errorf("Expected non-negative number, got %f", value)
	}

	return value / 100, nil
}

func wholeNumber(form LispForm) (int32, error) {
	value := form.(LispNumber).Value

	if value != float32(int32(value)) {
		return 0, fmt.Errorf("Expected whole number, got %f", value)
	}

	return int32(value), nil
}

func percentage(form LispForm) (float32, error) {
	value := form.(LispNumber).Value

	if value < 0 || value > 100 {
		return 0, fmt.Errorf("Value not between 0 and 100: %f", value)
	}

	return value / 100, nil
}

func isDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

func noteLength(str string) (NoteLength, error) {
	chars := []rune(str)

	if len(str) == 0 || !isDigit(chars[0]) {
		return NoteLength{}, fmt.Errorf("Invalid note length: %q", str)
	}

	i := 0

	denominatorChars := []rune{}

	// Consume the digits of an integer.
	for i < len(chars) && isDigit(chars[i]) {
		denominatorChars = append(denominatorChars, chars[i])
		i++
	}

	// If there is a period followed by more digits, treat the period as a decimal
	// and consume the digits on the right-hand side of the decimal.
	if i < len(chars) && chars[i] == '.' &&
		i+1 < len(chars) && isDigit(chars[i+1]) {
		// Consume the decimal.
		denominatorChars = append(denominatorChars, chars[i])
		i++
		// Consume digits.
		for i < len(chars) && isDigit(chars[i]) {
			denominatorChars = append(denominatorChars, chars[i])
			i++
		}
	}

	denominator, _ := strconv.ParseFloat(string(denominatorChars), 32)

	// Any periods remaining are treated as dots.
	dots := 0
	for i < len(chars) && chars[i] == '.' {
		dots++
		i++
	}

	// At this point, we should be at the end of the string. If there's anything
	// left over, consider the string invalid.
	if i < len(chars)-1 {
		return NoteLength{}, fmt.Errorf("Invalid note length: %q", str)
	}

	return NoteLength{Denominator: float32(denominator), Dots: int32(dots)}, nil
}

func duration(form LispForm) (Duration, error) {
	strs := strings.Split(form.(LispString).Value, "~")

	duration := Duration{}

	for _, str := range strs {
		noteLength, err := noteLength(str)
		if err != nil {
			return Duration{}, err
		}

		duration.Components = append(duration.Components, noteLength)
	}

	return duration, nil
}

func isNoteLetter(c rune) bool {
	return 'a' <= c && c <= 'g'
}

func letterAndAccidentals(str string) (NoteLetter, []Accidental, error) {
	validityError := fmt.Errorf(
		"Invalid \"letter and accidentals\" component: %q", str,
	)

	chars := []rune(str)

	if len(str) < 2 || !isNoteLetter(chars[0]) {
		return 0, nil, validityError
	}

	letter, err := NewNoteLetter(chars[0])
	if err != nil {
		return 0, nil, err
	}

	accidentals := []Accidental{}

	for _, c := range chars[1:] {
		switch c {
		case '+':
			accidentals = append(accidentals, Sharp)
		case '-':
			accidentals = append(accidentals, Flat)
		default:
			return 0, []Accidental{}, validityError
		}
	}

	return letter, accidentals, nil
}

func keySignatureFromString(form LispForm) (KeySignature, error) {
	strs := strings.Fields(form.(LispString).Value)

	keySig := KeySignature{}

	for _, str := range strs {
		letter, accidentals, err := letterAndAccidentals(str)
		if err != nil {
			return KeySignature{}, err
		}

		keySig[letter] = accidentals
	}

	return keySig, nil
}

func scaleType(forms []LispForm) (ScaleType, error) {
	validityError := fmt.Errorf("Invalid scale type: %#v", forms)

	// All of the currently supported scale types are a single word. If/when we
	// add more that are multiple words, we'll need to adjust this function
	// accordingly.
	if len(forms) != 1 {
		return 0, validityError
	}

	switch form := forms[0]; form.(type) {
	case LispSymbol:
		switch form.(LispSymbol).Name {
		case "major", "ionian":
			return Ionian, nil
		case "dorian":
			return Dorian, nil
		case "phrygian":
			return Phrygian, nil
		case "lydian":
			return Lydian, nil
		case "mixolydian":
			return Mixolydian, nil
		case "minor", "aeolian":
			return Aeolian, nil
		case "locrian":
			return Locrian, nil
		default:
			return 0, validityError
		}
	default:
		return 0, validityError
	}
}

func keySignatureFromScaleName(forms []LispForm) (KeySignature, error) {
	validityError := fmt.Errorf("Invalid scale name: %#v", forms)

	letter := NoteLetter(0)
	switch form := forms[0]; form.(type) {
	case LispSymbol:
		chars := []rune(form.(LispSymbol).Name)
		if len(chars) > 1 {
			return KeySignature{}, validityError
		}

		ltr, err := NewNoteLetter(chars[0])
		if err != nil {
			return KeySignature{}, err
		}

		letter = ltr
	default:
		return KeySignature{}, validityError
	}

	tonic := Note{NoteLetter: letter}
	passedAccidentals := false
	remainingForms := []LispForm{}

	for _, form := range forms[1:] {
		switch form.(type) {
		case LispSymbol:
			if accidental, err := NewAccidental(form.(LispSymbol).Name); err == nil {
				if passedAccidentals {
					return KeySignature{}, validityError
				}

				tonic.Accidentals = append(tonic.Accidentals, accidental)
				continue
			}

			passedAccidentals = true
			remainingForms = append(remainingForms, form)
		default:
			return KeySignature{}, validityError
		}
	}

	scaleType, err := scaleType(remainingForms)
	if err != nil {
		return KeySignature{}, err
	}

	return KeySignatureFromScale(tonic, scaleType), nil
}

func keySignatureFromAccidentals(forms []LispForm) (KeySignature, error) {
	validityError := fmt.Errorf(
		"Expected pairs of note letter and accidentals, got %#v", forms,
	)

	// We expect to be provided with a list of pairs of letters and accidentals,
	// e.g. (key-signature '(b (flat) e (flat)))
	if len(forms)%2 != 0 {
		return KeySignature{}, validityError
	}

	keySig := KeySignature{}

	for i := 0; i < len(forms); i += 2 {
		letter := NoteLetter(0)
		switch form := forms[i]; form.(type) {
		case LispSymbol:
			chars := []rune(form.(LispSymbol).Name)
			ltr, err := NewNoteLetter(chars[0])
			if err != nil {
				return KeySignature{}, err
			}

			letter = ltr
		default:
			return KeySignature{}, validityError
		}

		accidentals := []Accidental{}
		switch form := forms[i+1]; form.(type) {
		case LispList:
			for _, form := range form.(LispList).Elements {
				switch form.(type) {
				case LispSymbol:
					switch form.(LispSymbol).Name {
					case "flat":
						accidentals = append(accidentals, Flat)
					case "sharp":
						accidentals = append(accidentals, Sharp)
					default:
						return KeySignature{}, validityError
					}
				default:
					return KeySignature{}, validityError
				}
			}
		default:
			return KeySignature{}, validityError
		}

		keySig[letter] = accidentals
	}

	return keySig, nil
}

func keySignatureFromList(form LispForm) (KeySignature, error) {
	forms := form.(LispList).Elements
	validityError := fmt.Errorf("Invalid key signature: %#v", forms)

	if len(forms) < 2 {
		return KeySignature{}, validityError
	}

	switch forms[1].(type) {
	case LispSymbol:
		return keySignatureFromScaleName(forms)
	case LispList:
		return keySignatureFromAccidentals(forms)
	default:
		return KeySignature{}, validityError
	}
}

func init() {
	// Current octave. Used to calculate the pitch of notes.
	defattribute([]string{"octave"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				octaveNumber, err := wholeNumber(args[0])
				if err != nil {
					return nil, err
				}

				return OctaveSet{OctaveNumber: octaveNumber}, nil
			},
		},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispSymbol{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				symbol := args[0].(LispSymbol)

				switch symbol.Name {
				case "up":
					return OctaveUp{}, nil
				case "down":
					return OctaveDown{}, nil
				default:
					return nil, fmt.Errorf("Invalid argument to `octave`: %s", symbol.String())
				}
			},
		},
	)

	// Current tempo. Used to calculate the duration of notes.
	defattribute([]string{"tempo"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				bpm, err := positiveNumber(args[0])
				if err != nil {
					return nil, err
				}
				return TempoSet{Tempo: bpm}, nil
			},
		},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}, LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				noteLength, err := positiveNumber(args[0])
				if err != nil {
					return nil, err
				}

				pseudoBpm, err := positiveNumber(args[1])
				if err != nil {
					return nil, err
				}

				beats, err := NoteLength{Denominator: noteLength}.Beats()
				if err != nil {
					return nil, err
				}

				return TempoSet{Tempo: beats * pseudoBpm}, nil
			},
		},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispString{}, LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				duration, err := duration(args[0])
				if err != nil {
					return nil, err
				}

				pseudoBpm, err := positiveNumber(args[1])
				if err != nil {
					return nil, err
				}

				beats, err := duration.Beats()
				if err != nil {
					return nil, err
				}

				return TempoSet{Tempo: beats * pseudoBpm}, nil
			},
		},
	)

	// Express tempo in terms of metric modulation, where the new note takes the
	// same amount of time (one beat) as the old note.
	//
	// (e.g. (metric-modulation "4." 2) means that the new length of a half note
	// equals the length of a dotted quarter note in the previous measure)
	defattribute([]string{"metric-modulation"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}, LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				oldValue, err := positiveNumber(args[0])
				if err != nil {
					return nil, err
				}

				newValue, err := positiveNumber(args[1])
				if err != nil {
					return nil, err
				}

				oldBeats, err := NoteLength{Denominator: oldValue}.Beats()
				if err != nil {
					return nil, err
				}

				newBeats, err := NoteLength{Denominator: newValue}.Beats()
				if err != nil {
					return nil, err
				}

				return MetricModulation{Ratio: newBeats / oldBeats}, nil
			},
		},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}, LispString{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				oldValue, err := positiveNumber(args[0])
				if err != nil {
					return nil, err
				}

				newValue, err := duration(args[1])
				if err != nil {
					return nil, err
				}

				oldBeats, err := NoteLength{Denominator: oldValue}.Beats()
				if err != nil {
					return nil, err
				}

				newBeats, err := newValue.Beats()
				if err != nil {
					return nil, err
				}

				return MetricModulation{Ratio: newBeats / oldBeats}, nil
			},
		},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispString{}, LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				oldValue, err := duration(args[0])
				if err != nil {
					return nil, err
				}

				newValue, err := positiveNumber(args[1])
				if err != nil {
					return nil, err
				}

				oldBeats, err := oldValue.Beats()
				if err != nil {
					return nil, err
				}

				newBeats, err := NoteLength{Denominator: newValue}.Beats()
				if err != nil {
					return nil, err
				}

				return MetricModulation{Ratio: newBeats / oldBeats}, nil
			},
		},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispString{}, LispString{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				oldValue, err := duration(args[0])
				if err != nil {
					return nil, err
				}

				newValue, err := duration(args[1])
				if err != nil {
					return nil, err
				}

				oldBeats, err := oldValue.Beats()
				if err != nil {
					return nil, err
				}

				newBeats, err := newValue.Beats()
				if err != nil {
					return nil, err
				}

				return MetricModulation{Ratio: newBeats / oldBeats}, nil
			},
		},
	)

	// The percentage of a note that is heard. Used to put a little space between
	// notes.
	//
	// e.g. with a quantization value of 90%, a note that would otherwise last 500
	// ms will be quantized to last 450 ms. The resulting note event will have a
	// duration of 450 ms, and the next event will be set to occur in 500 ms.
	defattribute([]string{"quantization", "quantize", "quant"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				percentage, err := nonNegativeNumber(args[0])
				if err != nil {
					return nil, err
				}
				return QuantizationSet{Quantization: percentage}, nil
			},
		},
	)

	// Current volume. For MIDI purposes, the velocity of individual notes.
	defattribute([]string{"volume", "vol"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				percentage, err := percentage(args[0])
				if err != nil {
					return nil, err
				}
				return VolumeSet{Volume: percentage}, nil
			},
		},
	)

	// More general volume for the track as a whole. Although this can be changed
	// just as often as volume, to do so is not idiomatic. For MIDI purposes, this
	// corresponds to the volume of a channel."
	defattribute([]string{"track-volume", "track-vol"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				percentage, err := percentage(args[0])
				if err != nil {
					return nil, err
				}
				return TrackVolumeSet{TrackVolume: percentage}, nil
			},
		},
	)

	// Current panning. 0 = hard left, 100 = hard right.
	defattribute([]string{"panning", "pan"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				percentage, err := percentage(args[0])
				if err != nil {
					return nil, err
				}
				return PanningSet{Panning: percentage}, nil
			},
		},
	)

	// Default note duration in beats.
	defattribute([]string{"set-duration"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				beats, err := positiveNumber(args[0])
				if err != nil {
					return nil, err
				}
				return DurationSet{Duration: Duration{
					Components: []DurationComponent{NoteLengthBeats{Quantity: beats}},
				}}, nil
			},
		},
	)

	// Default note duration in milliseconds.
	defattribute([]string{"set-duration-ms"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				ms, err := positiveNumber(args[0])
				if err != nil {
					return nil, err
				}
				return DurationSet{Duration: Duration{
					Components: []DurationComponent{NoteLengthMs{Quantity: ms}},
				}}, nil
			},
		},
	)

	// Default note duration, expressed as a note length.
	// e.g. 4 = quarter note, "2.." = dotted half note
	//
	// Can also be expressed as multiple note lengths, tied together.
	// e.g. "1~1" = the length of two whole notes
	defattribute([]string{"set-note-length"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				denominator, err := positiveNumber(args[0])
				if err != nil {
					return nil, err
				}
				return DurationSet{Duration: Duration{
					Components: []DurationComponent{NoteLength{Denominator: denominator}},
				}}, nil
			},
		},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispString{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				duration, err := duration(args[0])
				if err != nil {
					return nil, err
				}
				return DurationSet{Duration: duration}, nil
			},
		},
	)

	defattribute([]string{"key-signature", "key-sig"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispString{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				keySig, err := keySignatureFromString(args[0])
				if err != nil {
					return nil, err
				}
				return KeySignatureSet{KeySignature: keySig}, nil
			},
		},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispList{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				keySig, err := keySignatureFromList(args[0])
				if err != nil {
					return nil, err
				}
				return KeySignatureSet{KeySignature: keySig}, nil
			},
		},
	)

	// The number of semitones to transpose. A negative number means transpose
	// down, a positive number means transpose up.
	defattribute([]string{"transposition", "transpose"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				semitones, err := wholeNumber(args[0])
				if err != nil {
					return nil, err
				}
				return TranspositionSet{Semitones: semitones}, nil
			},
		},
	)

	// The reference pitch, expressed as the frequency of A4 in Hz.
	defattribute([]string{"reference-pitch", "tuning-constant"},
		attributeFunctionSignature{
			argumentTypes: []LispForm{LispNumber{}},
			implementation: func(args ...LispForm) (PartUpdate, error) {
				frequency, err := positiveNumber(args[0])
				if err != nil {
					return nil, err
				}
				return ReferencePitchSet{Frequency: frequency}, nil
			},
		},
	)
}

type LispNil struct{}

func (LispNil) TypeString() string {
	return "nil"
}

func (LispNil) Eval() (LispForm, error) {
	return LispNil{}, nil
}

type LispQuotedForm struct {
	Form LispForm
}

func (LispQuotedForm) TypeString() string {
	return "quoted form"
}

func (qf LispQuotedForm) Eval() (LispForm, error) {
	return qf.Form, nil
}

type LispSymbol struct {
	Name string
}

func (LispSymbol) TypeString() string {
	return "symbol"
}

func (sym LispSymbol) String() string {
	return "'" + sym.Name
}

func (sym LispSymbol) Eval() (LispForm, error) {
	value, hit := environment[sym.Name]

	if !hit {
		return nil, fmt.Errorf("Unresolvable symbol: %s", sym.Name)
	}

	return value, nil
}

type LispNumber struct {
	Value float32
}

func (LispNumber) TypeString() string {
	return "number"
}

func (n LispNumber) Eval() (LispForm, error) {
	return n, nil
}

type LispString struct {
	Value string
}

func (LispString) TypeString() string {
	return "string"
}

func (s LispString) Eval() (LispForm, error) {
	return s, nil
}

type LispScoreUpdate struct {
	ScoreUpdate ScoreUpdate
}

func (su LispScoreUpdate) TypeString() string {
	return "score update"
}

func (su LispScoreUpdate) Eval() (LispForm, error) {
	return su, nil
}

type LispList struct {
	Elements []LispForm
}

func (LispList) TypeString() string {
	return "list"
}

func (l LispList) Eval() (LispForm, error) {
	operator, err := l.Elements[0].Eval()
	if err != nil {
		return nil, err
	}

	arguments := []LispForm{}
	for _, argument := range l.Elements[1:] {
		value, err := argument.Eval()
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, value)
	}

	switch operator.(type) {
	case Operator:
		return operator.(Operator).Operate(arguments)
	default:
		return nil, fmt.Errorf("Value is not an Operator: %#v", operator)
	}
}

// UpdateScore implements ScoreUpdate.UpdateScore by evaluating the S-expression
// and using the resulting value to update the score.
func (l LispList) UpdateScore(score *Score) error {
	result, err := l.Eval()
	if err != nil {
		return err
	}

	switch result.(type) {
	case LispScoreUpdate:
		return result.(LispScoreUpdate).ScoreUpdate.UpdateScore(score)
	default:
		log.Warn().
			Interface("result", result).
			Msg("S-expression result is not a ScoreUpdate.")
		return nil
	}
}

// DurationMs implements ScoreUpdate.DurationMs by evaluating the S-expression
// and returning the duration of the resulting value.
func (l LispList) DurationMs(part *Part) float32 {
	// FIXME: We end up evaluating this a second time when UpdateScore is called.
	// This will be problematic if/when we add functions that have side effects.
	//
	// At that point, we should probably memoize the evaluation result (and error)
	// so that they are simply returned on successive evaluations.
	result, err := l.Eval()

	// If there is an error during evaluation, it will be propagated through when
	// we evaluate again for UpdateScore. So we can safely ignore it here and fall
	// back to a duration of 0.
	if err != nil {
		return 0
	}

	switch result.(type) {
	case LispScoreUpdate:
		return result.(LispScoreUpdate).ScoreUpdate.DurationMs(part)
	default:
		log.Warn().
			Interface("result", result).
			Msg("S-expression result is not a ScoreUpdate.")
		return 0
	}
}

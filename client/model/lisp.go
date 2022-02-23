package model

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"alda.io/client/json"
	log "alda.io/client/logging"
)

// Alda includes a minimal Lisp implementation as a subset of the language, in
// order to facilitate adding new features to the language without accumulating
// syntax.
//
// The Lisp might also facilitate some sort of metaprogramming, but the
// preferred approach is to drive Alda from a Turing-complete programming
// language like Clojure. (See: https://github.com/daveyarwood/alda-clj)

// The LispForm interface is implemented by the various types available in
// alda-lisp.
type LispForm interface {
	json.RepresentableAsJSON

	// TypeString returns a human-readable name of a form's type.
	TypeString() string

	// Eval returns the value of a form when evaluated, or an error if
	// evaluation is unsuccessful.
	Eval() (LispForm, error)
}

// NOTE: I'm not sure yet whether it actually makes sense for special forms to
// implement the LispForm interface. I'm not sure that all of the functions in
// that interface are really applicable to special forms. For now, I'm just
// going to assume that they are and we'll see how it goes.

// LispSpecialFormQuote is the special form `quote`.
type LispSpecialFormQuote struct{}

// JSON implements RepresentableAsJSON.JSON.
func (LispSpecialFormQuote) JSON() *json.Container {
	return json.Object(
		"type", "special-form",
		"value", "quote",
	)
}

// TypeString implements LispForm.TypeString.
func (LispSpecialFormQuote) TypeString() string {
	return "special-form"
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (LispSpecialFormQuote) GetSourceContext() AldaSourceContext {
	// I'm not sure yet how to handle source context for special forms. It seems
	// like I might not need to, because really, the thing that has the source
	// context is the LispSymbol that evaluated to the special form, and I'm not
	// sure that anything is ever going to call GetSourceContext() on a special
	// form.
	//
	// I could easily be wrong about this. Let's just roll with this for now and
	// see how it goes.
	return AldaSourceContext{}
}

// Eval implements LispForm.Eval by returning the special form.
func (q LispSpecialFormQuote) Eval() (LispForm, error) {
	return q, nil
}

// An Operator is something that takes 0 or more forms and returns a form or an
// error..
type Operator interface {
	// Operate takes 0 or more forms and returns a form.
	//
	// Returns an error if the operation was unsuccessful.
	Operate([]LispForm) (LispForm, error)
}

// A FunctionSignature defines what a function does when called with arguments
// that are a certain combination of types.
type FunctionSignature struct {
	ArgumentTypes  []LispForm
	Implementation func(...LispForm) (LispForm, error)
}

// LispAny represents any type of value.
//
// LispAny is not meant to be an argument or return type itself. It should never
// occur as a value in alda-lisp.
type LispAny struct{}

// JSON implements LispForm.JSON.
//
// Note that this should never get called, as LispAny should never occur as a
// value. I'm implementing it here simply because the compiler is forcing me to!
// (And so that if it does, for some weird reason, occur as a value, at least it
// will be represented as JSON in a sensible way.)
func (LispAny) JSON() *json.Container {
	return json.Object("type", "any-indicator")
}

// TypeString implements LispForm.TypeString.
func (LispAny) TypeString() string {
	return "any"
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (LispAny) GetSourceContext() AldaSourceContext {
	// It isn't possible to parse a LispAny instance from an Alda score, so there
	// is no way we can provide source context.
	return AldaSourceContext{}
}

// Eval implements LispForm.Eval by returning an error, because LispAny is not
// meant to be used as an argument or return type.
func (LispAny) Eval() (LispForm, error) {
	return nil, &AldaSourceError{
		Context: AldaSourceContext{},
		Err:     fmt.Errorf("LispAny is not a valid value type"),
	}
}

// LispVariadic wraps a LispForm as a way of representing that it appears 0 or
// more times in a list of arguments.
//
// LispVariadic is not meant to be an argument or return type itself. It should
// never occur as a value in alda-lisp.
type LispVariadic struct {
	Type LispForm
}

// JSON implements LispForm.JSON.
//
// Note that this should never get called, as LispVariadic should never occur as
// a value. I'm implementing it here simply because the compiler is forcing me
// to! (And so that if it does, for some weird reason, occur as a value, at
// least it will be represented as JSON in a sensible way.)
func (v LispVariadic) JSON() *json.Container {
	return json.Object("type", "variadic-indicator", "value", v.Type.JSON())
}

// TypeString implements LispForm.TypeString.
func (v LispVariadic) TypeString() string {
	return v.Type.TypeString() + "*"
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (v LispVariadic) GetSourceContext() AldaSourceContext {
	// It isn't possible to parse a LispVariadic instance from an Alda score, so
	// there is no way we can provide source context.
	return AldaSourceContext{}
}

// Eval implements LispForm.Eval by returning an error, because LispVariadic is
// not meant to be used as an argument or return type.
func (v LispVariadic) Eval() (LispForm, error) {
	return nil, &AldaSourceError{
		Context: v.GetSourceContext(),
		Err:     fmt.Errorf("LispVariadic is not a valid value type"),
	}
}

// A LispFunction is a function.
type LispFunction struct {
	Name       string
	Signatures []FunctionSignature
}

// JSON implements RepresentableAsJSON.JSON.
func (f LispFunction) JSON() *json.Container {
	return json.Object(
		"type", "function",
		// We could represent the signatures here, too, but at some point, we are
		// not feasibly able to serialize the value of the function, so we have to
		// draw a line somewhere.
		"value", fmt.Sprintf("<fn %s>", f.Name),
	)
}

// TypeString implements LispForm.TypeString.
func (LispFunction) TypeString() string {
	return "function"
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (LispFunction) GetSourceContext() AldaSourceContext {
	// It isn't possible to parse a LispFunction instance* from an Alda score, so
	// there is no way we can provide source context.
	//
	// *At some point in the future, we may add a `fn` special form, but even
	// then, the thing that has the source context is the LispList that represents
	// the S-expression `(fn ...)`
	return AldaSourceContext{}
}

// Eval implements LispForm.Eval by returning the function.
func (f LispFunction) Eval() (LispForm, error) {
	return f, nil
}

// Validate returns an error if the function is invalid.
func (f LispFunction) Validate() error {
	for _, signature := range f.Signatures {
		// Check that the argument list doesn't have a LispVariadic type somewhere
		// other than at the end.
		for i, argType := range signature.ArgumentTypes {
			switch argType.(type) {
			case LispVariadic:
				if i != len(signature.ArgumentTypes)-1 {
					return fmt.Errorf(
						"varargs not at the end of the argument list: %#v",
						signature.ArgumentTypes,
					)
				}
			}
		}
	}

	return nil
}

func argumentsMatchSignature(
	arguments []LispForm, signature FunctionSignature,
) bool {
	anyType := reflect.TypeOf(LispAny{})

	totalArgs := len(signature.ArgumentTypes)

	if totalArgs == 0 && len(arguments) == 0 {
		return true
	}

	variadic := false
	switch signature.ArgumentTypes[totalArgs-1].(type) {
	case LispVariadic:
		variadic = true
	}

	fixedArgs := totalArgs
	if variadic {
		fixedArgs--
	}

	if variadic && len(arguments) < fixedArgs {
		return false
	}

	if !variadic && len(arguments) != fixedArgs {
		return false
	}

	for i := 0; i < fixedArgs; i++ {
		expectedType := reflect.TypeOf(signature.ArgumentTypes[i])
		actualType := reflect.TypeOf(arguments[i])

		if actualType != expectedType && expectedType != anyType {
			return false
		}
	}

	if len(arguments) > fixedArgs {
		variadicArgType := signature.ArgumentTypes[totalArgs-1].(LispVariadic).Type
		expectedType := reflect.TypeOf(variadicArgType)

		for _, argument := range arguments[fixedArgs:] {
			actualType := reflect.TypeOf(argument)
			if actualType != expectedType && expectedType != anyType {
				return false
			}
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

// Operate determines which function signature to use based on the types of the
// arguments and calls the appropriate function on the arguments.
//
// Returns an error if the arguments do not match any of the function's
// signatures.
func (f LispFunction) Operate(arguments []LispForm) (LispForm, error) {
	for _, signature := range f.Signatures {
		if argumentsMatchSignature(arguments, signature) {
			return signature.Implementation(arguments...)
		}
	}

	return nil, fmt.Errorf(
		`provided arguments do not match the signature of %s

Expected:
%s

Got:
%s`,
		"`"+f.Name+"`",
		signatureLines(f.Signatures),
		argumentTypesLine(arguments),
	)
}

var specialForms = map[string]LispForm{
	"quote": LispSpecialFormQuote{},
}

var environment = map[string]LispForm{}

type attributeFunctionSignature struct {
	argumentTypes  []LispForm
	implementation func(...LispForm) (PartUpdate, error)
}

func def(name string, value LispForm) {
	environment[name] = value
}

func defn(name string, signatures ...FunctionSignature) {
	fn := LispFunction{Name: name, Signatures: signatures}

	// This `defn` is only used for defining built-in functions as part of the
	// runtime. `panic` makes sense here because we want the whole system to fall
	// over if any of the built-in functions we've defined are invalid.
	//
	// If/when we implement `defn` in user-space, we should handle the validation
	// error instead.
	if err := fn.Validate(); err != nil {
		panic(err)
	}

	def(name, fn)
}

func defattribute(names []string, signatures ...attributeFunctionSignature) {
	type defattributeImpl struct {
		name        string
		scoreUpdate func(PartUpdate) ScoreUpdate
	}

	for _, name := range names {
		for _, _impl := range []defattributeImpl{
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
			// We have to do this in order to close over `impl` in the function below.
			// This is because closures don't work properly in Go.
			//
			// ref: https://www.calhoun.io/gotchas-and-common-mistakes-with-closures-in-go/
			impl := _impl

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

			defn(impl.name, functionSignatures...)
		}
	}
}

func positiveNumber(form LispForm) (float64, error) {
	number := form.(LispNumber)

	if number.Value < 1 {
		return 0, &AldaSourceError{
			Context: number.SourceContext,
			Err:     fmt.Errorf("expected positive number, got %f", number.Value),
		}
	}

	return number.Value, nil
}

func nonNegativeNumber(form LispForm) (float64, error) {
	number := form.(LispNumber)

	if number.Value < 0 {
		return 0, &AldaSourceError{
			Context: number.SourceContext,
			Err:     fmt.Errorf("expected non-negative number, got %f", number.Value),
		}
	}

	return number.Value / 100, nil
}

func integer(form LispForm) (int32, error) {
	number := form.(LispNumber)

	if number.Value != float64(int32(number.Value)) {
		return 0, &AldaSourceError{
			Context: number.SourceContext,
			Err:     fmt.Errorf("expected integer, got %f", number.Value),
		}
	}

	return int32(number.Value), nil
}

func percentage(form LispForm) (float64, error) {
	number := form.(LispNumber)

	if number.Value < 0 || number.Value > 100 {
		return 0, &AldaSourceError{
			Context: number.SourceContext,
			Err:     fmt.Errorf("value not between 0 and 100: %f", number.Value),
		}
	}

	return number.Value / 100, nil
}

func isDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

func noteLength(str string) (NoteLength, error) {
	chars := []rune(str)

	if len(str) == 0 || !isDigit(chars[0]) {
		return NoteLength{}, fmt.Errorf("invalid note length: %q", str)
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

	denominator, _ := strconv.ParseFloat(string(denominatorChars), 64)

	// Any periods remaining are treated as dots.
	dots := 0
	for i < len(chars) && chars[i] == '.' {
		dots++
		i++
	}

	// At this point, we should be at the end of the string. If there's anything
	// left over, consider the string invalid.
	if i < len(chars)-1 {
		return NoteLength{}, fmt.Errorf("invalid note length: %q", str)
	}

	return NoteLength{Denominator: denominator, Dots: int32(dots)}, nil
}

func duration(form LispForm) (Duration, error) {
	stringLiteral := form.(LispString)

	strs := strings.Split(stringLiteral.Value, "~")

	duration := Duration{}

	for _, str := range strs {
		noteLength, err := noteLength(str)
		if err != nil {
			return Duration{}, &AldaSourceError{
				Context: stringLiteral.SourceContext,
				Err:     err,
			}
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
		"invalid \"letter and accidentals\" component: %q", str,
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
	stringLiteral := form.(LispString)

	strs := strings.Fields(stringLiteral.Value)

	keySig := KeySignature{}

	for _, str := range strs {
		letter, accidentals, err := letterAndAccidentals(str)
		if err != nil {
			return KeySignature{}, &AldaSourceError{
				Context: stringLiteral.SourceContext,
				Err:     err,
			}
		}

		keySig[letter] = accidentals
	}

	return keySig, nil
}

func scaleType(forms []LispForm) (ScaleType, error) {
	validityError := fmt.Errorf("invalid scale type: %#v", forms)

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
	validityError := fmt.Errorf("invalid scale name: %#v", forms)

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

	tonic := LetterAndAccidentals{NoteLetter: letter}
	passedAccidentals := false
	remainingForms := []LispForm{}

	for _, form := range forms[1:] {
		switch form := form.(type) {
		case LispSymbol:
			if accidental, err := NewAccidental(form.Name); err == nil {
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
		"expected pairs of note letter and accidentals, got %#v", forms,
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
				switch form := form.(type) {
				case LispSymbol:
					switch form.Name {
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
	list := form.(LispList)

	sourceError := func(err error) error {
		return &AldaSourceError{
			Context: list.SourceContext,
			Err:     err,
		}
	}

	forms := list.Elements
	validityError := sourceError(fmt.Errorf("invalid key signature: %#v", forms))

	if len(forms) < 2 {
		return KeySignature{}, validityError
	}

	switch forms[1].(type) {
	case LispSymbol:
		keySig, err := keySignatureFromScaleName(forms)
		if err != nil {
			return KeySignature{}, sourceError(err)
		}
		return keySig, nil
	case LispList:
		keySig, err := keySignatureFromAccidentals(forms)
		if err != nil {
			return KeySignature{}, sourceError(err)
		}
		return keySig, nil
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
				octaveNumber, err := integer(args[0])
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
					return nil, &AldaSourceError{
						Context: symbol.SourceContext,
						Err: fmt.Errorf(
							"invalid argument to `octave`: %s", symbol.String(),
						),
					}
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

				beats := NoteLength{Denominator: noteLength}.Beats()

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

				beats := duration.Beats()

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

				oldBeats := NoteLength{Denominator: oldValue}.Beats()
				newBeats := NoteLength{Denominator: newValue}.Beats()

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

				oldBeats := NoteLength{Denominator: oldValue}.Beats()
				newBeats := newValue.Beats()

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

				oldBeats := oldValue.Beats()
				newBeats := NoteLength{Denominator: newValue}.Beats()

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

				oldBeats := oldValue.Beats()
				newBeats := newValue.Beats()

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

	// Dynamic markings corresponding to a volume set
	var dynamicImplementation = func(marking string) func(args ...LispForm) (PartUpdate, error) {
		return func(args ...LispForm) (PartUpdate, error) {
			return DynamicMarking{Marking: marking}, nil
		}
	}

	for marking := range DynamicVolumes {
		defattribute([]string{marking},
			attributeFunctionSignature{
				argumentTypes:  []LispForm{},
				implementation: dynamicImplementation(marking),
			},
		)
	}

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
				semitones, err := integer(args[0])
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

	defn("list",
		FunctionSignature{
			ArgumentTypes: []LispForm{LispVariadic{LispAny{}}},
			Implementation: func(args ...LispForm) (LispForm, error) {
				return LispList{Elements: args}, nil
			},
		},
	)

	defn("ms",
		FunctionSignature{
			ArgumentTypes: []LispForm{LispNumber{}},
			Implementation: func(args ...LispForm) (LispForm, error) {
				quantity, err := positiveNumber(args[0])
				if err != nil {
					return nil, err
				}
				return LispDuration{NoteLengthMs{Quantity: quantity}}, nil
			},
		},
	)

	defn("note-length",
		FunctionSignature{
			ArgumentTypes: []LispForm{LispNumber{}},
			Implementation: func(args ...LispForm) (LispForm, error) {
				denominator, err := positiveNumber(args[0])
				if err != nil {
					return nil, err
				}

				noteLength := NoteLength{Denominator: denominator}
				return LispDuration{DurationComponent: noteLength}, nil
			},
		},
		FunctionSignature{
			ArgumentTypes: []LispForm{LispString{}},
			Implementation: func(args ...LispForm) (LispForm, error) {
				noteLength, err := noteLength(args[0].(LispString).Value)
				if err != nil {
					return nil, err
				}

				return LispDuration{DurationComponent: noteLength}, nil
			},
		},
	)

	defn("duration",
		FunctionSignature{
			ArgumentTypes: []LispForm{LispVariadic{LispDuration{}}},
			Implementation: func(args ...LispForm) (LispForm, error) {
				duration := Duration{}

				for _, arg := range args {
					component := arg.(LispDuration).DurationComponent
					duration.Components = append(duration.Components, component)
				}

				return LispDuration{DurationComponent: duration}, nil
			},
		},
	)

	defn("midi-note",
		FunctionSignature{
			ArgumentTypes: []LispForm{LispNumber{}},
			Implementation: func(args ...LispForm) (LispForm, error) {
				noteNumber, err := integer(args[0])
				if err != nil {
					return nil, err
				}
				return LispPitch{MidiNoteNumber{MidiNote: noteNumber}}, nil
			},
		},
	)

	defn("pitch",
		FunctionSignature{
			ArgumentTypes: []LispForm{LispList{}},
			Implementation: func(args ...LispForm) (LispForm, error) {
				forms := args[0].(LispList).Elements
				validityError := fmt.Errorf("invalid letter/accidentals: %#v", forms)

				if len(forms) == 0 {
					return nil, validityError
				}

				pitch := LetterAndAccidentals{}

				switch forms[0].(type) {
				case LispSymbol:
					symbol := forms[0].(LispSymbol).Name
					chars := []rune(symbol)

					if len(chars) != 1 {
						return nil, validityError
					}

					noteLetter, err := NewNoteLetter(chars[0])
					if err != nil {
						return nil, err
					}

					pitch.NoteLetter = noteLetter
				default:
					return nil, validityError
				}

				if len(forms) > 1 {
					accidentals := []Accidental{}

					for _, form := range forms[1:] {
						switch form := form.(type) {
						case LispSymbol:
							accidental, err := NewAccidental(form.Name)
							if err != nil {
								return nil, err
							}

							accidentals = append(accidentals, accidental)
						default:
							return nil, validityError
						}
					}

					pitch.Accidentals = accidentals
				}

				return LispPitch{PitchIdentifier: pitch}, nil
			},
		},
	)

	defn("note",
		FunctionSignature{
			ArgumentTypes: []LispForm{LispPitch{}},
			Implementation: func(args ...LispForm) (LispForm, error) {
				pitch := args[0].(LispPitch).PitchIdentifier
				note := Note{Pitch: pitch}
				return LispScoreUpdate{ScoreUpdate: note}, nil
			},
		},
		FunctionSignature{
			ArgumentTypes: []LispForm{LispPitch{}, LispDuration{}},
			Implementation: func(args ...LispForm) (LispForm, error) {
				pitch := args[0].(LispPitch).PitchIdentifier
				duration := args[1].(LispDuration).DurationComponent
				note := Note{
					Pitch:    pitch,
					Duration: Duration{Components: []DurationComponent{duration}},
				}
				return LispScoreUpdate{ScoreUpdate: note}, nil
			},
		},
	)

	defn("slur",
		FunctionSignature{
			ArgumentTypes: []LispForm{LispScoreUpdate{}},
			Implementation: func(args ...LispForm) (LispForm, error) {
				scoreUpdate := args[0].(LispScoreUpdate).ScoreUpdate

				switch scoreUpdate := scoreUpdate.(type) {
				case Note:
					scoreUpdate.Slurred = true
					return LispScoreUpdate{ScoreUpdate: scoreUpdate}, nil
				default:
					return nil, fmt.Errorf(
						"only notes can be slurred. Expected Note, got: %#v", scoreUpdate,
					)
				}
			},
		},
	)
}

// LispNil is the value nil.
type LispNil struct {
	SourceContext AldaSourceContext
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (n LispNil) GetSourceContext() AldaSourceContext {
	return n.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (LispNil) JSON() *json.Container {
	return json.Object("type", "nil")
}

// TypeString implements LispForm.TypeString.
func (LispNil) TypeString() string {
	return "nil"
}

// Eval implements LispForm.Eval by returning the LispNil.
func (n LispNil) Eval() (LispForm, error) {
	return n, nil
}

// UpdateScore implements ScoreUpdate.UpdateScore by doing nothing.
func (n LispNil) UpdateScore(score *Score) error {
	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0.
func (n LispNil) DurationMs(part *Part) float64 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (n LispNil) VariableValue(score *Score) (ScoreUpdate, error) {
	return n, nil
}

// LispQuotedForm wraps a form by quoting it.
type LispQuotedForm struct {
	SourceContext AldaSourceContext
	Form          LispForm
}

// JSON implements RepresentableAsJSON.JSON.
func (qf LispQuotedForm) JSON() *json.Container {
	return json.Object("type", "quoted-form", "value", qf.Form.JSON())
}

// TypeString implements LispForm.TypeString.
func (LispQuotedForm) TypeString() string {
	return "quoted form"
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (qf LispQuotedForm) GetSourceContext() AldaSourceContext {
	return qf.SourceContext
}

// Eval implements LispForm.Eval by returning the unquoted form.
func (qf LispQuotedForm) Eval() (LispForm, error) {
	return qf.Form, nil
}

// LispSymbol is a Lisp symbol.
type LispSymbol struct {
	SourceContext AldaSourceContext
	Name          string
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (sym LispSymbol) GetSourceContext() AldaSourceContext {
	return sym.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (sym LispSymbol) JSON() *json.Container {
	return json.Object("type", "symbol", "value", sym.Name)
}

// TypeString implements LispForm.TypeString.
func (LispSymbol) TypeString() string {
	return "symbol"
}

func (sym LispSymbol) String() string {
	return "'" + sym.Name
}

// Eval implements LispForm.Eval by resolving the symbol and returning the
// corresponding value.
//
// Returns an error if the symbol cannot be resolved.
func (sym LispSymbol) Eval() (LispForm, error) {
	specialForm, hit := specialForms[sym.Name]
	if hit {
		return specialForm, nil
	}

	value, hit := environment[sym.Name]
	if hit {
		return value, nil
	}

	return nil, &AldaSourceError{
		Context: sym.SourceContext,
		Err:     fmt.Errorf("unresolvable symbol: %s", sym.Name),
	}
}

// LispNumber is a floating point number.
type LispNumber struct {
	SourceContext AldaSourceContext
	Value         float64
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (n LispNumber) GetSourceContext() AldaSourceContext {
	return n.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (n LispNumber) JSON() *json.Container {
	return json.Object("type", "number", "value", n.Value)
}

// TypeString implements LispForm.TypeString.
func (LispNumber) TypeString() string {
	return "number"
}

// Eval implements LispForm.Eval by returning the number.
func (n LispNumber) Eval() (LispForm, error) {
	return n, nil
}

// LispString is a string value.
type LispString struct {
	SourceContext AldaSourceContext
	Value         string
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (s LispString) GetSourceContext() AldaSourceContext {
	return s.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (s LispString) JSON() *json.Container {
	return json.Object("type", "string", "value", s.Value)
}

// TypeString implements LispForm.TypeString.
func (LispString) TypeString() string {
	return "string"
}

// Eval implements LispForm.Eval by returning the string.
func (s LispString) Eval() (LispForm, error) {
	return s, nil
}

// LispScoreUpdate is a ScoreUpdate value.
type LispScoreUpdate struct {
	ScoreUpdate ScoreUpdate
}

// JSON implements RepresentableAsJSON.JSON.
func (su LispScoreUpdate) JSON() *json.Container {
	return json.Object("type", "score-update", "value", su.ScoreUpdate.JSON())
}

// TypeString implements LispForm.TypeString.
func (su LispScoreUpdate) TypeString() string {
	return "score update"
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (su LispScoreUpdate) GetSourceContext() AldaSourceContext {
	return su.ScoreUpdate.GetSourceContext()
}

// Eval implements LispForm.Eval by returning the score update.
func (su LispScoreUpdate) Eval() (LispForm, error) {
	return su, nil
}

// LispPitch is a PitchIdentifier value.
type LispPitch struct {
	PitchIdentifier PitchIdentifier
}

// JSON implements RepresentableAsJSON.JSON.
func (p LispPitch) JSON() *json.Container {
	return json.Object("type", "pitch", "value", p.PitchIdentifier.JSON())
}

// TypeString implements LispForm.TypeString.
func (p LispPitch) TypeString() string {
	return "pitch"
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (LispPitch) GetSourceContext() AldaSourceContext {
	// It isn't possible to parse a LispPitch instance* from an Alda score, so
	// there is no way we can provide source context.
	//
	// *You _can_ write an S-expression that returns a LispPitch, like
	// `(pitch ...)`, but the LispPitch itself is only used internally. The thing
	// that has the source context is the LispList that represents the
	// S-expression `(pitch ...)`.
	return AldaSourceContext{}
}

// Eval implements LispForm.Eval by returning the pitch.
func (p LispPitch) Eval() (LispForm, error) {
	return p, nil
}

// LispDuration is a DurationComponent value.
type LispDuration struct {
	DurationComponent DurationComponent
}

// JSON implements RepresentableAsJSON.JSON.
func (d LispDuration) JSON() *json.Container {
	return json.Object("type", "duration", "value", d.DurationComponent.JSON())
}

// TypeString implements LispForm.TypeString.
func (d LispDuration) TypeString() string {
	return "duration"
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (LispDuration) GetSourceContext() AldaSourceContext {
	// It isn't possible to parse a LispDuration instance* from an Alda score, so
	// there is no way we can provide source context.
	//
	// *You _can_ write an S-expression that returns a LispDuration, like
	// `(duration ...)`, but the LispDuration itself is only used internally. The
	// thing that has the source context is the LispList that represents the
	// S-expression `(duration ...)`.
	return AldaSourceContext{}
}

// Eval implements LispForm.Eval by returning the pitch.
func (d LispDuration) Eval() (LispForm, error) {
	return d, nil
}

// LispList is a list of forms.
type LispList struct {
	SourceContext AldaSourceContext
	Elements      []LispForm
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (l LispList) GetSourceContext() AldaSourceContext {
	return l.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (l LispList) JSON() *json.Container {
	elements := json.Array()
	for _, element := range l.Elements {
		elements.ArrayAppend(element.JSON())
	}

	return json.Object("type", "list", "value", elements)
}

// TypeString implements LispForm.TypeString.
func (LispList) TypeString() string {
	return "list"
}

// Eval implements LispForm.Eval by treating the (unquoted) list as an
// S-expression.
//
// All forms in the list are evaluated, unless the operator is a
// special form (TODO: also add support for macros).
//
// The first form is treated as an operator and the remaining forms are treated
// as arguments.
func (l LispList) Eval() (LispForm, error) {
	sourceError := func(err error) error {
		return &AldaSourceError{
			Context: l.SourceContext,
			Err:     err,
		}
	}

	operator, err := l.Elements[0].Eval()
	if err != nil {
		return nil, err
	}

	// Handle special forms
	// TODO: consider turning this into an interface?
	switch operator.(type) {
	case LispSpecialFormQuote:
		if len(l.Elements[1:]) > 1 {
			return nil, fmt.Errorf(
				"expected 1 argument to quote, got %d", len(l.Elements[1:]),
			)
		}

		return l.Elements[1], nil
	}

	arguments := []LispForm{}
	for _, argument := range l.Elements[1:] {
		value, err := argument.Eval()
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, value)
	}

	switch operator := operator.(type) {
	case Operator:
		result, err := operator.Operate(arguments)
		if err != nil {
			return nil, sourceError(err)
		}
		return result, nil
	default:
		return nil, sourceError(
			fmt.Errorf("value is not an Operator: %#v", operator),
		)
	}
}

func unpackScoreUpdate(form LispForm) ScoreUpdate {
	switch form := form.(type) {
	case LispScoreUpdate:
		return form.ScoreUpdate
	default:
		log.Warn().
			Interface("form", form).
			Msg("S-expression result is not a ScoreUpdate.")
		return LispNil{}
	}
}

// UpdateScore implements ScoreUpdate.UpdateScore by evaluating the S-expression
// and using the resulting value to update the score.
func (l LispList) UpdateScore(score *Score) error {
	result, err := l.Eval()
	if err != nil {
		return err
	}

	return score.Update(unpackScoreUpdate(result))
}

// DurationMs implements ScoreUpdate.DurationMs by evaluating the S-expression
// and returning the duration of the resulting value.
func (l LispList) DurationMs(part *Part) float64 {
	// FIXME: We end up evaluating this a second time when UpdateScore is called.
	// This will be problematic if/when we add functions that have side effects.
	//
	// At that point, we should probably cache the evaluation result (and error)
	// so that they are simply returned on successive evaluations.
	result, err := l.Eval()

	// If there is an error during evaluation, it will be propagated through when
	// we evaluate again for UpdateScore. So we can safely ignore it here and fall
	// back to a duration of 0.
	if err != nil {
		return 0
	}

	return unpackScoreUpdate(result).DurationMs(part)
}

// VariableValue implements ScoreUpdate.VariableValue by evaluating the
// S-expression and capturing the value of the result.
func (l LispList) VariableValue(score *Score) (ScoreUpdate, error) {
	// FIXME: We end up evaluating this a second time when UpdateScore is called.
	// This will be problematic if/when we add functions that have side effects.
	//
	// At that point, we should probably cache the evaluation result (and error)
	// so that they are simply returned on successive evaluations.
	result, err := l.Eval()

	if err != nil {
		return nil, err
	}

	return unpackScoreUpdate(result).VariableValue(score)
}

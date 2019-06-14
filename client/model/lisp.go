package model

import (
	"fmt"
	"reflect"
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
		return 0, fmt.Errorf("Expected value to be positive: %f", value)
	}

	return value, nil
}

func wholeNumber(form LispForm) (int32, error) {
	value := form.(LispNumber).Value

	if value != float32(int32(value)) {
		return 0, fmt.Errorf("Expected a whole number: %f", value)
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
		// TODO: implement 2-arity that allows expressing tempo in terms of a note
		// length other than quarter note
		//
		// https://github.com/alda-lang/alda-core/blob/e27b1d07224b5951fca63277374e957ed399c8f5/src/alda/lisp/attributes.clj#L92-L93
		//
		// May need to make two versions of this, one that takes a number like 2
		// (half) and one that takes a string like "4." (dotted quarter)
	)

	// The percentage of a note that is heard. Used to put a little space between
	// notes.
	//
	// e.g. with a quantization value of 90%, a note that would otherwise last 500
	// ms will be quantized to last 450 ms. The resulting note event will have a
	// duration of 450 ms, and the next event will be set to occur in 500 ms.
	// defattribute([]string{"quantization", "quantize", "quant"},
	// 	attributeFunctionSignature{
	// 		argumentTypes: []LispForm{LispNumber{}},
	// 		implementation: func(args ...LispForm) (PartUpdate, error) {
	// 			percentage, err := percentage(args[0])
	// 			if err != nil {
	// 				return nil, err
	// 			}
	// 			return QuantizationSet{Quantization: percentage}, nil
	// 		},
	// 	},
	// )

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
		return nil, fmt.Errorf("Unresolved symbol: %s", sym.Name)
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

func (l LispList) updateScore(score *Score) error {
	result, err := l.Eval()
	if err != nil {
		return err
	}

	switch result.(type) {
	case LispScoreUpdate:
		return result.(LispScoreUpdate).ScoreUpdate.updateScore(score)
	case ScoreUpdate:
		return result.(ScoreUpdate).updateScore(score)
	default:
		log.Warn().
			Interface("result", result).
			Msg("S-expression result is not a ScoreUpdate.")
		return nil
	}
}

package repl

import (
	"fmt"
	"reflect"
)

var typeString = reflect.TypeOf("")

type requestValidationRule interface {
	validate(request map[string]interface{}) []string
}

// Given a request and a list of validation rules, returns a list of validation
// error messages.
//
// If there were no problems, then the list of error messages is empty.
func validateRequest(
	request map[string]interface{}, rules ...requestValidationRule,
) []string {
	errors := []string{}

	for _, rule := range rules {
		errors = append(errors, (rule.validate(request))...)
	}

	return errors
}

type requestFieldSpec struct {
	name      string
	valueType reflect.Type
	required  bool
}

func (spec requestFieldSpec) validate(req map[string]interface{}) []string {
	value, present := req[spec.name]

	if spec.required && !present {
		return []string{"Request field missing: " + spec.name}
	}

	actualType := reflect.TypeOf(value)
	if present && actualType != spec.valueType {
		return []string{
			fmt.Sprintf(
				"Expected \"%s\" to be of type `%s`, but it was of type `%s`.",
				spec.name,
				spec.valueType,
				actualType,
			),
		}
	}

	return []string{}
}

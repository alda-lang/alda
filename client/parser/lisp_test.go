package parser

import (
	"testing"

	"alda.io/client/model"
	_ "alda.io/client/testing"
)

func lispSymbol(name string) model.LispSymbol {
	return model.LispSymbol{Name: name}
}

func lispNumber(value float32) model.LispNumber {
	return model.LispNumber{Value: value}
}

func lispString(value string) model.LispString {
	return model.LispString{Value: value}
}

func lispList(elements ...model.LispForm) model.LispList {
	return model.LispList{Elements: elements}
}

func lispQuotedForm(form model.LispForm) model.LispQuotedForm {
	return model.LispQuotedForm{Form: form}
}

func lispQuotedList(elements ...model.LispForm) model.LispQuotedForm {
	return lispQuotedForm(lispList(elements...))
}

func TestLisp(t *testing.T) {
	executeParseTestCases(
		t,
		parseTestCase{
			label: "attribute change with number value",
			given: "(volume 50)",
			expect: []model.ScoreUpdate{
				lispList(lispSymbol("volume"), lispNumber(50)),
			},
		},
		parseTestCase{
			label: "attribute change with string value",
			given: `(key-signature "f+ c+ g+")`,
			expect: []model.ScoreUpdate{
				lispList(lispSymbol("key-signature"), lispString("f+ c+ g+")),
			},
		},
		parseTestCase{
			label: "global attribute change",
			given: "(tempo! 200)",
			expect: []model.ScoreUpdate{
				lispList(lispSymbol("tempo!"), lispNumber(200)),
			},
		},
		parseTestCase{
			label: "attribute change with quoted list argument",
			given: "(key-sig '(a major))",
			expect: []model.ScoreUpdate{
				lispList(
					lispSymbol("key-sig"),
					lispQuotedList(lispSymbol("a"), lispSymbol("major")),
				),
			},
		},
		parseTestCase{
			label: "attribute change with quoted nested list argument",
			given: "(key-signature '(e (flat) b (flat)))",
			expect: []model.ScoreUpdate{
				lispList(
					lispSymbol("key-signature"),
					lispQuotedList(
						lispSymbol("e"),
						lispList(lispSymbol("flat")),
						lispSymbol("b"),
						lispList(lispSymbol("flat")),
					),
				),
			},
		},
	)
}

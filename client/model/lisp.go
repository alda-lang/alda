package model

import (
	"errors"
)

// Alda includes a minimal Lisp implementation as a subset of the language, in
// order to facilitate adding new features to the language without accumulating
// syntax.
//
// The Lisp might also facilitate some sort of metaprogramming, but the
// preferred approach is to drive Alda from a Turing-complete programming
// language like Clojure. (See: https://github.com/daveyarwood/alda-clj)

type LispForm interface{}

type LispQuotedForm struct {
	Form LispForm
}

type LispSymbol struct {
	Name string
}

type LispNumber struct {
	Value float32
}

type LispString struct {
	Value string
}

type LispList struct {
	Elements []LispForm
}

func (LispList) updateScore(score *Score) error {
	return errors.New("LispList.updateScore not implemented")
}

package model

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

// Alda includes a minimal Lisp implementation as a subset of the language, in
// order to facilitate adding new features to the language without accumulating
// syntax.
//
// The Lisp might also facilitate some sort of metaprogramming, but the
// preferred approach is to drive Alda from a Turing-complete programming
// language like Clojure. (See: https://github.com/daveyarwood/alda-clj)

// A Sexp is a Lisp S-expression that is evaluated when the score is compiled.
type LispSexp struct {
	Operator  LispForm
	Arguments []LispForm
}

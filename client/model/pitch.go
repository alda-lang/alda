package model

import (
	"fmt"
)

type NoteLetter int

const (
	A NoteLetter = iota
	B
	C
	D
	E
	F
	G
)

func NewNoteLetter(letter rune) (NoteLetter, error) {
	switch letter {
	case 'a':
		return A, nil
	case 'b':
		return B, nil
	case 'c':
		return C, nil
	case 'd':
		return D, nil
	case 'e':
		return E, nil
	case 'f':
		return F, nil
	case 'g':
		return G, nil
	default:
		return -1, fmt.Errorf("Invalid note letter: %c", letter)
	}
}

type Accidental int

const (
	Flat Accidental = iota
	Natural
	Sharp
)

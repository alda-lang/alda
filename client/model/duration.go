package model

type NoteLength struct {
	Denominator float32
	Dots        int32
}

type NoteLengthMs struct {
	Quantity float32
}

type DurationComponent interface{}

type Duration struct {
	Components []DurationComponent
	// When the duration of a note is slurred, it means there is minimal space
	// between that note and the next.
	Slurred bool
}

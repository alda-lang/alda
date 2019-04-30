package model

type NoteLength struct {
	Denominator int32
	Dots        int32
}

type NoteLengthMs struct {
	Quantity int32
}

type NoteLengthS struct {
	Quantity int32
}

type DurationComponent interface{}

type Duration struct {
	Components []DurationComponent
	// When the duration of a note is slurred, it means there is minimal space
	// between that note and the next.
	Slurred bool
}

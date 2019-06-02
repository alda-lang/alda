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

// TempoRole describes the relationship a part has with the global tempo of a
// score.
type TempoRole int

const (
	// TempoRoleUnspecified means that the part has no special relationship with
	// the global tempo.
	TempoRoleUnspecified TempoRole = 0
	// TempoRoleMaster means that any tempo changes in the part apply to the score
	// as a whole.
	TempoRoleMaster TempoRole = 1
)

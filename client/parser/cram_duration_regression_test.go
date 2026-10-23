package parser

import (
	"math"
	"testing"

	"alda.io/client/model"
)

func TestNestedCramDurationInheritance(t *testing.T) {
	ast, err := Parse("synthetic cram", "piano: {{c}2 d}1")
	if err != nil {
		t.Fatal(err)
	}
	updates, err := ast.Updates()
	if err != nil {
		t.Fatal(err)
	}
	score := model.NewScore()
	if err := score.Update(updates...); err != nil {
		t.Fatal(err)
	}
	if len(score.Events) != 2 {
		t.Fatalf("expected two notes, got %d", len(score.Events))
	}
	for i, event := range score.Events {
		note := event.(model.NoteEvent)
		if math.Abs(note.Offset-float64(i)*1000) > 0.001 || math.Abs(note.Duration-1000) > 0.001 {
			t.Errorf("note %d: expected offset %d and duration 1000, got offset %v and duration %v",
				i, i*1000, note.Offset, note.Duration)
		}
	}
	if offset := score.CurrentParts[0].CurrentOffset; math.Abs(offset-2000) > 0.001 {
		t.Errorf("expected cram to occupy 2000 ms, got %v", offset)
	}
}

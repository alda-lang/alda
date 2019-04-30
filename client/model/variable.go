package model

type VariableDefinition struct {
	VariableName string
	Events       []ScoreUpdate
}

type VariableReference struct {
	VariableName string
}

package parser

import (
	"bytes"
	"math"
	"testing"

	"alda.io/client/model"
	"github.com/go-test/deep"
)

// parseTestCase models a test of the `parser` module
// parseTestCase also tests Score Updating which tests the `model` module
// Perhaps this should live in another package that consumes both `parser` and
// `model`, but for now we keep it here
type parseTestCase struct {
	label            string
	given            string
	expectUpdates    []model.ScoreUpdate // optional
	expectAST        *ASTNode            // optional
	scoreApplyOptOut bool                // optional
}

// trimASTComments recursively trims all comments from an ASTNode
func trimASTComments(node *ASTNode) {
	node.Doc = ""
	node.Comment = ""
	for i := 0; i < len(node.Children); i++ {
		trimASTComments(&node.Children[i])
	}
}

// executeParseTestCases parses each test case's given string of Alda code and
// tests the parser, scanner, ASTNode.Updates, GenerateASTFromScoreUpdates, and
// Score Updating (with the ability to opt out)
func executeParseTestCases(t *testing.T, testCases ...parseTestCase) {
	for _, testCase := range testCases {
		deep.MaxDepth = math.MaxInt32

		// Test parser
		actualAST, err := Parse(
			// We suppress source context to facilitate deep diff comparison
			testCase.label, testCase.given, SuppressSourceContext)
		if err != nil {
			t.Errorf("%v\n", err)
			return
		}
		if testCase.expectAST != nil {
			diff := deep.Equal(testCase.expectAST, actualAST)
			if diff != nil {
				t.Error(testCase.label)
				for _, diffItem := range diff {
					t.Errorf("%v", diffItem)
				}
			}
		}

		// Test ASTNode -> ScoreUpdate
		actualUpdates, err := actualAST.Updates()
		if err != nil {
			t.Errorf("%v\n", err)
			return
		}
		if testCase.expectUpdates != nil {
			diff := deep.Equal(testCase.expectUpdates, actualUpdates)
			if diff != nil {
				t.Error(testCase.label)
				for _, diffItem := range diff {
					t.Errorf("%v", diffItem)
				}
			}
		}

		// Test code generation by ensuring round-trip generated AST is the same
		generatedAST, err := GenerateASTFromScoreUpdates(actualUpdates)
		if err != nil {
			t.Errorf("%v\n", err)
			return
		}
		if diff := deep.Equal(actualAST, generatedAST); diff != nil {
			t.Error(testCase.label)
			for _, diffItem := range diff {
				t.Errorf("%v", diffItem)
			}
		}

		// Test formatter by ensuring round-trip formatted + parsed AST is same
		buffer := bytes.Buffer{}
		err = FormatASTToCode(actualAST, &buffer)
		if err != nil {
			t.Errorf("%v\n", err)
			return
		}
		formattedAST, err := Parse(
			// The newly formatted file will have different source context's
			testCase.label, buffer.String(), SuppressSourceContext)
		if diff := deep.Equal(actualAST, formattedAST); diff != nil {
			t.Error(testCase.label)
			for _, diffItem := range diff {
				t.Errorf("%v", diffItem)
			}
		}

		if !testCase.scoreApplyOptOut {
			score := model.NewScore()
			err = score.Update(actualUpdates...)
			if err != nil {
				t.Errorf(testCase.label)
				t.Errorf(err.Error())
			}
		}
	}
}

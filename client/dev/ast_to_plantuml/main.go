package main

// Reads Alda AST JSON from stdin and prints PlantUML to stdout.
//
// Example workflow to turn an Alda score file into an SVG image of the AST:
//
//   alda parse -f my-score.alda -o ast \
//     | go run dev/ast_to_plantuml/main.go \
//     | java -jar /path/to/plantuml-1.2021.16.jar -pipe -tsvg \
//     > my-score.svg

import (
	"bytes"
	"fmt"
	"strings"

	"alda.io/client/json"
	"alda.io/client/system"
	"github.com/google/uuid"
)

var buffer bytes.Buffer

func randomID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func writeHeader() {
	buffer.WriteString("@startuml\n")
}

func writeFooter() {
	buffer.WriteString("@enduml\n")
}

func writeNodeDefinition(node *json.Container, nodeID string) {
	nodeType := node.Search("type").Data().(string)

	buffer.WriteString(fmt.Sprintf("usecase %s as \"%s", nodeID, nodeType))

	sc, ok := node.Search("source-context").Data().(map[string]interface{})
	if ok {
		sourceContext := fmt.Sprintf(
			"[%d:%d]",
			int(sc["line"].(float64)),
			int(sc["column"].(float64)),
		)

		buffer.WriteString(fmt.Sprintf(" //%s//", sourceContext))
	}

	if literal := node.Search("literal").Data(); literal != nil {
		var literalString string

		switch literal.(type) {
		case map[string]interface{}:
			// HACK to make e.g. `map[string]interface{}{denominator: 4, dots: 2}`
			// display as user-friendly JSON instead: `{"denominator":4,"dots":2}`
			literalString = fmt.Sprintf("%s", json.ToJSON(literal))
		case string:
			literalString = literal.(string)
		default:
			literalString = fmt.Sprintf("%#v", literal)
		}

		buffer.WriteString(fmt.Sprintf("\n--\n\"\"%s\"\"", literalString))
	}

	buffer.WriteString("\"\n")
}

func writeNodeLineage(nodeID string, childID string) {
	buffer.WriteString(fmt.Sprintf("%s --> %s\n", nodeID, childID))
}

func writeNodeUML(node *json.Container, nodeID string) {
	var recursivelyWriteNodes func(node *json.Container, nodeID string)
	recursivelyWriteNodes = func(node *json.Container, nodeID string) {
		children := node.Search("children")

		childrenByID := map[string]*json.Container{}

		for _, child := range children.Children() {
			childID := randomID()
			childrenByID[childID] = child
			writeNodeDefinition(child, childID)
			writeNodeLineage(nodeID, childID)
		}

		for childID, child := range childrenByID {
			recursivelyWriteNodes(child, childID)
		}
	}

	writeNodeDefinition(node, nodeID)

	recursivelyWriteNodes(node, nodeID)
}

func main() {
	input, err := system.ReadStdin()
	if err != nil {
		panic(err)
	}

	node, err := json.ParseJSON(input)
	if err != nil {
		panic(err)
	}

	writeHeader()
	writeNodeUML(node, randomID())
	writeFooter()

	fmt.Println(buffer.String())
}

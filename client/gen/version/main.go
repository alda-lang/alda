package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	dir, err := os.Getwd()
	check(err)

	// Read the current version from the top-level VERSION file in the Alda repo.
	versionFile := filepath.Join(filepath.Dir(dir), "VERSION")
	contents, err := ioutil.ReadFile(versionFile)
	check(err)
	version := strings.TrimSpace(string(contents))

	// Create generated/version.go
	outputFilepath := filepath.Join(dir, "generated", "version.go")
	err = os.MkdirAll(filepath.Dir(outputFilepath), os.ModePerm)
	check(err)
	outputFile, err := os.Create(outputFilepath)
	check(err)
	defer outputFile.Close()

	// Generate source code into generated/version.go.
	tmpl, err := template.New("generated/version.go").
		Parse(`package generated

// ClientVersion is the version of the Alda client.
const ClientVersion = "{{.}}"`)

	err = tmpl.Execute(outputFile, version)
	check(err)
}

package main

import (
	"os"

	"alda.io/client/cmd"
)

// This directive makes it so that when you run `go generate`, it runs some Go
// code (see: gen/version/main.go) that spits out a file `generated/version.go`
// where the constant ClientVersion is defined. The `generated` directory is
// gitignored so that we don't have to worry about keeping its contents in sync
// (in version control) with the top-level VERSION file in the alda repository.
// The idea is that the top-level VERSION file is the only place where we
// specify the version of Alda, and any code in the client and player that need
// to refer to the version should do so in a way that ultimately uses the
// top-level VERSION file as the source of truth.
//
// tl;dr: Always run `go generate` before `go build`, otherwise there will be
// generated code missing and the build won't compile. Or just use `bin/build`
// instead, because it will handle this for you.
//
//go:generate go run gen/version/main.go

func main() {
	if err := cmd.Execute(); err != nil {
		// Ordinarily, we would also print the error message here, but Cobra is
		// already doing that before it returns the error, and we don't want to
		// print the same message twice.
		os.Exit(1)
	}
}

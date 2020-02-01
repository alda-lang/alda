package main

import (
	"fmt"
	"os"

	"alda.io/client/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

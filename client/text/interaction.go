package text

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func PromptForConfirmation(prompt string, defaultToYes bool) bool {
	options := "[yN]"
	if defaultToYes {
		options = "[Yn]"
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s %s ", prompt, options)

		response, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		switch strings.ToLower(strings.TrimSpace(response)) {
		case "":
			return defaultToYes
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
	}
}

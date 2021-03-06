package system

import (
	"testing"

	_ "alda.io/client/testing"
)

func TestIsRenamedExecutable(t *testing.T) {
	for _, basename := range []string{
		"alda.1615061099.2081.old",
		"alda-player.1615061099.2081.old",
	} {
		if !IsRenamedExecutable(basename) {
			t.Errorf("%s should be considered a renamed executable", basename)
		}
	}

	for _, basename := range []string{
		"alda",
		"alda-player",
		"alda.old",
		"alda-player.old",
		"alda-old",
		"alda-player-old",
		"alda2",
		"README.md",
		"2020-taxes-IMPORTANT.xslx",
	} {
		if IsRenamedExecutable(basename) {
			t.Errorf("%s should not be considered a renamed executable", basename)
		}
	}
}

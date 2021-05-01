package system

import (
	"testing"

	_ "alda.io/client/testing"
)

func TestIsRenamedExecutable(t *testing.T) {
	for _, basename := range []string{
		"alda.1615061099.2081.old",
		"alda-player.1615061099.2081.old",
		"alda.exe.1615061099.2081.old",
		"alda-player.exe.1615061099.2081.old",
	} {
		if !IsRenamedExecutable(basename) {
			t.Errorf("%s should be considered a renamed executable", basename)
		}
	}

	for _, basename := range []string{
		"alda",
		"alda-player",
		"alda.exe",
		"alda-player.exe",
		"alda.old",
		"alda-player.old",
		"alda.exe.old",
		"alda-player.exe.old",
		"alda.old.exe",
		"alda-player.old.exe",
		"alda-old",
		"alda-player-old",
		"alda-old.exe",
		"alda-player-old.exe",
		"alda2",
		"README.md",
		"2020-taxes-IMPORTANT.xslx",
	} {
		if IsRenamedExecutable(basename) {
			t.Errorf("%s should not be considered a renamed executable", basename)
		}
	}
}

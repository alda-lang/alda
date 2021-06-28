package color

import (
	"os"

	auroraLib "github.com/logrusorgru/aurora"
)

var Aurora auroraLib.Aurora

func init() {
	// HACK: Ideally, aurora would support NO_COLOR, but at least they give us a
	// config option so that we can disable color manually.
	noColor := len(os.Getenv("NO_COLOR")) > 0
	Aurora = auroraLib.NewAurora(!noColor)
}

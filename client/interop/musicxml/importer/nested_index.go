package importer

import "math"

// nestedIndex identifies a model.ScoreUpdate in a []model.ScoreUpdate
// nestedIndex is designed to work with slices as built by musicXMLImporter
// indices (in order) represent the indices within slices of score updates
// Multiple indices means we enter the sequences for chords, repeats,
// repetitions, or event sequences (see getNestedUpdates)
// Note that since the musicXMLImporter imports into a model.EventSequence
// any valid nestedIndex will always have the first index = 0
type nestedIndex struct {
	indices []int
}

func (ni nestedIndex) valid(importer *musicXMLImporter) bool {
	if len(ni.indices) == 0 {
		return false
	}
	return importer.getAt(ni) != nil
}

func (ni nestedIndex) last() int {
	return ni.indices[len(ni.indices)-1]
}

// lastDiff returns the diff between indices at the deepest recursive level
func lastDiff(minuend nestedIndex, subtrahend nestedIndex) int {
	index := math.Min(
		float64(len(minuend.indices)),
		float64(len(subtrahend.indices)),
	) - 1
	return minuend.indices[int(index)] - subtrahend.indices[int(index)]
}

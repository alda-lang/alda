package importer

// postProcess produces an equivalent slice of updates with more idiomatic Alda
func postProcess(importer *musicXMLImporter) {
	removeRedundantRests(importer)
	removeRedundantDurations(importer)
	removeRedundantAccidentals(importer)
}

func removeRedundantRests(importer *musicXMLImporter) {
	// We pad rests into empty parts after every measure
	// We do this to manage barlines
	// Then parts may have trailing rests
	// We can optionally get rid of these
}

func removeRedundantDurations(importer *musicXMLImporter) {
	// When importing, we keep all duration attributes for notes and rests
	// This is to facilitate tied notes
	// We remove these in postprocessing to align with idiomatic Alda
	// Note that we cannot remove durations for tied notes (e.x. c4 d4~4)

	// Iterate through and maintain current duration
	// Remove durations as applicable
}

func removeRedundantAccidentals(importer *musicXMLImporter) {
	// With MusicXML, the key signature is only for appearance, and any altered
	// notes have alter tags
	// With Alda, the key signature applies the accidentals while playing
	// So we can remove these unnecessary alter tags in our generated Alda

	// Iterate through and maintain current key signature
	// Remove accidentals as applicable
}

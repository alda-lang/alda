package model

import (
	"fmt"
	"sort"
)

// AttributeUpdate updates the value of an attribute for all current parts.
type AttributeUpdate struct {
	PartUpdate PartUpdate
}

// UpdateScore implements ScoreUpdate.UpdateScore by updating an attribute value
// for all current parts.
func (au AttributeUpdate) UpdateScore(score *Score) error {
	for _, part := range score.CurrentParts {
		au.PartUpdate.updatePart(part)
	}

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since an
// attribute update is conceptually instantaneous.
func (au AttributeUpdate) DurationMs(part *Part) float32 {
	return 0
}

// GlobalAttributes are attribute updates to be applied at specific points of
// time in a score.
//
// A common example is a global tempo change, e.g. at 5000ms into the score, the
// tempo should be set to 127 bpm for all parts.
type GlobalAttributes struct {
	// We store both the map of offset => updates and an ordered list of offsets,
	// so that we can easily traverse the map in order.
	itinerary map[OffsetMs][]PartUpdate
	offsets   []OffsetMs
}

// NewGlobalAttributes returns an initialized GlobalAttributes structure.
func NewGlobalAttributes() *GlobalAttributes {
	return &GlobalAttributes{
		itinerary: map[OffsetMs][]PartUpdate{},
		offsets:   []OffsetMs{},
	}
}

// Record adds attribute changes at a particular offset to the itinerary of
// global attribute changes.
func (ga *GlobalAttributes) Record(offset OffsetMs, update PartUpdate) {
	_, hit := ga.itinerary[offset]
	if !hit {
		ga.itinerary[offset] = []PartUpdate{}
	}

	ga.itinerary[offset] = append(ga.itinerary[offset], update)

	if !hit {
		ga.offsets = append(ga.offsets, offset)
		sort.Float64s(ga.offsets)
	}
}

// InWindow returns the list of global attribute updates that fall within the
// window between `startOffset` and `endOffset`.
func (ga GlobalAttributes) InWindow(
	startOffset OffsetMs, endOffset OffsetMs,
) []PartUpdate {
	updates := []PartUpdate{}

	for _, offset := range ga.offsets {
		if offset > endOffset {
			return updates
		}

		if offset > startOffset {
			updates = append(updates, ga.itinerary[offset]...)
		}
	}

	return updates
}

// ApplyGlobalAttributes uses a score's itinerary of global attribute updates to
// update the attributes of all current parts.
//
// Any global attribute updates registered at an offset that is between a part's
// LastOffset and CurrentOffset are applied.
func (score *Score) ApplyGlobalAttributes() {
	for _, part := range score.CurrentParts {
		for _, update := range score.GlobalAttributes.InWindow(
			part.LastOffset, part.CurrentOffset,
		) {
			for _, part := range score.CurrentParts {
				update.updatePart(part)
			}
		}
	}
}

// GlobalAttributeUpdate updates the value of an attribute for all parts.
type GlobalAttributeUpdate struct {
	PartUpdate PartUpdate
}

// UpdateScore implements ScoreUpdate.UpdateScore by recording that at a point
// in time, an attribute update should be applied for all parts.
//
// The attribute is also immediately applied to all parts.
func (gau GlobalAttributeUpdate) UpdateScore(score *Score) error {
	// Record this attribute update in the record of global attributes.
	var offset OffsetMs
	switch len(score.CurrentParts) {
	case 0:
		offset = 0
	case 1:
		offset = score.CurrentParts[0].CurrentOffset
	default:
		offset = score.CurrentParts[0].CurrentOffset
		for _, part := range score.CurrentParts[1:] {
			if part.CurrentOffset != offset {
				return fmt.Errorf(
					"can't set global attribute %#v; offset unclear. There are "+
						"multiple current parts with different offsets: %#v",
					gau,
					score.CurrentParts,
				)
			}
		}
	}

	score.GlobalAttributes.Record(offset, gau.PartUpdate)

	// Immediately apply the attribute update to the current parts.
	for _, part := range score.CurrentParts {
		gau.PartUpdate.updatePart(part)
	}

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since an
// attribute update is conceptually instantaneous.
func (gau GlobalAttributeUpdate) DurationMs(part *Part) float32 {
	return 0
}

// TempoSet sets the tempo of all active parts.
type TempoSet struct {
	Tempo float32
}

func (ts TempoSet) updatePart(part *Part) {
	part.Tempo = ts.Tempo
}

// MetricModulation sets the tempo of all active parts, defining the tempo as a
// ratio of new tempo : old tempo.
type MetricModulation struct {
	Ratio float32
}

func (mm MetricModulation) updatePart(part *Part) {
	part.Tempo *= mm.Ratio
}

// OctaveSet sets the octave of all active parts.
type OctaveSet struct {
	OctaveNumber int32
}

func (os OctaveSet) updatePart(part *Part) {
	part.Octave = os.OctaveNumber
}

// OctaveUp increments the octave of all active parts.
type OctaveUp struct{}

func (OctaveUp) updatePart(part *Part) {
	part.Octave++
}

// OctaveDown decrements the octave of all active parts.
type OctaveDown struct{}

func (OctaveDown) updatePart(part *Part) {
	part.Octave--
}

// VolumeSet sets the volume of all active parts.
type VolumeSet struct {
	Volume float32
}

func (vs VolumeSet) updatePart(part *Part) {
	part.Volume = vs.Volume
}

// TrackVolumeSet sets the track volume of all active parts.
type TrackVolumeSet struct {
	TrackVolume float32
}

func (tvs TrackVolumeSet) updatePart(part *Part) {
	part.TrackVolume = tvs.TrackVolume
}

// PanningSet sets the panning of all active parts.
type PanningSet struct {
	Panning float32
}

func (ps PanningSet) updatePart(part *Part) {
	part.Panning = ps.Panning
}

// QuantizationSet sets the quantization of all active parts.
type QuantizationSet struct {
	Quantization float32
}

func (qs QuantizationSet) updatePart(part *Part) {
	part.Quantization = qs.Quantization
}

// DurationSet sets the quantization of all active parts.
type DurationSet struct {
	Duration Duration
}

func (ds DurationSet) updatePart(part *Part) {
	part.Duration = ds.Duration
}

// KeySignatureSet sets the key signature of all active parts.
type KeySignatureSet struct {
	KeySignature KeySignature
}

func (kss KeySignatureSet) updatePart(part *Part) {
	part.KeySignature = kss.KeySignature
}

// TranspositionSet sets the transposition of all active parts.
type TranspositionSet struct {
	Semitones int32
}

func (ts TranspositionSet) updatePart(part *Part) {
	part.Transposition = ts.Semitones
}

// ReferencePitchSet sets the reference pitch of all active parts. The reference
// pitch is represented as the frequency of A4.
type ReferencePitchSet struct {
	Frequency float32
}

func (rps ReferencePitchSet) updatePart(part *Part) {
	part.ReferencePitch = rps.Frequency
}

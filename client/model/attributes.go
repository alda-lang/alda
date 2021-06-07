package model

import (
	"fmt"
	"reflect"
	"sort"

	"alda.io/client/json"
	log "alda.io/client/logging"
)

// AttributeUpdate updates the value of an attribute for all current parts.
type AttributeUpdate struct {
	SourceContext AldaSourceContext
	PartUpdate    PartUpdate
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (au AttributeUpdate) GetSourceContext() AldaSourceContext {
	return au.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (au AttributeUpdate) JSON() *json.Container {
	object := au.PartUpdate.JSON()
	object.Set("attribute-update", "type")
	return object
}

// UpdateScore implements ScoreUpdate.UpdateScore by updating an attribute value
// for all current parts.
func (au AttributeUpdate) UpdateScore(score *Score) error {
	for _, part := range score.CurrentParts {
		au.PartUpdate.updatePart(part, false)
		// Here, we record that this local (part-specific) attribute was updated.
		// This is so that we can track the case where a local attribute change is
		// applied at the exact same time as a global attribute change, and we want
		// the local attribute change to take precedence.
		part.localAttributeOverride = au.PartUpdate
	}

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since an
// attribute update is conceptually instantaneous.
func (au AttributeUpdate) DurationMs(part *Part) float64 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (au AttributeUpdate) VariableValue(score *Score) (ScoreUpdate, error) {
	return au, nil
}

// GlobalAttributes are attribute updates to be applied at specific points of
// time in a score.
//
// A common example is a global tempo change, e.g. at 5000ms into the score, the
// tempo should be set to 127 bpm for all parts.
type GlobalAttributes struct {
	// We store both the map of offset => updates and an ordered list of offsets,
	// so that we can easily traverse the map in order.
	itinerary map[float64][]PartUpdate
	offsets   []float64
}

// JSON implements RepresentableAsJSON.JSON.
func (ga *GlobalAttributes) JSON() *json.Container {
	object := json.Object()
	for offset, updates := range ga.itinerary {
		updatesArray := json.Array()
		for _, update := range updates {
			updatesArray.ArrayAppend(update.JSON())
		}

		object.Set(updatesArray, fmt.Sprintf("%f", offset))
	}

	return object
}

// NewGlobalAttributes returns an initialized GlobalAttributes structure.
func NewGlobalAttributes() *GlobalAttributes {
	return &GlobalAttributes{
		itinerary: map[float64][]PartUpdate{},
		offsets:   []float64{},
	}
}

// Record adds attribute changes at a particular offset to the itinerary of
// global attribute changes.
func (ga *GlobalAttributes) Record(offset float64, update PartUpdate) {
	log.Debug().
		Float64("offset", offset).
		Interface("update", update).
		Msg("Recording global attribute update.")

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
	startOffset float64, endOffset float64,
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
			if reflect.TypeOf(part.localAttributeOverride) == reflect.TypeOf(update) {
				log.Debug().
					Str("part", part.Name).
					Interface("globalUpdate", update).
					Interface("localUpdate", part.localAttributeOverride).
					Msg("Skipping global attribute update. " +
						"Overridden by local attribute update.")
			} else {
				log.Debug().
					Str("part", part.Name).
					Interface("update", update).
					Msg("Applying global attribute update.")

				update.updatePart(part, true)
			}
		}

		part.localAttributeOverride = nil
	}
}

// GlobalAttributeUpdate updates the value of an attribute for all parts.
type GlobalAttributeUpdate struct {
	SourceContext AldaSourceContext
	PartUpdate    PartUpdate
}

// GetSourceContext implements HasSourceContext.GetSourceContext.
func (gau GlobalAttributeUpdate) GetSourceContext() AldaSourceContext {
	return gau.SourceContext
}

// JSON implements RepresentableAsJSON.JSON.
func (gau GlobalAttributeUpdate) JSON() *json.Container {
	object := gau.PartUpdate.JSON()
	object.Set("global-attribute-update", "type")
	return object
}

// UpdateScore implements ScoreUpdate.UpdateScore by recording that at a point
// in time, an attribute update should be applied for all parts.
//
// The attribute is also immediately applied to all parts.
func (gau GlobalAttributeUpdate) UpdateScore(score *Score) error {
	// Record this attribute update in the record of global attributes.
	var offset float64
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
					"can't set global attribute; there are multiple current parts with " +
						"different offsets",
				)
			}
		}
	}

	score.GlobalAttributes.Record(offset, gau.PartUpdate)

	return nil
}

// DurationMs implements ScoreUpdate.DurationMs by returning 0, since an
// attribute update is conceptually instantaneous.
func (gau GlobalAttributeUpdate) DurationMs(part *Part) float64 {
	return 0
}

// VariableValue implements ScoreUpdate.VariableValue.
func (gau GlobalAttributeUpdate) VariableValue(
	score *Score,
) (ScoreUpdate, error) {
	return gau, nil
}

// TempoSet sets the tempo of all active parts.
type TempoSet struct {
	Tempo float64
}

// JSON implements RepresentableAsJSON.JSON.
func (ts TempoSet) JSON() *json.Container {
	return json.Object("attribute", "tempo", "value", ts.Tempo)
}

func (ts TempoSet) updatePart(part *Part, globalUpdate bool) {
	part.Tempo = ts.Tempo

	// Global updates are recorded separately, and we would end up getting
	// incorrect results anyway if we recorded the tempo at the part's current
	// offset, because it might be different than the offset at which the global
	// attribute change was actually placed.
	if !globalUpdate {
		part.RecordTempoValue()
	}
}

// MetricModulation sets the tempo of all active parts, defining the tempo as a
// ratio of new tempo : old tempo.
type MetricModulation struct {
	Ratio float64
}

// JSON implements RepresentableAsJSON.JSON.
func (mm MetricModulation) JSON() *json.Container {
	return json.Object(
		"attribute", "tempo",
		"value", json.Object("ratio", mm.Ratio),
	)
}

func (mm MetricModulation) updatePart(part *Part, globalUpdate bool) {
	part.Tempo *= mm.Ratio

	// Global updates are recorded separately, and we would end up getting
	// incorrect results anyway if we recorded the tempo at the part's current
	// offset, because it might be different than the offset at which the global
	// attribute change was actually placed.
	if !globalUpdate {
		part.RecordTempoValue()
	}
}

// OctaveSet sets the octave of all active parts.
type OctaveSet struct {
	OctaveNumber int32
}

// JSON implements RepresentableAsJSON.JSON.
func (os OctaveSet) JSON() *json.Container {
	return json.Object("attribute", "octave", "value", os.OctaveNumber)
}

func (os OctaveSet) updatePart(part *Part, globalUpdate bool) {
	part.Octave = os.OctaveNumber
}

// OctaveUp increments the octave of all active parts.
type OctaveUp struct{}

// JSON implements RepresentableAsJSON.JSON.
func (os OctaveUp) JSON() *json.Container {
	return json.Object("attribute", "octave", "value", "up")
}

func (OctaveUp) updatePart(part *Part, globalUpdate bool) {
	part.Octave++
}

// OctaveDown decrements the octave of all active parts.
type OctaveDown struct{}

// JSON implements RepresentableAsJSON.JSON.
func (os OctaveDown) JSON() *json.Container {
	return json.Object("attribute", "octave", "value", "down")
}

func (OctaveDown) updatePart(part *Part, globalUpdate bool) {
	part.Octave--
}

// VolumeSet sets the volume of all active parts.
type VolumeSet struct {
	Volume float64
}

// JSON implements RepresentableAsJSON.JSON.
func (vs VolumeSet) JSON() *json.Container {
	return json.Object("attribute", "volume", "value", vs.Volume)
}

func (vs VolumeSet) updatePart(part *Part, globalUpdate bool) {
	part.Volume = vs.Volume
}

// TrackVolumeSet sets the track volume of all active parts.
type TrackVolumeSet struct {
	TrackVolume float64
}

// JSON implements RepresentableAsJSON.JSON.
func (tvs TrackVolumeSet) JSON() *json.Container {
	return json.Object("attribute", "track-volume", "value", tvs.TrackVolume)
}

func (tvs TrackVolumeSet) updatePart(part *Part, globalUpdate bool) {
	part.TrackVolume = tvs.TrackVolume
}

// PanningSet sets the panning of all active parts.
type PanningSet struct {
	Panning float64
}

// JSON implements RepresentableAsJSON.JSON.
func (ps PanningSet) JSON() *json.Container {
	return json.Object("attribute", "panning", "value", ps.Panning)
}

func (ps PanningSet) updatePart(part *Part, globalUpdate bool) {
	part.Panning = ps.Panning
}

// QuantizationSet sets the quantization of all active parts.
type QuantizationSet struct {
	Quantization float64
}

// JSON implements RepresentableAsJSON.JSON.
func (qs QuantizationSet) JSON() *json.Container {
	return json.Object("attribute", "quantization", "value", qs.Quantization)
}

func (qs QuantizationSet) updatePart(part *Part, globalUpdate bool) {
	part.Quantization = qs.Quantization
}

// DurationSet sets the quantization of all active parts.
type DurationSet struct {
	Duration Duration
}

// JSON implements RepresentableAsJSON.JSON.
func (ds DurationSet) JSON() *json.Container {
	object := ds.Duration.JSON()
	object.Set("duration", "attribute")
	return object
}

func (ds DurationSet) updatePart(part *Part, globalUpdate bool) {
	part.Duration = ds.Duration
}

// KeySignatureSet sets the key signature of all active parts.
type KeySignatureSet struct {
	KeySignature KeySignature
}

// JSON implements RepresentableAsJSON.JSON.
func (kss KeySignatureSet) JSON() *json.Container {
	return json.Object(
		"attribute", "key-signature",
		"value", kss.KeySignature.JSON(),
	)
}

func (kss KeySignatureSet) updatePart(part *Part, globalUpdate bool) {
	part.KeySignature = kss.KeySignature
}

// TranspositionSet sets the transposition of all active parts.
type TranspositionSet struct {
	Semitones int32
}

// JSON implements RepresentableAsJSON.JSON.
func (ts TranspositionSet) JSON() *json.Container {
	return json.Object(
		"attribute", "transposition",
		"value", ts.Semitones,
	)
}

func (ts TranspositionSet) updatePart(part *Part, globalUpdate bool) {
	part.Transposition = ts.Semitones
}

// ReferencePitchSet sets the reference pitch of all active parts. The reference
// pitch is represented as the frequency of A4.
type ReferencePitchSet struct {
	Frequency float64
}

// JSON implements RepresentableAsJSON.JSON.
func (rps ReferencePitchSet) JSON() *json.Container {
	return json.Object(
		"attribute", "reference-pitch",
		"value", rps.Frequency,
	)
}

func (rps ReferencePitchSet) updatePart(part *Part, globalUpdate bool) {
	part.ReferencePitch = rps.Frequency
}

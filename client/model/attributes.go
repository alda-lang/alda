package model

import (
	"fmt"
	"reflect"
	"sort"

	"alda.io/client/help"
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
		if err := au.PartUpdate.updatePart(part, false); err != nil {
			return err
		}
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
func (score *Score) ApplyGlobalAttributes() error {
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

				if err := update.updatePart(part, true); err != nil {
					return err
				}
			}
		}

		part.localAttributeOverride = nil
	}

	return nil
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

func (ts TempoSet) updatePart(part *Part, globalUpdate bool) error {
	part.Tempo = ts.Tempo

	// Global updates are recorded separately, and we would end up getting
	// incorrect results anyway if we recorded the tempo at the part's current
	// offset, because it might be different than the offset at which the global
	// attribute change was actually placed.
	if !globalUpdate {
		part.RecordTempoValue()
	}

	return nil
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

func (mm MetricModulation) updatePart(part *Part, globalUpdate bool) error {
	part.Tempo *= mm.Ratio

	// Global updates are recorded separately, and we would end up getting
	// incorrect results anyway if we recorded the tempo at the part's current
	// offset, because it might be different than the offset at which the global
	// attribute change was actually placed.
	if !globalUpdate {
		part.RecordTempoValue()
	}

	return nil
}

// OctaveSet sets the octave of all active parts.
type OctaveSet struct {
	OctaveNumber int32
}

// JSON implements RepresentableAsJSON.JSON.
func (os OctaveSet) JSON() *json.Container {
	return json.Object("attribute", "octave", "value", os.OctaveNumber)
}

func (os OctaveSet) updatePart(part *Part, globalUpdate bool) error {
	part.Octave = os.OctaveNumber

	return nil
}

// OctaveUp increments the octave of all active parts.
type OctaveUp struct{}

// JSON implements RepresentableAsJSON.JSON.
func (os OctaveUp) JSON() *json.Container {
	return json.Object("attribute", "octave", "value", "up")
}

func (OctaveUp) updatePart(part *Part, globalUpdate bool) error {
	part.Octave++

	return nil
}

// OctaveDown decrements the octave of all active parts.
type OctaveDown struct{}

// JSON implements RepresentableAsJSON.JSON.
func (os OctaveDown) JSON() *json.Container {
	return json.Object("attribute", "octave", "value", "down")
}

func (OctaveDown) updatePart(part *Part, globalUpdate bool) error {
	part.Octave--

	return nil
}

// VolumeSet sets the volume of all active parts.
type VolumeSet struct {
	Volume float64
}

// JSON implements RepresentableAsJSON.JSON.
func (vs VolumeSet) JSON() *json.Container {
	return json.Object("attribute", "volume", "value", vs.Volume)
}

func (vs VolumeSet) updatePart(part *Part, globalUpdate bool) error {
	part.Volume = vs.Volume

	return nil
}

// TrackVolumeSet sets the track volume of all active parts.
type TrackVolumeSet struct {
	TrackVolume float64
}

// JSON implements RepresentableAsJSON.JSON.
func (tvs TrackVolumeSet) JSON() *json.Container {
	return json.Object("attribute", "track-volume", "value", tvs.TrackVolume)
}

func (tvs TrackVolumeSet) updatePart(part *Part, globalUpdate bool) error {
	part.TrackVolume = tvs.TrackVolume

	return nil
}

var DynamicVolumes map[string]float64

func init() {
	// Dynamic volumes in Alda follow a uniform distribution from 0 to 1
	// This follows the standard set by MIDI and existing software programs
	// Alda supports the full range of MusicXML dynamics from pppppp to ffffff
	// Volumes are mapped to MIDI velocity [0, 127] by multiplying by 127
	// The default Alda volume is mf
	// MIDI velocities are commented
	DynamicVolumes = map[string]float64{
		"pppppp": 0.00787, // 1
		"ppppp":  0.08419, // 11
		"pppp":   0.16051, // 20
		"ppp":    0.23683, // 30
		"pp":     0.31314, // 40
		"p":      0.38946, // 49
		"mp":     0.46578, // 59
		"mf":     0.54210, // 69
		"f":      0.61841, // 79
		"ff":     0.69473, // 88
		"fff":    0.77105, // 98
		"ffff":   0.84737, // 108
		"fffff":  0.92368, // 117
		"ffffff": 1.00000, // 127
	}
}

type DynamicMarking struct {
	Marking string
}

func (dm DynamicMarking) JSON() *json.Container {
	return json.Object("attribute", "dynamic-marking", "value", dm.Marking)
}

func (dm DynamicMarking) updatePart(part *Part, globalUpdate bool) error {
	part.Volume = DynamicVolumes[dm.Marking]

	return nil
}

// PanningSet sets the panning of all active parts.
type PanningSet struct {
	Panning float64
}

// JSON implements RepresentableAsJSON.JSON.
func (ps PanningSet) JSON() *json.Container {
	return json.Object("attribute", "panning", "value", ps.Panning)
}

func (ps PanningSet) updatePart(part *Part, globalUpdate bool) error {
	part.Panning = ps.Panning

	return nil
}

// QuantizationSet sets the quantization of all active parts.
type QuantizationSet struct {
	Quantization float64
}

// JSON implements RepresentableAsJSON.JSON.
func (qs QuantizationSet) JSON() *json.Container {
	return json.Object("attribute", "quantization", "value", qs.Quantization)
}

func (qs QuantizationSet) updatePart(part *Part, globalUpdate bool) error {
	part.Quantization = qs.Quantization

	return nil
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

func (ds DurationSet) updatePart(part *Part, globalUpdate bool) error {
	part.Duration = ds.Duration

	return nil
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

func (kss KeySignatureSet) updatePart(part *Part, globalUpdate bool) error {
	part.KeySignature = kss.KeySignature

	return nil
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

func (ts TranspositionSet) updatePart(part *Part, globalUpdate bool) error {
	part.Transposition = ts.Semitones

	return nil
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

func (rps ReferencePitchSet) updatePart(part *Part, globalUpdate bool) error {
	part.ReferencePitch = rps.Frequency

	return nil
}

// MidiChannelSet sets the MIDI channel to use for all active parts. By default,
// a MIDI channel is assigned automatically, reusing another channel and
// switching patches if necessary (i.e. if there are > 15 non-percussion
// instruments in the score).
//
// This attribute can be used in cases where you want to explicitly control
// which MIDI channel is used.
type MidiChannelSet struct {
	ChannelNumber int32
}

// JSON implements RepresentableAsJSON.JSON.
func (mcs MidiChannelSet) JSON() *json.Container {
	return json.Object(
		"attribute", "midi-channel",
		"value", mcs.ChannelNumber,
	)
}

func (mcs MidiChannelSet) updatePart(part *Part, globalUpdate bool) error {
	// TODO: Update this type assertion if/when we add non-MIDI instruments.
	if mcs.ChannelNumber == 9 &&
		!part.StockInstrument.(MidiInstrument).IsPercussion {

		return help.UserFacingErrorf(
			`Can't use MIDI channel 9 for part "%s"; channel 9 can only be used for
percussion.`,
			part.Name,
		)
	}

	part.MidiChannel = mcs.ChannelNumber
	part.HasExplicitMidiChannel = true

	return nil
}

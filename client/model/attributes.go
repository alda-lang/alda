package model

// AttributeUpdate updates the value of an attribute for all active parts.
type AttributeUpdate struct {
	PartUpdate PartUpdate
}

func (au AttributeUpdate) updateScore(score *Score) error {
	for _, part := range score.CurrentParts {
		au.PartUpdate.updatePart(part)
	}

	return nil
}

// GlobalAttributeUpdate updates the value of an attribute for all parts.
type GlobalAttributeUpdate struct {
	PartUpdate PartUpdate
}

func (gau GlobalAttributeUpdate) updateScore(score *Score) error {
	// TODO: keep a record of global attribute changes, like alda v1 does
	// This is necessary so that new parts can pick up on global attribute changes
	// as they cross the offsets where the global attribute changes occur.

	for _, part := range score.Parts {
		gau.PartUpdate.updatePart(part)
	}

	return nil
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

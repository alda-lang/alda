package io.alda.player

interface SoundEngine {
  fun midiNote(
    startOffset : Int, endOffset : Int, channel : Int, noteNumber : Int,
    velocity : Int
  )

  fun midiPanning(offset : Int, channel : Int, panning : Int)

  fun midiPatch(offset : Int, channel : Int, patch : Int)

  fun midiPercussionImmediate(trackNumber : Int)

  fun midiPercussionScheduled(trackNumber : Int, offset : Int)

  fun midiVolume(offset : Int, channel : Int, volume : Int)
}

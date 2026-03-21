package io.alda.player

import kotlinx.coroutines.Job

interface SoundEngine {
  fun isPlaying() : Boolean

  fun startSequencer()

  fun stopSequencer()

  fun currentOffset() : Double

  fun midiNote(
    startOffset : Int, endOffset : Int, channel : Int, noteNumber : Int,
    velocity : Int
  )

  fun midiClearChannel(channelNumber : Int)

  fun midiMuteChannel(channelNumber : Int)

  fun midiPanning(offset : Int, channel : Int, panning : Int)

  fun midiPatch(offset : Int, channel : Int, patch : Int)

  fun midiPercussionImmediate(trackNumber : Int)

  fun midiPercussionScheduled(trackNumber : Int, offset : Int)

  fun midiUnmuteChannel(channelNumber : Int)

  fun midiVolume(offset : Int, channel : Int, volume : Int)

  fun scheduleEvent(offset : Int, eventName : String) : Job
}

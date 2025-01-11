package io.alda.player

interface SoundEngine {
  fun midiNote(
    startOffset : Int, endOffset : Int, channel : Int, noteNumber : Int,
    velocity : Int
  )

  fun midiPanning(offset : Int, channel : Int, panning : Int)

  fun midiPatch(offset : Int, channel : Int, patch : Int)

  fun midiVolume(offset : Int, channel : Int, volume : Int)

  fun schedule(event: Event) {
    when (event["type"]) {
      "midi-note" -> {
        val noteStart = event["offset"] as Int
        val noteEnd = noteStart + event["audible-duration"] as Int

        midiNote(
          noteStart,
          noteEnd,
          event["channel"] as Int,
          event["note-number"] as Int,
          event["velocity"] as Int
        )
      }

      "midi-panning" -> midiPanning(
        event["offset"] as Int,
        event["channel"] as Int,
        event["panning"] as Int
      )

      "midi-patch" -> midiPatch(
        event["offset"] as Int,
        event["channel"] as Int,
        event["patch"] as Int
      )

      "midi-volume" -> midiVolume(
        event["offset"] as Int,
        event["channel"] as Int,
        event["volume"] as Int
      )
    }
  }
}

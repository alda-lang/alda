package io.alda.player

import com.illposed.osc.OSCMessage

enum class SystemAction {
  PLAY, STOP, CLEAR
}

enum class TrackAction {
  MUTE, UNMUTE, CLEAR
}

enum class PatternAction {
  CLEAR
}

interface Event {}

class MidiPatchEvent(val offset : Int, val patch : Int) : Event {}

class MidiPercussionEvent(val offset : Int) : Event {}

class MidiNoteEvent(
  val offset : Int, val noteNumber : Int, val duration : Int,
  val audibleDuration : Int, val velocity : Int
) : Event {
  fun addOffset(o : Int) : MidiNoteEvent {
    return MidiNoteEvent(
      offset + o, noteNumber, duration, audibleDuration, velocity
    )
  }
}

class PatternEvent(
  val offset : Int, val patternName : String, val times : Int
) : Event {}

class PatternLoopEvent(val offset : Int, val patternName : String) : Event {}

class FinishLoopEvent(val offset : Int) : Event {}

class Updates() {
  var systemActions  = mutableSetOf<SystemAction>()
  var trackActions   = mutableMapOf<Int, Set<TrackAction>>()
  var patternActions = mutableMapOf<String, Set<PatternAction>>()

  var trackEvents    = mutableMapOf<Int, List<Event>>()
  var patternEvents  = mutableMapOf<String, List<Event>>()

  private fun addTrackAction(track : Int, action : TrackAction) {
    if (!trackActions.containsKey(track))
      trackActions[track] = mutableSetOf<TrackAction>()

    (trackActions.getValue(track) as MutableSet).add(action)
  }

  private fun addPatternAction(pattern : String, action : PatternAction) {
    if (!patternActions.containsKey(pattern))
      patternActions[pattern] = mutableSetOf<PatternAction>()

    (patternActions.getValue(pattern) as MutableSet).add(action)
  }

  private fun addTrackEvent(track : Int, event : Event) {
    if (!trackEvents.containsKey(track))
      trackEvents[track] = mutableListOf<Event>()

    (trackEvents.getValue(track) as MutableList).add(event)
  }

  private fun addPatternEvent(pattern : String, event : Event) {
    if (!patternEvents.containsKey(pattern))
      patternEvents[pattern] = mutableListOf<Event>()

    (patternEvents.getValue(pattern) as MutableList).add(event)
  }

  private fun trackNumber(address : String) : Int {
    val (track) = "/track/(\\d+)".toRegex()
                                 .find(address)!!
                                 .destructured
    return track.toInt()
  }

  private fun patternName(address : String) : String {
    val (pattern) = "/pattern/([^/]+)".toRegex()
                                      .find(address)!!
                                      .destructured
    return pattern
  }

  fun parse(msg : OSCMessage) {
    val address = msg.getAddress()
    val args    = msg.getArguments()

    try {
      when {
        Regex("/system/play").matches(address) -> {
          systemActions.add(SystemAction.PLAY)
        }

        Regex("/system/stop").matches(address) -> {
          systemActions.add(SystemAction.STOP)
        }

        Regex("/system/clear").matches(address) -> {
          systemActions.add(SystemAction.CLEAR)
        }

        Regex("/track/\\d+/unmute").matches(address) -> {
          addTrackAction(trackNumber(address), TrackAction.UNMUTE)
        }

        Regex("/track/\\d+/mute").matches(address) -> {
          addTrackAction(trackNumber(address), TrackAction.MUTE)
        }

        Regex("/track/\\d+/clear").matches(address) -> {
          addTrackAction(trackNumber(address), TrackAction.CLEAR)
        }

        Regex("/track/\\d+/midi/patch").matches(address) -> {
          val offset = args.get(0) as Int
          val patch = args.get(1) as Int
          addTrackEvent(trackNumber(address), MidiPatchEvent(offset, patch))
        }

        Regex("/track/\\d+/midi/percussion").matches(address) -> {
          val offset = args.get(0) as Int
          addTrackEvent(trackNumber(address), MidiPercussionEvent(offset))
        }

        Regex("/track/\\d+/midi/note").matches(address) -> {
          val offset          = args.get(0) as Int
          val noteNumber      = args.get(1) as Int
          val duration        = args.get(2) as Int
          val audibleDuration = args.get(3) as Int
          val velocity        = args.get(4) as Int

          addTrackEvent(
            trackNumber(address),
            MidiNoteEvent(
              offset, noteNumber, duration, audibleDuration, velocity
            )
          )
        }

        Regex("/track/\\d+/pattern").matches(address) -> {
          val offset      = args.get(0) as Int
          val patternName = args.get(1) as String
          val times       = args.get(2) as Int

          addTrackEvent(
            trackNumber(address),
            PatternEvent(offset, patternName, times)
          )
        }

        Regex("/track/\\d+/pattern-loop").matches(address) -> {
          val offset      = args.get(0) as Int
          val patternName = args.get(1) as String
          addTrackEvent(trackNumber(address), PatternLoopEvent(offset, patternName))
        }

        Regex("/track/\\d+/finish-loop").matches(address) -> {
          val offset = args.get(0) as Int
          addTrackEvent(trackNumber(address), FinishLoopEvent(offset))
        }

        Regex("/pattern/[^/]+/clear").matches(address) -> {
          addPatternAction(patternName(address), PatternAction.CLEAR)
        }

        Regex("/pattern/[^/]+/midi/note").matches(address) -> {
          val offset          = args.get(0) as Int
          val noteNumber      = args.get(1) as Int
          val duration        = args.get(2) as Int
          val audibleDuration = args.get(3) as Int
          val velocity        = args.get(4) as Int

          addPatternEvent(
            patternName(address),
            MidiNoteEvent(
              offset, noteNumber, duration, audibleDuration, velocity
            )
          )
        }

        Regex("/pattern/[^/]+/pattern").matches(address) -> {
          val offset      = args.get(0) as Int
          val patternName = args.get(1) as String
          val times       = args.get(2) as Int

          addPatternEvent(
            patternName(address),
            PatternEvent(offset, patternName, times)
          )
        }

        else -> {
          println("WARN: Unrecognized address: ${address}")
        }
      }
    } catch (e : Throwable) {
      println("WARN: Error while processing ${address} :: ${args}")
      e.printStackTrace()
    }
  }
}

fun parseUpdates(instructions : List<OSCMessage>) : Updates {
  val updates = Updates()
  instructions.forEach { updates.parse(it) }
  return updates
}

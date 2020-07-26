package io.alda.player

import com.illposed.osc.OSCMessage
import mu.KotlinLogging

private val log = KotlinLogging.logger {}

enum class SystemAction {
  SHUTDOWN, PLAY, STOP, CLEAR
}

enum class TrackAction {
  MUTE, UNMUTE, CLEAR
}

enum class PatternAction {
  CLEAR
}

interface Event {
  fun addOffset(o : Int) : Event
  fun endOffset() : Int
}

interface Schedulable {
  fun schedule(channel : Int)
}

class ShutdownEvent(val offset : Int) : Event {
  override fun addOffset(o : Int) : ShutdownEvent {
    return ShutdownEvent(offset + o)
  }

  override fun endOffset() = 0
}

class SetOffsetEvent(val offset : Int) : Event {
  override fun addOffset(o : Int) : SetOffsetEvent {
    return SetOffsetEvent(offset + o)
  }

  override fun endOffset() = 0
}

class TempoEvent(val offset : Int, val bpm : Float) : Event {
  override fun addOffset(o : Int) : TempoEvent {
    return TempoEvent(offset + o, bpm)
  }

  override fun endOffset() = 0
}

class MidiPatchEvent(val offset : Int, val patch : Int) : Event, Schedulable {
  override fun addOffset(o : Int) : MidiPatchEvent {
    return MidiPatchEvent(offset + o, patch)
  }

  override fun schedule(channel : Int) {
    midi().patch(offset, channel, patch)
  }

  override fun endOffset() = 0
}

class MidiPercussionEvent(val offset : Int) : Event {
  override fun addOffset(o : Int) : MidiPercussionEvent {
    return MidiPercussionEvent(offset + o)
  }

  override fun endOffset() = 0
}

class MidiNoteEvent(
  val offset : Int, val noteNumber : Int, val duration : Int,
  val audibleDuration : Int, val velocity : Int
) : Event, Schedulable {
  override fun addOffset(o : Int) : MidiNoteEvent {
    return MidiNoteEvent(
      offset + o, noteNumber, duration, audibleDuration, velocity
    )
  }

  override fun schedule(channel : Int) {
    val noteStart = offset
    val noteEnd = noteStart + audibleDuration
    midi().note(noteStart, noteEnd, channel, noteNumber, velocity)
  }

  override fun endOffset() = offset + duration
}

class MidiVolumeEvent(
  val offset : Int, val volume : Int
) : Event, Schedulable {
  override fun addOffset(o : Int) : MidiVolumeEvent {
    return MidiVolumeEvent(offset + o, volume)
  }

  override fun schedule(channel : Int) {
    midi().volume(offset, channel, volume)
  }

  override fun endOffset() = 0
}

class MidiPanningEvent(
  val offset : Int, val panning : Int
) : Event, Schedulable {
  override fun addOffset(o : Int) : MidiPanningEvent {
    return MidiPanningEvent(offset + o, panning)
  }

  override fun schedule(channel : Int) {
    midi().panning(offset, channel, panning)
  }

  override fun endOffset() = 0
}

abstract class PatternEventBase(
  open val offset : Int, open val patternName : String
) {
  abstract fun isDone(iteration : Int) : Boolean
}

class PatternEvent(
  override val offset : Int, override val patternName : String, val times : Int
) : Event, PatternEventBase(offset, patternName) {
  override fun addOffset(o : Int) : PatternEvent {
    return PatternEvent(offset + o, patternName, times)
  }

  override fun isDone(iteration : Int) : Boolean {
    return iteration > times
  }

  override fun endOffset() = 0
}

class PatternLoopEvent(
  override val offset : Int, override val patternName : String
) : Event, PatternEventBase(offset, patternName) {
  override fun addOffset(o : Int) : PatternLoopEvent {
    return PatternLoopEvent(offset + o, patternName)
  }

  override fun isDone(iteration : Int) : Boolean = false

  override fun endOffset() = 0
}

class FinishLoopEvent(val offset : Int) : Event {
  override fun addOffset(o : Int) : FinishLoopEvent {
    return FinishLoopEvent(offset + o)
  }

  override fun endOffset() = 0
}

class MidiExportEvent(val filepath : String) : Event {
  override fun addOffset(o : Int) : MidiExportEvent {
    return MidiExportEvent(filepath)
  }

  override fun endOffset() = 0
}

class Updates() {
  var systemActions  = mutableSetOf<SystemAction>()
  var trackActions   = mutableMapOf<Int, Set<TrackAction>>()
  var patternActions = mutableMapOf<String, Set<PatternAction>>()

  var systemEvents   = mutableListOf<Event>()
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

    log.trace { "${address} ${args}" }

    try {
      when {
        // When a client (e.g. an Alda REPL server) sends a /ping message, it
        // has the effect of "claiming" the player process, i.e. putting the
        // player into the "active" state and prolonging the expiration of the
        // player process.
        //
        // This is helpful because it allows an Alda REPL server to continuously
        // use the same player process for playback across multiple evaluations,
        // which enables live coding. The Alda REPL server repeatedly sends
        // /ping messages, which ensures that the player process will not expire
        // before the server is done using it.
        Regex("/ping").matches(address) -> {
          log.trace("received ping")
          stateManager!!.markActive()
          stateManager!!.delayExpiration()
        }

        Regex("/system/shutdown").matches(address) -> {
          val offset = args.get(0) as Int

          // There are two "modes" of shutting down:
          //
          // 1. A system action that immediately shuts the player process down.
          //    Use case: immediately shutting the player down
          //
          // 2. A system event that schedules the player to be shut down when a
          //    particular offset is reached.
          //    Use case: shutting the player down after the end of a score
          if (offset == 0) {
            systemActions.add(SystemAction.SHUTDOWN)
          } else {
            systemEvents.add(ShutdownEvent(offset))
          }
        }

        Regex("/system/play").matches(address) -> {
          systemActions.add(SystemAction.PLAY)
        }

        Regex("/system/stop").matches(address) -> {
          systemActions.add(SystemAction.STOP)
        }

        Regex("/system/offset").matches(address) -> {
          val offset = args.get(0) as Int
          systemEvents.add(SetOffsetEvent(offset))
        }

        Regex("/system/clear").matches(address) -> {
          systemActions.add(SystemAction.CLEAR)
        }

        Regex("/system/tempo").matches(address) -> {
          val offset = args.get(0) as Int
          val bpm = args.get(1) as Float
          systemEvents.add(TempoEvent(offset, bpm))
        }

        Regex("/system/midi/export").matches(address) -> {
          val filepath = args.get(0) as String
          systemEvents.add(MidiExportEvent(filepath))
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

        Regex("/track/\\d+/midi/volume").matches(address) -> {
          val offset = args.get(0) as Int
          val volume = args.get(1) as Int
          addTrackEvent(trackNumber(address), MidiVolumeEvent(offset, volume))
        }

        Regex("/track/\\d+/midi/panning").matches(address) -> {
          val offset  = args.get(0) as Int
          val panning = args.get(1) as Int
          addTrackEvent(trackNumber(address), MidiPanningEvent(offset, panning))
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

        Regex("/pattern/[^/]+/midi/volume").matches(address) -> {
          val offset = args.get(0) as Int
          val volume = args.get(1) as Int
          addPatternEvent(
            patternName(address), MidiVolumeEvent(offset, volume)
          )
        }

        Regex("/pattern/[^/]+/midi/panning").matches(address) -> {
          val offset  = args.get(0) as Int
          val panning = args.get(1) as Int
          addPatternEvent(
            patternName(address), MidiPanningEvent(offset, panning)
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
          log.warn { "Unrecognized address: ${address}" }
        }
      }
    } catch (e : Throwable) {
      log.warn { "Error while processing ${address} :: ${args}" }
      e.printStackTrace()
    }
  }
}

fun parseUpdates(instructions : List<OSCMessage>) : Updates {
  val updates = Updates()
  instructions.forEach { updates.parse(it) }
  return updates
}

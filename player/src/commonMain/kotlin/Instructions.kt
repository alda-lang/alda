package io.alda.player

import mu.KotlinLogging

private val log = KotlinLogging.logger {}

enum class SystemAction {
  SHUTDOWN, PLAY, STOP, CLEAR
}

enum class TrackAction {
  CLEAR
}

enum class PatternAction {
  CLEAR
}

interface Schedulable {
  fun schedule(soundEngine : SoundEngine, channel : Int)
}

data class Message(val address : String, val args : List<Any>) {}

class Updates(val messages : List<Message>, val stateManager : StateManager) {
  var systemActions  = mutableSetOf<SystemAction>()
  var trackActions   = mutableMapOf<Int, Set<TrackAction>>()
  var patternActions = mutableMapOf<String, Set<PatternAction>>()

  var systemEvents   = mutableListOf<Event>()
  var trackEvents    = mutableMapOf<Int, List<Event>>()
  var patternEvents  = mutableMapOf<String, List<Event>>()

  init {
    for (message in messages) {
      parseMessage(message, stateManager)
    }
  }

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

  fun parseMessage(message : Message, stateManager : StateManager) {
    val address = message.address
    val args = message.args

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
          log.debug { "received ping" }
          stateManager.markActive()
          stateManager.delayExpiration()
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
            systemEvents.add(mapOf("type" to "shutdown", "offset" to offset))
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

          systemEvents.add(mapOf("type" to "set-offset", "offset" to offset))
        }

        Regex("/system/clear").matches(address) -> {
          systemActions.add(SystemAction.CLEAR)
        }

        Regex("/system/tempo").matches(address) -> {
          val offset = args.get(0) as Int
          val bpm = args.get(1) as Float

          systemEvents.add(
            mapOf("type" to "tempo", "offset" to offset, "bpm" to bpm)
          )
        }

        Regex("/system/midi/export").matches(address) -> {
          val filepath = args.get(0) as String

          systemEvents.add(
            mapOf("type" to "midi-export", "filepath" to filepath)
          )
        }

        Regex("/track/\\d+/clear").matches(address) -> {
          addTrackAction(trackNumber(address), TrackAction.CLEAR)
        }

        Regex("/track/\\d+/midi/patch").matches(address) -> {
          val channel = args.get(0) as Int
          val offset  = args.get(1) as Int
          val patch   = args.get(2) as Int

          addTrackEvent(
            trackNumber(address),
            mapOf(
              "type" to "midi-patch",
              "channel" to channel,
              "offset" to offset,
              "patch" to patch
            )
          )
        }

        Regex("/track/\\d+/midi/note").matches(address) -> {
          val channel         = args.get(0) as Int
          val offset          = args.get(1) as Int
          val noteNumber      = args.get(2) as Int
          val duration        = args.get(3) as Int
          val audibleDuration = args.get(4) as Int
          val velocity        = args.get(5) as Int

          addTrackEvent(
            trackNumber(address),
            mapOf(
              "type" to "midi-note",
              "channel" to channel,
              "offset" to offset,
              "note-number" to noteNumber,
              "duration" to duration,
              "audible-duration" to audibleDuration,
              "velocity" to velocity
            )
          )
        }

        Regex("/track/\\d+/midi/volume").matches(address) -> {
          val channel = args.get(0) as Int
          val offset  = args.get(1) as Int
          val volume  = args.get(2) as Int

          addTrackEvent(
            trackNumber(address),
            mapOf(
              "type" to "midi-volume",
              "channel" to channel,
              "offset" to offset,
              "volume" to volume
            )
          )
        }

        Regex("/track/\\d+/midi/panning").matches(address) -> {
          val channel = args.get(0) as Int
          val offset  = args.get(1) as Int
          val panning = args.get(2) as Int

          addTrackEvent(
            trackNumber(address),
            mapOf(
              "type" to "midi-panning",
              "channel" to channel,
              "offset" to offset,
              "panning" to panning
            )
          )
        }

        Regex("/track/\\d+/pattern").matches(address) -> {
          val channel     = args.get(0) as Int
          val offset      = args.get(1) as Int
          val patternName = args.get(2) as String
          val times       = args.get(3) as Int

          addTrackEvent(
            trackNumber(address),
            mapOf(
              "type" to "pattern",
              "channel" to channel,
              "offset" to offset,
              "pattern-name" to patternName,
              "times" to times
            )
          )
        }

        Regex("/track/\\d+/pattern-loop").matches(address) -> {
          val channel     = args.get(0) as Int
          val offset      = args.get(1) as Int
          val patternName = args.get(2) as String

          addTrackEvent(
            trackNumber(address),
            mapOf(
              "type" to "pattern-loop",
              "channel" to channel,
              "offset" to offset,
              "pattern-name" to patternName
            )
          )
        }

        Regex("/track/\\d+/finish-loop").matches(address) -> {
          val offset = args.get(0) as Int

          addTrackEvent(
            trackNumber(address),
            mapOf("type" to "finish-loop", "offset" to offset)
          )
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
            mapOf(
              "type" to "midi-note",
              "offset" to offset,
              "note-number" to noteNumber,
              "duration" to duration,
              "audible-duration" to audibleDuration,
              "velocity" to velocity
            )
          )
        }

        Regex("/pattern/[^/]+/midi/volume").matches(address) -> {
          val offset = args.get(0) as Int
          val volume = args.get(1) as Int

          addPatternEvent(
            patternName(address),
            mapOf(
              "type" to "midi-volume",
              "offset" to offset,
              "volume" to volume
            )
          )
        }

        Regex("/pattern/[^/]+/midi/panning").matches(address) -> {
          val offset  = args.get(0) as Int
          val panning = args.get(1) as Int

          addPatternEvent(
            patternName(address),
            mapOf(
              "type" to "midi-panning",
              "offset" to offset,
              "panning" to panning
            )
          )
        }

        Regex("/pattern/[^/]+/pattern").matches(address) -> {
          val offset      = args.get(0) as Int
          val patternName = args.get(1) as String
          val times       = args.get(2) as Int

          addPatternEvent(
            patternName(address),
            mapOf(
              "type" to "pattern",
              "offset" to offset,
              "pattern-name" to patternName,
              "times" to times
            )
          )
        }

        else -> {
          log.warn { "Unrecognized address: ${address}" }
        }
      }
    } catch (e : Throwable) {
      log.warn(e) { "Error while processing ${address} :: ${args}" }
      e.printStackTrace()
    }
  }
}

package io.alda.player

import kotlin.math.roundToInt
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.async
import mu.KotlinLogging

private val log = KotlinLogging.logger {}

// We often want to schedule a note to be played "right now," but that's not
// actually possible unless time has stopped. If we schedule the note for "now"
// (i.e. the current offset), the note will not be played because it takes a
// non-zero amount of time to do the work of scheduling the note, and by the
// time we're done, the current offset has advanced past the offset where the
// note should have been played.
//
// This value is the amount of latency that we are reasonably willing to add to
// allow the scheduler time to schedule the note.
//
// TODO: make this configurable?
const val SCHEDULE_BUFFER_TIME_MS = 400

val availableChannels2 = ((0..15).toSet() - 9).toMutableSet()

class Track2(val trackNumber : Int, val engine : SoundEngine) {
  private var _midiChannel : Int? = null
  fun midiChannel() : Int? {
    // NOTE: This used to be `synchronized(availableChannels) { ... }`
    // Removed in order to work with both JVM and JS.
    // TODO: Do we need to conditionally synchronize in the JVM version?
    // Guidance is to write common code such that it does not use shared state.
    // FIXME: Was this a mistake? I think my use of `synchronized` in the JVM
    // code was probably justified.
    // Maybe I won't need `synchronized` if I convert all of the multithreaded
    // stuff to use coroutines instead?
    if (_midiChannel == null && !availableChannels2.isEmpty()) {
      val channel = availableChannels2.first()
      availableChannels2.remove(channel)
      _midiChannel = channel
    }

    return _midiChannel
  }

  private fun withMidiChannel(f : (Int) -> Unit) {
    midiChannel()?.also { channel ->
      f(channel)
    } ?: run {
      log.warn { "No MIDI channel available for track ${trackNumber}." }
    }
  }

  fun useMidiPercussionChannel() { _midiChannel = 9 }

  data class BufferedEvents(val era: Int, val events: List<Event>)

  // TODO: figure out how to make this cross-platform
  // Try using a coroutine Channel. You can't clear them like you can a
  // LinkedBlockingQueue, but I found this SO question where it was suggested
  // that you can set some kind of expiry or counter on the message, and
  // dropping any messages that aren't up to date. This is similar to my `era`
  // idea. In fact, maybe we could use era for this? Would it work if we
  // included the current era on events when sending them on the channel, and we
  // will only process them if the era hasn't incremented by the time they are
  // received?
  //
  // val eventBufferQueue = LinkedBlockingQueue<List<Event>>()
  val eventBufferQueue = Channel<BufferedEvents>()

  // A count of tasks (List<Event>) that have been taken off of the
  // `eventBufferQueue` and are currently being processed.
  //
  // The count is incremented before a task is placed on the queue and
  // decremented once the task is complete (i.e. events have been scheduled).
  // TODO: Figure out what to do about this in JS.
  // Am I going about this all the wrong way? Should it really be the case that
  // the common library is fully stateless? Can we adjust the API so that there
  // is a small part that represents state management, and it can vary for JVM
  // vs. JS?
  //
  // Update 2022-08-19: Maybe the key is to make `activeTasks` hold the actual
  // tasks, which could be Deferreds (which are the result of `async { ... }`).
  // Then, I could define `awaitActiveTasks` on Track by having it loop through
  // the Deferreds and call `.await()` on each of them. (Or maybe there is
  // already something similar to `Promise.all` in the stdlib?)
  var activeTasks = 0

  // The set of patterns that are currently looping (whether that be a finite
  // number of times or indefinitely).
  val activePatterns = mutableSetOf<String>()

  // A monotonically increasing integer representing all of the events we are
  // going to schedule until the track is cleared. Event scheduling is done
  // asynchronously, so incrementing `era` is a way to signal to the various
  // coroutines that the track has been cleared, i.e. don't proceed to schedule
  // events.
  var era = 0

  // The base offset that is added to upcoming notes to be scheduled. As notes
  // are scheduled, this base offset is updated to reflect the offset at which
  // the last note to be scheduled will end, so that subsequent notes will line
  // up in time right after the last note.
  var startOffset = 0

  fun clear() {
    // This used to be `synchronized(era) { ... }`
    era++
    startOffset = 0
    // TODO: Confirm that we don't need this anymore after making it so that
    // incrementing the era results is discarding any BufferedEvents messages on
    // the channel that have an older era.
    // eventBufferQueue.clear()
    activeTasks = 0
    activePatterns.clear()
    withMidiChannel { engine.midiClearChannel(it) }
  }

  fun mute() {
    withMidiChannel { engine.midiMuteChannel(it) }
  }

  fun unmute() {
    withMidiChannel { engine.midiUnmuteChannel(it) }
  }

  fun schedule(event : Schedulable) {
    withMidiChannel { channel -> event.schedule(engine, channel) }
  }

  /**
   * Schedules the notes of a pattern, blocking until all iterations of the
   * pattern have been scheduled.
   *
   * Patterns can be looped, and the pattern can be changed while it's looping.
   * When this happens, the change is picked up upon the next iteration of the
   * loop. We accomplish this by scheduling each iteration in a "just in time"
   * manner, i.e. shortly before it is due to be played.
   *
   * @param patternEvent An event that specifies a pattern, a relative offset
   * where it should begin, and a number of times to play it.
   * @param _startOffset The absolute offset to which the relative offset is
   * added.
   * @return The list of scheduled events across all iterations of the pattern.
   */
  suspend fun schedulePattern(patternEvent : PatternEventBase, _startOffset : Int)
  : List<Schedulable> {
    var startOffset = _startOffset + patternEvent.offset
    val patternEvents = mutableListOf<Schedulable>()

    // A loop can be stopped externally by removing the pattern from
    // `activePatterns`. If this happens, we stop looping.
    activePatterns.add(patternEvent.patternName)

    try {
      var iteration = 1

      while (!patternEvent.isDone(iteration) &&
             activePatterns.contains(patternEvent.patternName)) {
        log.debug {
          "scheduling iteration $iteration; startOffset: $startOffset; " +
          "patternEvent.offset: ${patternEvent.offset}"
        }

        // This value is the point in time where we schedule the metamessage
        // that signals the lookup and scheduling of the pattern's events.
        //
        // This scheduling happens shortly before the pattern is to be played.
        val patternSchedule = adjustStartOffset(startOffset)

        log.debug { "patternSchedule: $patternSchedule" }

        // This returns a kotlinx.coroutines.Job that completes when the
        // `patternSchedule` offset is reached in the sequence.
        val job = engine.scheduleEvent(
          patternSchedule, patternEvent.patternName
        )

        // Wait until it's time to look up the pattern's current value and
        // schedule the events.
        //
        // We check the `era` now and then again when it's time to schedule, and
        // see if the track has been cleared in the meantime. We only proceed if
        // the track hasn't been cleared (i.e. if the `era` hasn't changed).
        //
        // This used to be `synchronized(era) { era }`
        val eraBefore = era
        log.debug { "awaiting scheduleEvent job" }
        job.join()
        // This used to be `synchronized(era) { era }`
        val eraAfter = era
        log.debug { "eraBefore: $eraBefore; eraAfter: $eraAfter" }
        if (eraBefore != eraAfter) break

        log.debug { "scheduling pattern ${patternEvent.patternName}" }

        val pattern = pattern(patternEvent.patternName)

        // It's safe to filter a List<Event> down to just the ones that are
        // Schedulable and then cast it to a List<Schedulable>.
        @Suppress("UNCHECKED_CAST")
        val events : MutableList<Schedulable> =
          (pattern.events.map { it.addOffset(startOffset) }
                         .filter { it is Schedulable }
                         as List<Schedulable>)
            .toMutableList()

        events.forEach { schedule(it) }

        // Now that we've scheduled at least one iteration, we can start
        // playing. (Unless we've already started playing, in which case this is
        // a no-op.)
        //
        // This used to be `synchronized(engine.isPlaying) { ... }`
        if (engine.isPlaying()) {
          engine.startSequencer()
        }

        // It's safe to filter a List<Event> down to just the ones that are
        // PatternEvents and then cast it to a List<PatternEvent>.
        @Suppress("UNCHECKED_CAST")
        // Here, we handle the case where the pattern's events include further
        // pattern events, i.e. the pattern references another pattern.
        //
        // NB: Because of the "just in time" semantics of scheduling patterns,
        // this means we block here until the subpattern is about due to be
        // played.
        (pattern.events.filter { it is PatternEvent } as List<PatternEvent>)
          .forEach { events.addAll(schedulePattern(it, startOffset)) }

        if (!events.isEmpty())
          startOffset = events.map { (it as Event).endOffset() }.maxOrNull()!!

        patternEvents.addAll(events)

        iteration++
      }
    } finally {
      activePatterns.remove(patternEvent.patternName)
    }

    return patternEvents
  }

  private fun adjustStartOffset(_startOffset : Int) : Int {
    var startOffset = _startOffset

    val now = engine.currentOffset().roundToInt()

    // If we're not scheduling into the future, then whatever we're supposed to
    // be scheduling should happen ASAP.
    if (startOffset < now) startOffset = now

    // Ensure that there is time to schedule the events before they're due to
    // come up in the sequence.
    if (engine.isPlaying() && (startOffset - now < SCHEDULE_BUFFER_TIME_MS)) {
      log.trace { "The note would be due in ${startOffset - now} ms, so " +
                  "adding ${SCHEDULE_BUFFER_TIME_MS} to the scheduled offset" }
      startOffset += SCHEDULE_BUFFER_TIME_MS
    }

    return startOffset
  }

  suspend fun scheduleEvents(events : List<Event>, _startOffset : Int) : Int {
    val startOffset = adjustStartOffset(_startOffset)

    events.filter { it is MidiPatchEvent }.forEach {
      schedule((it as MidiPatchEvent).addOffset(startOffset))
    }

    events.filter { it is MidiPercussionEvent }.forEach {
      val event = it as MidiPercussionEvent

      if (event.offset == 0) {
        engine.midiPercussionImmediate(trackNumber)
      } else {
        engine.midiPercussionScheduled(
          trackNumber, startOffset + event.offset
        )
      }
    }

    val scheduledEvents = mutableListOf<Schedulable>()

    // It's safe to filter a List<Event> down to just the ones that are
    // Schedulable and then cast it to a List<Schedulable>.
    @Suppress("UNCHECKED_CAST")
    val immediateEvents : List<Schedulable> =
      events.map { it.addOffset(startOffset) }
            .filter { it is Schedulable }
            as List<Schedulable>

    immediateEvents.forEach { schedule(it) }

    scheduledEvents.addAll(immediateEvents)

    // For each pattern event, we...
    // * wait until right before the pattern is supposed to be played
    // * look up the pattern
    // * schedule the pattern's events
    // * add the pattern's events to `scheduledEvents`
    //
    // Patterns can include other patterns, and to support dynamically changing
    // pattern contents on the fly, we look up each pattern's contents shortly
    // before it is scheduled to play.
    //
    // Patterns can also loop indefinitely.
    //
    // To support multiple patterns looping concurrently on the same track, we
    // do the on-the-fly scheduling of each pattern in a separate coroutine and
    // collect the results in a channel.
    events.filter { it is PatternEventBase }.map {
      val event = it as PatternEventBase

      coroutineScope {
        // Returns a Deferred
        async {
          schedulePattern(event, startOffset)
        }
      }
    }.forEach { patternEvents ->
      scheduledEvents.addAll(patternEvents.await())
    }

    // Now that all the notes have been scheduled, we can start the sequencer
    // (assuming it hasn't been started already, in which case this is a no-op).
    //
    // This used to be `synchronized(engine.isPlaying()) { ... }`
    if (engine.isPlaying()) {
      engine.startSequencer()
    }

    // At this point, `noteEvents` should contain all of the notes we've
    // scheduled, including the values of patterns at the moment right before
    // they were scheduled.
    //
    // We can now calculate the latest note end offset, which shall be our new
    // `startOffset`.

    if (scheduledEvents.isEmpty())
      return _startOffset

    return scheduledEvents.map { (it as Event).endOffset() }.maxOrNull()!!
  }

  // FIXME: This needs to be in a suspend function, and
  // constructors/initializers aren't allowed to be suspendable.
  //
  // TODO: Turn this into a `start()` function and make it suspendable?
  // Or maybe have an external pseudo-constructor function that's suspendable.

  // FIXME: This is a total mess. There has to be a better way to go about this!
  //companion object {
  //  suspend fun start(trackNumber : Int, engine : SoundEngine): Track2 {
  //    val track = Track2(trackNumber, engine)

  //    coroutineScope {
  //      // Schedule events on this track in the background.
  //      async {
  //        // When new events come in on the `eventBufferQueue`, it may be the case
  //        // that previous events are still lined up to be scheduled (e.g. a
  //        // pattern is looping). When this is the case, the new events wait in
  //        // line until the previous scheduling has completed and the offset where
  //        // the next events should start is updated.
  //        val scheduling = ReentrantLock(true) // fairness enabled

  //        // FIXME, not portable
  //        while (!Thread.currentThread().isInterrupted()) {
  //          try {
  //            // FIXME, not portable
  //            val events = eventBufferQueue.take()

  //            events.filter { it is FinishLoopEvent }.forEach {
  //              thread {
  //                val event = it as FinishLoopEvent
  //                val offset = adjustStartOffset(startOffset) + event.offset
  //                val latch = engine.scheduleEvent(offset, "FinishLoop")
  //                latch.await()
  //                log.debug { "clearing active patterns" }
  //                activePatterns.clear()
  //              }
  //            }

  //            // We start a new "thread" here so that we can wait for the
  //            // opportunity to schedule new events, while the parent "thread"
  //            // continues to receive new events on the queue.
  //            async {
  //              // Wait for the previous scheduling of events to finish.
  //              //
  //              // We check the `era` now and then again when it's time to
  //              // schedule, and see if the track has been cleared in the
  //              // meantime. We only proceed if the track hasn't been cleared
  //              // (i.e. if the `era` hasn't changed).
  //              val eraBefore = synchronized(era) { era }
  //              scheduling.lock()
  //              val eraAfter = synchronized(era) { era }

  //              log.debug { "TRACK ${trackNumber}: startOffset is ${startOffset}" }

  //              try {
  //                log.debug { "eraBefore: $eraBefore; eraAfter: $eraAfter" }
  //                if (eraBefore == eraAfter) {
  //                  // Schedule events and update `startOffset` to be the offset
  //                  // at which the next events should start (after the ones we're
  //                  // scheduling here).
  //                  startOffset = scheduleEvents(events, startOffset)
  //                }
  //              } finally {
  //                // FIXME: can't use updateAndGet, that was for AtomicInteger
  //                activeTasks.updateAndGet { n -> if (n == 0) 0 else n - 1 }
  //                scheduling.unlock()
  //              }
  //            }
  //            // FIXME, not portable
  //          } catch (iex : InterruptedException) {
  //            Thread.currentThread().interrupt()
  //          }
  //        }
  //      }
  //    }

  //    return track
  //  }
  //}
}

// Trying this another way, with an external constructor function:

suspend fun newTrack(trackNumber : Int, engine : SoundEngine): Track2 {
  val track = Track2(trackNumber, engine)

  // TODO: Consider storing actual Deferreds in `activeTasks`, so that we can
  // easily `awaitActiveTasks` by calling `.await()` on all of them. Or maybe
  // there is already something similar to `Promise.all` in the stdlib?

  coroutineScope {
    // Schedule events on this track in the background.
    async {
      //// When new events come in on the `eventBufferQueue`, it may be the case
      //// that previous events are still lined up to be scheduled (e.g. a
      //// pattern is looping). When this is the case, the new events wait in
      //// line until the previous scheduling has completed and the offset where
      //// the next events should start is updated.
      //val scheduling = ReentrantLock(true) // fairness enabled

      while (true) {
        val events = track.eventBufferQueue.receive()

        events.events.filter { it is FinishLoopEvent }.forEach {
          // TODO: Move some of this logic into the Track class?
          async {
            val event = it as FinishLoopEvent
            val offset = track.adjustStartOffset(track.startOffset) + event.offset
            val job = track.engine.scheduleEvent(offset, "FinishLoop")
            job.join()
            log.debug { "Clearing active patterns" }
            track.activePatterns.clear()
          }
        }

        // When new events come in on the `eventBufferQueue`, it may be the case
        // that previous events are still lined up to be scheduled (e.g. a
        // pattern is looping). When this is the case, the new events wait in
        // line until the previous scheduling has completed and the offset where
        // the next events start is updated.
      }

      //// FIXME, not portable
      //while (!Thread.currentThread().isInterrupted()) {
      //  try {
      //    // FIXME, not portable
      //    val events = eventBufferQueue.take()

      //    events.filter { it is FinishLoopEvent }.forEach {
      //      thread {
      //        val event = it as FinishLoopEvent
      //        val offset = adjustStartOffset(startOffset) + event.offset
      //        val latch = engine.scheduleEvent(offset, "FinishLoop")
      //        latch.await()
      //        log.debug { "clearing active patterns" }
      //        activePatterns.clear()
      //      }
      //    }

      //    // We start a new "thread" here so that we can wait for the
      //    // opportunity to schedule new events, while the parent "thread"
      //    // continues to receive new events on the queue.
      //    async {
      //      // Wait for the previous scheduling of events to finish.
      //      //
      //      // We check the `era` now and then again when it's time to
      //      // schedule, and see if the track has been cleared in the
      //      // meantime. We only proceed if the track hasn't been cleared
      //      // (i.e. if the `era` hasn't changed).
      //      val eraBefore = synchronized(era) { era }
      //      scheduling.lock()
      //      val eraAfter = synchronized(era) { era }

      //      log.debug { "TRACK ${trackNumber}: startOffset is ${startOffset}" }

      //      try {
      //        log.debug { "eraBefore: $eraBefore; eraAfter: $eraAfter" }
      //        if (eraBefore == eraAfter) {
      //          // Schedule events and update `startOffset` to be the offset
      //          // at which the next events should start (after the ones we're
      //          // scheduling here).
      //          startOffset = scheduleEvents(events, startOffset)
      //        }
      //      } finally {
      //        // FIXME: can't use updateAndGet, that was for AtomicInteger
      //        activeTasks.updateAndGet { n -> if (n == 0) 0 else n - 1 }
      //        scheduling.unlock()
      //      }
      //    }
      //    // FIXME, not portable
      //  } catch (iex : InterruptedException) {
      //    Thread.currentThread().interrupt()
      //  }
      //}
    }
  }

  return track
}

package io.alda.player

import java.io.File;
import java.nio.ByteBuffer;
import java.util.Arrays;
import java.util.UUID;
import java.util.concurrent.CountDownLatch
import javax.sound.midi.MetaEventListener
import javax.sound.midi.MetaMessage
import javax.sound.midi.MidiChannel
import javax.sound.midi.MidiEvent
import javax.sound.midi.MidiMessage
import javax.sound.midi.MidiSystem
import javax.sound.midi.Sequence
import javax.sound.midi.ShortMessage
import kotlin.concurrent.thread
import mu.KotlinLogging

private val log = KotlinLogging.logger {}

// ref: https://www.csie.ntu.edu.tw/~r92092/ref/midi/
//
// There are also various sources of Java MIDI example programs that use the
// value 0x2F to create an "end of track" message.
const val MIDI_SET_TEMPO    = 0x51
const val MIDI_END_OF_TRACK = 0x2F

// ref: https://www.midi.org/specifications-old/item/table-3-control-change-messages-data-bytes-2
const val MIDI_CHANNEL_VOLUME = 7
const val MIDI_PANNING        = 10

const val DIVISION_TYPE = Sequence.PPQ
// This ought to allow for notes as fast as 512th notes at a tempo of 120 bpm,
// which is way faster than anyone should reasonably need.
//
// (4 PPQ = 4 ticks per quarter note, i.e. 16th note resolution; so 128 PPQ =
// 512th note resolution.)
const val RESOLUTION = 128

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

// The sequencer's clock will stop as soon as it reaches the end of the
// sequence, however this is not the behavior we want.
//
// We want the clock to continue indefinitely as long as the sequencer is
// playing. This is to support live coding use cases, where the sequence is
// being created in real time, and we want to preserve the live timing in the
// sequence, including any gaps where no notes were played.
//
// To get the clock to continue indefinitely, we regularly schedule a
// continuation metamessage. This value is the period of time in between these
// messages.
const val CONTINUATION_INTERVAL_MS = 1000

enum class CustomMetaMessage(val type : Int) {
  CONTINUE(0x30),
  PERCUSSION(0x31),
  EVENT(0x32)
}

// Returns the channel affected by a MidiEvent. For example, a MIDI NOTE_ON
// event affects the note on which the channel will be played.
//
// Returns null if the MidiEvent is the kind of event that does not affect any
// channel in particular.
private fun eventChannel(event : MidiEvent) : Int? {
  val msg = event.getMessage()
  if (msg !is ShortMessage) return null
  return msg.getChannel()
}

private fun isNoteOnEvent(event : MidiEvent) : Boolean {
  val msg = event.getMessage()
  return msg is ShortMessage && msg.getCommand() == ShortMessage.NOTE_ON
}

private fun isControlChangeEvent(event : MidiEvent) : Boolean {
  val msg = event.getMessage()
  return msg is ShortMessage && msg.getCommand() == ShortMessage.CONTROL_CHANGE
}

data class TempoEntry(
  val offsetMs : Int, val tempo : Float, val ticks : Long
) {}

private fun maxByteArrayValue(numBytes : Int) : Long {
  return Math.round(Math.pow(2.0, (8.0 * numBytes))) - 1
}

// In a "set tempo" metamessage, the desired tempo is expressed not in beats per
// minute (BPM), but in microseconds per quarter note (I'll abbreviate this as
// "uspq").
//
// There are 60 million microseconds in a minute, therefore the formula to
// convert BPM => uspq is 60,000,000 / BPM.
//
// Example conversion: 120 BPM / 60,000,000 = 500,000 uspq.
//
// The slower the tempo, the lower the BPM and the higher the uspq.
//
// For some reason, the MIDI spec limits the number of bytes available to
// express this number to a maximum of 3 bytes, even though there are extremely
// slow tempos (<4 BPM) that, when expressed in uspq, are numbers too large to
// fit into 3 bytes. Effectively, this means that the slowest supported tempo is
// about 3.58 BPM. That's extremely slow, so it probably won't cause any
// problems in practice, but this function will throw an assertion error below
// that tempo, so it's worth mentioning.
//
// ref:
// https://www.recordingblogs.com/wiki/midi-set-tempo-meta-message
// https://www.programcreek.com/java-api-examples/?api=javax.sound.midi.MetaMessage
// https://docs.oracle.com/javase/7/docs/api/javax/sound/midi/MetaMessage.html
// https://stackoverflow.com/a/22798636/2338327
private fun setTempoMessage(bpm : Float) : MetaMessage {
  val uspq = Math.round(60000000 / bpm)

  // Technically, a tempo less than ~3.58 BPM translates into a number of
  // microseconds per quarter note larger than 3 bytes can hold.
  //
  // Punting altogether in this scenario because it's better than overflowing
  // and secretly setting the tempo to an unexpected value.
  if (uspq > maxByteArrayValue(3)) {
    log.warn { "Tempo $bpm is < the minimum MIDI tempo of ~3.58 BPM." }
    return MetaMessage(CustomMetaMessage.CONTINUE.type, null, 0)
  }

  val byteArray = ByteBuffer.allocate(4).putInt(uspq).array()
  val msgData = Arrays.copyOfRange(byteArray, 1, 4)
  return MetaMessage(MIDI_SET_TEMPO, msgData, 3)
}

class MidiEngine {
  val sequencer = MidiSystem.getSequencer(false)
  val synthesizer = MidiSystem.getSynthesizer()
  val receiver = sequencer.getReceiver()
  val sequence = Sequence(DIVISION_TYPE, RESOLUTION)
  val track = sequence.createTrack()
  val pendingEvents = mutableMapOf<String, CountDownLatch>()

  // The sequencer automatically stops running when it reaches the end of the
  // sequence. We don't want that behavior; instead, we want to maintain our own
  // playing vs. not playing state so that if the sequencer is "playing"
  // (according to us), notes that get added are played right away.
  var isPlaying = false

  // We need to track the history of tempo changes throughout the score so that
  // we can convert millisecond values to ticks.
  val tempoItinerary = mutableListOf<TempoEntry>(
    TempoEntry(0, 120.toFloat(), 0)
  )

  // The tempo itinerary should always be in order by offset in ms.
  fun addTempoEntry(entry : TempoEntry) {
    var prev = -1
    for (itineraryEntry in tempoItinerary) {
      if (itineraryEntry.offsetMs > entry.offsetMs) break
      prev++
    }

    tempoItinerary.add(prev + 1, entry)
  }

  fun mostRecentTempoEntryByOffset(offsetMs : Double) : TempoEntry {
    return tempoItinerary.takeWhile { it.offsetMs <= offsetMs }.last()
  }

  fun mostRecentTempoEntryByTicks(ticks : Long) : TempoEntry {
    return tempoItinerary.takeWhile { it.ticks <= ticks }.last()
  }

  // MIDI sequence offset is expressed in ticks, so we can use this formula to
  // convert note offsets (which we prefer to think of in milliseconds) to
  // ticks.
  //
  // The conversion logic is complicated because the physical duration of a tick
  // varies depending on the tempo, and this has a cascading effect when it
  // comes to scheduling an event. We must consider not only the current tempo,
  // but the entire history of tempo changes in the score.
  private fun msToTicks(offsetMs : Double): Long {
    if (offsetMs == 0.0) return 0

    val tempoEntry = mostRecentTempoEntryByOffset(offsetMs)
    // source: https://stackoverflow.com/a/2038364/2338327
    val msPerTick = 60000.0 / (tempoEntry.tempo * RESOLUTION)
    val msDelta = offsetMs - tempoEntry.offsetMs
    val ticksDelta = msDelta / msPerTick
    return Math.round(tempoEntry.ticks + ticksDelta)
  }

  private fun ticksToMs(ticks : Long): Double {
    val tempoEntry = mostRecentTempoEntryByTicks(ticks)
    val msPerTick = 60000.0 / (tempoEntry.tempo * RESOLUTION)
    val ticksDelta = ticks - tempoEntry.ticks
    val msDelta = ticksDelta * msPerTick
    return tempoEntry.offsetMs + msDelta
  }

  fun currentOffset(): Double {
    return ticksToMs(sequencer.getTickPosition())
  }

  fun setTempo(offsetMs : Int, bpm : Float) {
    val ticks = msToTicks(offsetMs * 1.0)
    addTempoEntry(TempoEntry(offsetMs, bpm, ticks))
    track.add(MidiEvent(setTempoMessage(bpm), ticks))
  }

  private fun scheduleMidiMsg(offset : Int, midiMsg : MidiMessage) {
    track.add(MidiEvent(midiMsg, msToTicks(offset * 1.0)))
  }

  private fun scheduleShortMsg(
    offset : Int, command : Int, channel : Int, data1: Int, data2: Int
  ) {
    scheduleMidiMsg(offset, ShortMessage(command, channel, data1, data2))
  }

  private fun scheduleMetaMsg(
    offset : Int, msgType : CustomMetaMessage, msgData : ByteArray?,
    length : Int
  ) {
    scheduleMidiMsg(offset, MetaMessage(msgType.type, msgData, length))
  }

  private fun scheduleMetaMsg(offset : Int, type : CustomMetaMessage) {
    scheduleMetaMsg(offset, type, null, 0)
  }

  private fun scheduleContinueMsg(offset : Int) {
    scheduleMetaMsg(offset, CustomMetaMessage.CONTINUE)
  }

  init {
    log.info { "Initializing MIDI sequencer..." }
    sequencer.open()
    sequencer.setSequence(sequence)
    sequencer.setTickPosition(0)

    log.info { "Initializing MIDI synthesizer..." }
    // NB: This blocks for about a second.
    synthesizer.open()

    // Transmit messages from the sequencer to the synthesizer.
    sequencer.getTransmitter().setReceiver(synthesizer.getReceiver())

    sequencer.addMetaEventListener(MetaEventListener { msg ->
      when (val msgType = msg.getType()) {
        CustomMetaMessage.CONTINUE.type -> {
          synchronized(isPlaying) {
            if (isPlaying) sequencer.start()
          }
        }

        CustomMetaMessage.PERCUSSION.type -> {
          val trackNumber = msg.getData().first().toInt()
          track(trackNumber).useMidiPercussionChannel()
        }

        CustomMetaMessage.EVENT.type -> {
          val pendingEvent = String(msg.getData())

          synchronized(pendingEvents) {
            pendingEvents.get(pendingEvent)?.also { latch ->
              latch.countDown()
              pendingEvents.remove(pendingEvent)
            } ?: run {
              log.error { "$pendingEvent latch not found!" }
            }
          }
        }

        MIDI_END_OF_TRACK -> {
          // This metamessage is sent automatically when the end of the sequence
          // is reached.
        }

        MIDI_SET_TEMPO -> {
          // This metamessage is handled by the Sequencer out of the box.
        }

        else -> {
          log.warn { "MetaMessage type $msgType not implemented." }
        }
      }
    })

    thread {
      while (!Thread.currentThread().isInterrupted()) {
        try {
          synchronized(isPlaying) {
            if (isPlaying) {
              val now = currentOffset()
              val future = now + (CONTINUATION_INTERVAL_MS * 2)
              scheduleContinueMsg(Math.round(future).toInt())
              sequencer.start()
            }
          }

          Thread.sleep(CONTINUATION_INTERVAL_MS.toLong())
        } catch (iex : InterruptedException) {
          Thread.currentThread().interrupt()
        }
      }
    }

    // for debugging
    thread {
      while (!Thread.currentThread().isInterrupted()) {
        try {
          log.trace { "${if (isPlaying) "PLAYING; " else ""}current offset: ${currentOffset()}" }
          Thread.sleep(500)
        } catch (iex : InterruptedException) {
          Thread.currentThread().interrupt()
        }
      }
    }
  }

  fun startSequencer() {
    synchronized(isPlaying) {
      sequencer.start()
      isPlaying = true
    }
  }

  fun stopSequencer() {
    synchronized(isPlaying) {
      sequencer.stop()
      isPlaying = false
    }
  }

  fun patch(offset : Int, channel : Int, patch : Int) {
    scheduleShortMsg(offset, ShortMessage.PROGRAM_CHANGE, channel, patch, 0)
  }

  // Immediately configure the track to use a percussion channel. That way, any
  // note messages in the same bundle will be scheduled on a percussion channel.
  fun percussionImmediate(trackNumber : Int) {
    track(trackNumber).useMidiPercussionChannel()
  }

  // Configure a track to be a percussion track in the future by scheduling a
  // metamessage that will do the above once the offset is reached.
  fun percussionScheduled(trackNumber : Int, offset : Int) {
    scheduleMetaMsg(
      offset,
      CustomMetaMessage.PERCUSSION,
      listOf(trackNumber.toByte()).toByteArray(),
      1
    )
  }

  fun note(
    startOffset : Int, endOffset : Int, channel : Int, noteNumber : Int,
    velocity : Int
  ) {
    log.trace { "channel ${channel}: scheduling note from ${startOffset} to ${endOffset}" }

    scheduleShortMsg(
      startOffset, ShortMessage.NOTE_ON, channel, noteNumber, velocity
    )
    scheduleShortMsg(
      endOffset, ShortMessage.NOTE_OFF, channel, noteNumber, velocity
    )
  }

  fun volume(offset : Int, channel : Int, volume : Int) {
    scheduleShortMsg(
      offset, ShortMessage.CONTROL_CHANGE, channel, MIDI_CHANNEL_VOLUME, volume
    )
  }

  fun panning(offset : Int, channel : Int, panning : Int) {
    scheduleShortMsg(
      offset, ShortMessage.CONTROL_CHANGE, channel, MIDI_PANNING, panning
    )
  }

  // Schedules an event to occur at the desired offset.
  //
  // Returns a CountDownLatch that will count down from 1 to 0 when the event is
  // scheduled to occur.
  //
  // This will signal the Player to perform a particular action at just the
  // right time.
  fun scheduleEvent(offset : Int, eventName : String) : CountDownLatch {
    val latch = CountDownLatch(1)
    val pendingEvent = eventName + "::" + UUID.randomUUID().toString()

    synchronized(pendingEvents){
      pendingEvents.put(pendingEvent, latch)
    }

    val msgData = pendingEvent.toByteArray()
    scheduleMetaMsg(offset, CustomMetaMessage.EVENT, msgData, msgData.size)
    return latch
  }

  private fun withChannel(channelNumber : Int, f : (MidiChannel) -> Unit) {
    synthesizer.getChannels()[channelNumber]?.also { channel ->
      f(channel)
    } ?: run {
      log.warn { "MIDI channel $channelNumber is null." }
    }
  }

  fun clearChannel(channelNumber : Int) {
    withChannel(channelNumber) { channel ->
      channel.allNotesOff()
      channel.allSoundOff()
    }

    val channelEvents = mutableListOf<MidiEvent>()
    for (i in 0..(track.size() - 1)) {
      val event = track.get(i)
      if (eventChannel(event) == channelNumber) {
        channelEvents.add(event)
      }
    }

    // To preserve the current tick position, we replace each message with a
    // no-op CONTINUE message instead of simply removing it.
    channelEvents.forEach {
      track.add(MidiEvent(
        MetaMessage(CustomMetaMessage.CONTINUE.type, null, 0),
        it.getTick())
      )
      track.remove(it)
    }
  }

  fun muteChannel(channelNumber : Int) {
    withChannel(channelNumber) { it.setMute(true) }
  }

  fun unmuteChannel(channelNumber : Int) {
    withChannel(channelNumber) { it.setMute(false) }
  }

  fun export(filepath : String) {
    // We make a copy of the sequence so that we can shift the tick position of
    // each event in the sequence back such that the first event starts at tick
    // position 0. This is to compensate for the SCHEDULE_BUFFER_TIME_MS buffer
    // time that tends to happen at the beginning of the sequence.
    val sequenceCopy = Sequence(DIVISION_TYPE, RESOLUTION)
    val trackCopy = sequenceCopy.createTrack()

    var earliestOffset = Long.MAX_VALUE
    val trackEvents = mutableListOf<MidiEvent>()
    for (i in 0..(track.size() - 1)) {
      val event = track.get(i)
      trackEvents.add(event)
      if (isNoteOnEvent(event) || isControlChangeEvent(event))
        earliestOffset = minOf(earliestOffset, event.getTick())
    }

    trackEvents.forEach { event ->
      val msgCopy = event.getMessage().clone() as MidiMessage
      val ticks = maxOf(0, event.getTick() - earliestOffset)
      trackCopy.add(MidiEvent(msgCopy, ticks))
    }

    val midiFileType = 0
    MidiSystem.write(sequenceCopy, midiFileType, File(filepath))
  }
}


package io.alda.player

import java.util.UUID;
import java.util.concurrent.CountDownLatch
import javax.sound.midi.MetaEventListener
import javax.sound.midi.MetaMessage
import javax.sound.midi.MidiEvent
import javax.sound.midi.MidiMessage
import javax.sound.midi.MidiSystem
import javax.sound.midi.Sequence
import javax.sound.midi.ShortMessage
import kotlin.concurrent.thread

const val DIVISION_TYPE = Sequence.SMPTE_24
const val RESOLUTION = 2

// Example:
// * SMPTE_24 means 24 frames per second
// * A resolution of 2 means 2 ticks per frame.
// * Therefore, there are 48 ticks per second.
const val TICKS_PER_SECOND : Double = DIVISION_TYPE * RESOLUTION * 1.0

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
  return (msg as ShortMessage).getChannel()
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
    println("Initializing MIDI sequencer...")
    sequencer.open()
    sequencer.setSequence(sequence)
    sequencer.setTickPosition(0)

    println("Initializing MIDI synthesizer...")
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
              println("ERROR: $pendingEvent latch not found!")
            }
          }
        }

        0x2f -> {
          // This metamessage is sent automatically when the end of the sequence
          // is reached.
        }

        else -> {
          println("WARN: MetaMessage type $msgType not implemented.")
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
          println("${if (isPlaying) "PLAYING; " else ""}current offset: ${currentOffset()}")
          Thread.sleep(500)
        } catch (iex : InterruptedException) {
          Thread.currentThread().interrupt()
        }
      }
    }
  }

  // MIDI sequence offset is expressed in ticks, so we can use this formula to
  // convert note offsets (which we prefer to think of in milliseconds) to
  // ticks.
  private fun msToTicks(ms : Double): Long {
    return Math.round((ms / 1000.0) * TICKS_PER_SECOND)
  }

  private fun ticksToMs(ticks : Long): Double {
    return (ticks / TICKS_PER_SECOND) * 1000
  }

  fun currentOffset(): Double {
    return ticksToMs(sequencer.getTickPosition())
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

  fun percussion(offset : Int, trackNumber : Int) {
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
    println("channel ${channel}: scheduling note from ${startOffset} to ${endOffset}")

    scheduleShortMsg(
      startOffset, ShortMessage.NOTE_ON, channel, noteNumber, velocity
    )
    scheduleShortMsg(
      endOffset, ShortMessage.NOTE_OFF, channel, noteNumber, velocity
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

  fun clearChannel(channelNumber : Int) {
    synthesizer.getChannels()[channelNumber]?.also { channel ->
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
    synthesizer.getChannels()[channelNumber]?.also { channel ->
      channel.setMute(true)
    }
  }

  fun unmuteChannel(channelNumber : Int) {
    synthesizer.getChannels()[channelNumber]?.also { channel ->
      channel.setMute(false)
    }
  }
}


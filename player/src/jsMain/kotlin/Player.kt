@file:OptIn(DelicateCoroutinesApi::class)

package io.alda.player

import kotlin.time.Duration
import kotlinx.coroutines.*
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.selects.*
import kotlinx.browser.window
import kotlin.js.Json
import mu.KotlinLogging

private val log = KotlinLogging.logger {}

val stateManager = NoOpStateManager()

enum class Signal {
  START, STOP
}

class Player() {
  var _osc : Json? = null
  private fun osc() : dynamic {
    if (_osc != null) {
      return _osc!!
    }

    val oscInWindow = window.asDynamic().osc

    if (oscInWindow == null) {
      throw RuntimeException("window.osc not defined. osc.js is required")
    }

    _osc = oscInWindow
    return _osc!!
  }

  var isRunning = false

  // We should probably make this buffered. If I'm understanding the docs right,
  // the default when you construct a channel without an argument is that it's a
  // "rendezvous" channel, where sends and receives block until the other side
  // is ready. I think that means if processing the message takes a while, then
  // if you send another message in the meantime, it won't actually end up in
  // the queue until the first message is done processing.
  //
  // But I'm also not sure if this will really be a problem or not, since we are
  // taking care to do work in small, suspendable chunks. So this might be
  // totally fine.
  //
  // TODO: Test with and without a buffer size argument.
  private val updatesChannel = Channel<UpdatesSpec>()

  private val signals = Channel<Signal>()

  @JsName("start")
  fun start() {
    log.info { "Starting player." }

    // This is required in order to establish a coroutine scope so that we can
    // call the suspend function `send`.
    //
    // It also happens to return a JS promise, so if we wanted to return the
    // promise here, we could.
    GlobalScope.promise {
      signals.send(Signal.START)
    }
  }

  @JsName("stop")
  fun stop() {
    log.info { "Stopping player." }

    // This is required in order to establish a coroutine scope so that we can
    // call the suspend function `send`.
    //
    // It also happens to return a JS promise, so if we wanted to return the
    // promise here, we could.
    GlobalScope.promise {
      signals.send(Signal.STOP)
    }
  }

  private suspend fun processMessages() {
    while (true) {
      if (!isRunning) {
        when (signals.receive()) {
          Signal.START -> { isRunning = true }
          Signal.STOP -> { continue }
        }
      }

      select<Unit> {
        signals.onReceive { signal ->
          if (signal == Signal.STOP) {
            isRunning = false
          }
        }

        updatesChannel.onReceive { updates ->
          log.debug { "Handling updates: ${updates}" }
        }
      }
    }
  }

  // Given an OSC packet that might be a bundle or a message, recursively
  // flattens all bundles and returns a list of messages.
  private fun parseMessages(oscPacket : Json) : List<Message> {
    val pkt = oscPacket.asDynamic()

    if (pkt.packets == null) {
      if (pkt.address == null || pkt.args == null) {
        throw RuntimeException("Unexpected OSC packet: ${JSON.stringify(pkt)}")
      }

      // Is there really not an easier way to turn a JS array into a Kotlin
      // ArrayList?
      val args = mutableListOf<Any>()
      for (arg in pkt.args) {
        args.add(arg)
      }

      return listOf(Message(pkt.address, args))
    }

    val messages = mutableListOf<Message>()

    for (packet in pkt.packets) {
      messages.addAll(parseMessages(packet))
    }

    return messages
  }

  @JsName("sendOSCBytes")
  fun sendOSCBytes(bytes : ByteArray) {
    val messages = parseMessages(osc().readPacket(bytes, jsObject {}))

    // This is required in order to establish a coroutine scope so that we can
    // call the suspend function `send`.
    //
    // It also happens to return a JS promise, so if we wanted to return the
    // promise here, we could.
    GlobalScope.promise {
      updatesChannel.send(UpdatesSpec(messages, stateManager))
    }
  }

  init {
    // Check for the presence of osc.js. This will throw an exception if osc.js
    // isn't available in the window.
    osc()

    // This just sets `isRunning` to true so that the `processMessages`
    // coroutine will process messages. When `player.stop()` is invoked,
    // `isRunning` will be set to false, which results in the coroutine holding
    // off on processing messages until `player.start()` is invoked again.
    start()

    // Process messages in a loop within a coroutine. JavaScript is
    // single-threaded, but we want to avoid blocking UI rendering, etc. so the
    // work of this coroutine is done in suspendable chunks, so that the JS
    // event loop has opportunities to do other things in between.
    //
    // NOTE: This returns a JS promise. I thought about exposing it somehow
    // through the AldaPlayer class so that end users could interact with it,
    // but I couldn't think of anything useful that they could do with the
    // promise. Promises are uncancellable, and if you await this particular
    // promise, you'll be waiting forever because `processMessages` runs a
    // `while (true) { ... }` infinite loop. So the fact that it's a promise is
    // just an implementation detail of launching a coroutine in the JS runtime.
    GlobalScope.promise {
      processMessages()
    }
  }
}

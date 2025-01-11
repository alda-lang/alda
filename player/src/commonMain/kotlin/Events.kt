package io.alda.player

import mu.KotlinLogging

private val log = KotlinLogging.logger {}

// NOTE: This code is written in a style that's unconventional for Java/Kotlin.
// Instead of defining a class hierarchy for all of the different kinds of
// events and interfaces for things that have offsets and things that can be
// scheduled (I actually did this before rewriting it using the current
// approach), I chose to adopt a Clojuresque approach and model the events as
// data, in the form of Maps.
//
// The trade-off is that the compiler doesn't know anything about the structure
// of these maps and it can't help us avoid certain mistakes at compile-time.
// But I think it's worth it, because the code is (IMHO) a lot easier to
// understand and maintain, and it's easy enough to test and find issues at
// runtime.
typealias Event = Map<String, Any>

fun addOffset(event: Event, amount: Int): Event {
  return update(event, "offset", { (it as Int) + amount })
}

fun endOffset(event: Event): Int {
  event["duration"]?.also {
    return (event["offset"] as Int) + (it as Int)
  }

  return 0
}

fun isDone(event: Event, iteration: Int): Boolean {
  event["times"]?.also {
    return iteration > it as Int
  }

  return false
}

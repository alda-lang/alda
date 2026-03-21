# Attempt at a Kotlin Multiplatform / In-Browser Player

**Branch:** `kotlin-js`

This document describes an attempt to extend the Alda player to support running
in a web browser, using [Kotlin Multiplatform (KMP)][kmp] to share code between
the existing JVM-based CLI player and a new JS-based in-browser player.

The effort was abandoned partway through. This document captures the goal, the
approach, what was accomplished, where things got stuck, and guidance for
picking it up again in the future.

[kmp]: https://kotlinlang.org/docs/multiplatform.html

## Goal

The idea was to make it possible to embed an Alda player in a web page — i.e.
a visitor to the page could play back Alda scores without installing anything.
The rough architecture envisioned:

1. The Go client (`alda`) compiled to WebAssembly (`alda.wasm`), running in the
   browser and producing OSC messages. **This part was completed and shipped —
   `alda.wasm` is included in current Alda releases.**
2. A JS build of the Kotlin player (`alda-player.js`), running in the same
   browser tab and receiving those OSC messages, then playing audio via the Web
   Audio API. **This part is what was never finished.**

The two would communicate via [osc.js][oscjs], a JavaScript OSC library that
supports passing OSC messages directly between JS objects (without a network
socket).

[oscjs]: https://github.com/colinbdclark/osc.js/

## What Changed

### Project restructure (Kotlin Multiplatform)

The original player code lived entirely in `src/main/kotlin/`. The first major
step was restructuring it into a proper KMP layout:

- `src/commonMain/kotlin/` — platform-agnostic code shared by both targets
- `src/jvmMain/kotlin/` — JVM-specific code (MIDI, file system, threads)
- `src/jsMain/kotlin/` — JS-specific code (Web Audio API, browser APIs)

Existing files were renamed to reflect their scope:

| Before | After |
|--------|-------|
| `MidiEngine.kt` | `jvmMain/JvmSoundEngine.kt` |
| `StateManager.kt` | `jvmMain/FileBasedStateManager.kt` |
| `Parser.kt` | `commonMain/Instructions.kt` |

Gradle was upgraded from 6.8.3 → 7.3.3 → 8.10, and Kotlin multiplatform from
1.6.10 (though a 2.1.0 upgrade was started and left in-progress in
`build.gradle.kts`).

### Platform-agnostic interfaces

Two interfaces were extracted into `commonMain` so that both targets can code
against them:

**`SoundEngine`** — the audio backend:

```kotlin
interface SoundEngine {
  fun isPlaying() : Boolean
  fun startSequencer()
  fun stopSequencer()
  fun currentOffset() : Double
  fun midiNote(startOffset, endOffset, channel, noteNumber, velocity)
  fun midiClearChannel(channelNumber)
  fun midiMuteChannel(channelNumber)
  fun midiUnmuteChannel(channelNumber)
  fun midiPanning(offset, channel, panning)
  fun midiPatch(offset, channel, patch)
  fun midiPercussionImmediate(trackNumber)
  fun midiPercussionScheduled(trackNumber, offset)
  fun midiVolume(offset, channel, volume)
  fun scheduleEvent(offset, eventName) : Job
}
```

The JVM implementation is `JvmSoundEngine` (the existing MIDI engine). A JS
implementation using the [Web Audio API][webaudio] was the next thing to write —
this is a significant piece of missing work (see below).

**`StateManager`** — lifecycle management (expiration, active/ready state):

```kotlin
interface StateManager {
  fun delayExpiration()
  fun markActive()
  fun markReady()
}
```

The JVM implementation is `FileBasedStateManager` (reads/writes state files).
A `NoOpStateManager` is provided for JS, since browser-based players don't
need the same file-based process lifecycle management.

[webaudio]: https://developer.mozilla.org/en-US/docs/Web/API/Web_Audio_API

### OSC message parsing moved to `commonMain`

`Instructions.kt` (née `Parser.kt`) handles parsing incoming OSC messages and
translating them into player actions. It was moved to `commonMain` so both the
JVM and JS players can share the same message-parsing logic.

As part of this, the action enums were renamed with an `Enum` suffix
(`SystemActionEnum`, `TrackActionEnum`, `PatternActionEnum`) to avoid a naming
conflict with a new `Update` sealed class:

```kotlin
sealed class Update {
  data class SystemAction(val action: SystemActionEnum) : Update()
  data class TrackAction(val track: Int, val action: TrackActionEnum) : Update()
  data class PatternAction(val pattern: String, val action: PatternActionEnum) : Update()
  data class SystemEvents(val events: List<Event>) : Update()
  data class TrackEvents(val track: Int, val events: List<Event>) : Update()
  data class PatternEvents(val pattern: String, val events: List<Event>) : Update()
}
```

This hierarchy provides a cleaner, typed way to represent the set of updates
derived from a batch of OSC messages, with the intent of eventually dispatching
them through common code to whichever `SoundEngine` implementation is in use.

### `UpdatesSpec` rename

The `Updates` class was renamed to `UpdatesSpec` to better reflect that it
represents a specification of updates to be applied, not the updates themselves.

### `JvmSoundEngine` improvements

While refactoring, a couple of improvements were made to the JVM engine:

- Pending events (used to synchronize scheduling with playback) were switched
  from `CountDownLatch` to kotlinx-coroutines `Job` — a more Kotlin-idiomatic
  approach and one that works cross-platform.
- `_isPlaying` access was wrapped in `synchronized` blocks for thread safety.

### Browser test page

A simple test page was added at `player/test-page/` to exercise the JS build:

- `index.html` — loads `osc.js` and `alda-player.js`, instantiates an
  `AldaPlayer`, and has a button to call `start()`.
- `osc-browser.js` — a browser-compatible build of osc.js.
- `serve` — a tiny script to serve the test page locally.

The test page was used to verify that the JS build loaded correctly and that
`AldaPlayer.start()` could be called from JavaScript.

### JS `Player` skeleton (`jsMain/Player.kt`)

A skeleton JS `Player` class was written, mirroring the structure of the JVM
player but using coroutines and Kotlin channels instead of threads and blocking
queues:

```kotlin
class Player() {
  private val updatesChannel = Channel<UpdatesSpec>()
  private val signals = Channel<Signal>()

  fun start() { GlobalScope.promise { signals.send(Signal.START) } }
  fun stop()  { GlobalScope.promise { signals.send(Signal.STOP)  } }
  // ...
}
```

It uses `window.osc` (osc.js) to receive OSC messages from the browser and
funnel them into `updatesChannel` for processing. The processing loop was not
yet implemented.

### `Track2` — coroutines-based track scheduling (`commonMain/Track.kt`)

This is the most significant piece of new work. The existing `Track` class (in
`jvmMain/Player.kt`) uses Java threads, `LinkedBlockingQueue`, `ReentrantLock`,
and `synchronized` — none of which are available in Kotlin/JS.

`Track2` is a rewrite of `Track` in `commonMain`, using:

- `kotlinx.coroutines.channels.Channel` instead of `LinkedBlockingQueue`
- `coroutineScope { async { ... } }` instead of `thread { ... }`
- `Job` instead of `CountDownLatch` for event synchronization

The key insight is in a comment in the code:

> `synchronized()` doesn't work cross-platform. Guidance is to write common
> code such that it does not use shared state. Maybe I won't need `synchronized`
> if I convert all of the multithreaded stuff to use coroutines instead?

Most of `Track2`'s scheduling logic was ported from `Track`, though the
implementation was left unfinished — in particular, the event-scheduling loop
inside `newTrack()` is incomplete and the old Java-thread-based version is
preserved in comments alongside the new coroutine-based version for reference.

## Where It Got Stuck

The work stalled at the intersection of two hard problems:

1. **The JS sound engine was never written.** `JvmSoundEngine` wraps the Java
   MIDI sequencer API extensively. A JS equivalent would need to implement the
   same `SoundEngine` interface using the Web Audio API. This is a substantial
   implementation effort — scheduling notes with precise timing, managing MIDI
   channels, handling tempo changes — all without the Java MIDI sequencer.

2. **The coroutines-based `Track2` was incomplete.** The scheduling loop inside
   `newTrack()` has the structure but not the full logic — the part where new
   events wait for prior scheduling to finish (formerly done with a
   `ReentrantLock`) had not yet been translated to a coroutine-friendly
   equivalent. The old code is preserved in comments.

3. **The build kept breaking.** As Kotlin and Gradle versions evolved, the KMP
   build configuration needed to keep up. The `build.gradle.kts` was left
   in-progress with some things commented out during an attempted upgrade to
   Kotlin 2.1.0 / Gradle 8.10.

## Suggested Path Forward

If picking this up again, the recommended order of attack:

1. **Fix the build.** Get the KMP project to compile for both JVM and JS
   targets cleanly. The `build.gradle.kts` needs the Kotlin 2.1.0 DSL changes
   completed (the `fatJar` task, JVM target configuration, etc.).

2. **Finish `Track2`.** The scheduling loop in `newTrack()` needs the
   "wait-for-prior-scheduling" logic ported to coroutines. The commented-out
   old code in `Track.kt` is the reference. The key challenge is replacing
   `ReentrantLock` — look at `Mutex` from kotlinx-coroutines.

3. **Write `JsSoundEngine`.** Implement the `SoundEngine` interface using the
   Web Audio API. This is the biggest remaining task. Key things to figure out:
   - Scheduling notes at precise future timestamps using `AudioContext.currentTime`
   - Managing 16 MIDI channels (likely as a map of `OscillatorNode` /
     `AudioBufferSourceNode` instances, or using a JS MIDI library like [JZZ][jzz])
   - Handling tempo changes and offset-to-time conversion

4. **Wire up the JS `Player`.** Complete the OSC message processing loop in
   `jsMain/Player.kt` and connect it to a `JsSoundEngine` instance.

5. **Test end-to-end.** Use the existing test page (`player/test-page/`) with
   osc.js to send OSC messages from JS and verify that audio plays.

[jzz]: https://jazz-soft.net/doc/JZZ/

## Notes on osc.js and OSC Transport

The test page uses [osc.js][oscjs] for OSC in the browser. The approach taken
was to load `osc.js` as a global (`window.osc`) and access it from Kotlin/JS
via `window.asDynamic().osc`. JZZ was also briefly experimented with for MIDI
playback (see the `JZZ test code` commit), but not pursued further.

It's worth reconsidering the OSC transport mechanism when picking this up:
osc.js can operate over WebSockets (for communicating with a server), as a
relay between JS objects (for same-page communication with `alda.wasm`), or
via raw UDP in Node.js. The in-browser use case would need the relay approach
if `alda.wasm` is in the same page, or WebSockets if communicating with a
remote client.

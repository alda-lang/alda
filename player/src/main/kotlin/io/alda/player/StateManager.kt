package io.alda.player

import com.beust.klaxon.Klaxon
import java.io.File
import java.nio.file.Paths
import kotlin.random.Random
import kotlin.concurrent.thread

val json = Klaxon()

class PlayerState(val port : Int, var expiry : Long, var condition : String)

class StateManager(val port : Int) {
  val thread = thread(start = false) {
    writeStateFile()

    while (!Thread.currentThread().isInterrupted()) {
      try {
        // Periodically "touch" the state file to update its last modified time.
        //
        // State files with sufficiently old last modified times are eligible to
        // be cleaned up so that we don't end up with old, unused state files
        // hanging around.
        Thread.sleep(10000)
        stateFile.setLastModified(System.currentTimeMillis())
      } catch (iex : InterruptedException) {
        Thread.currentThread().interrupt();
      }
    }
  }

  fun start() {
    thread.start()
  }

  fun stop() {
    thread.interrupt()
  }

  // A player process shuts down after a random length of inactivity between 5
  // and 10 minutes. This helps to ensure that a bunch of old player processes
  // aren't left hanging around, running idle in the background.
  val inactivityTimeoutMs = Random.nextInt(5 * 60000, 10 * 60000)

  val state = PlayerState(
    port,
    System.currentTimeMillis() + inactivityTimeoutMs,
    "new")

  val stateFilePath = Paths.get(
    projDirs.cacheDir, "state", "players", playerVersion, playerId + ".json"
  ).toString()

  val stateFile = File(stateFilePath)

  init {
    stateFile.getParentFile().mkdirs()
    stateFile.createNewFile()
    stateFile.deleteOnExit()
  }

  fun writeStateFile() {
    stateFile.writeText(json.toJsonString(state))
  }

  fun isExpired() = System.currentTimeMillis() > state.expiry

  fun delayExpiration(pointInTimeMs : Long) {
    synchronized(state) {
      val newExpiry = pointInTimeMs + inactivityTimeoutMs
      if (newExpiry > state.expiry) {
        state.expiry = newExpiry
        writeStateFile()
      }
    }
  }

  fun delayExpiration() {
    delayExpiration(System.currentTimeMillis())
  }

  fun markUsed() {
    synchronized(state) {
      if (state.condition != "used") {
        state.condition = "used"
        writeStateFile()
      }
    }
  }
}

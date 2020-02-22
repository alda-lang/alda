package io.alda.player

import com.beust.klaxon.Klaxon
import java.io.File
import java.nio.file.Paths
import java.time.Duration
import java.time.Instant
import java.util.Date
import kotlin.random.Random
import kotlin.concurrent.thread
import mu.KotlinLogging

private val json = Klaxon()
private val log = KotlinLogging.logger {}

class PlayerState(val port : Int, var expiry : Long, var condition : String)

class StateManager(val port : Int) {
  val thread = thread(start = false) {
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

  val stateFilesDir =
    Paths.get(projDirs.cacheDir, "state", "players").toString()

  val stateFilePath =
    Paths.get(stateFilesDir, playerVersion, playerId + ".json").toString()

  val stateFile = File(stateFilePath)

  fun writeStateFile() {
    stateFile.writeText(json.toJsonString(state))
  }

  init {
    stateFile.getParentFile().mkdirs()
    stateFile.createNewFile()
    stateFile.deleteOnExit()
    writeStateFile()

    // Clean up the state files directory. This is important because even though
    // we have a shutdown hook that removes a player's state file before
    // exiting, it won't run in all scenarios (e.g. OOM error, kill -9).
    log.info { "Cleaning up stale files in ${stateFilesDir}..." }

    File(stateFilesDir).walkTopDown().filter { it.isFile() }.forEach {
      val lastModified = Instant.ofEpochMilli(it.lastModified())
      val now = Instant.now()
      val age = Duration.between(lastModified, now)
      val maxAge = Duration.ofMinutes(10)

      if (age.compareTo(maxAge) > 0) {
        log.debug {
          "Deleting stale state file ${it.getAbsolutePath()} " +
          "(last update was ${age} ago)"
        }
        it.delete()
      }
    }

    File(stateFilesDir).walkBottomUp().filter { it.isDirectory() }.forEach {
      if (it.list().size == 0) {
        log.debug { "Deleting empty directory ${it.getAbsolutePath()}" }
        it.delete()
      }
    }
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

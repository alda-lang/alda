package io.alda.player

import com.beust.klaxon.Klaxon
import java.io.File
import java.nio.file.Paths
import java.lang.management.ManagementFactory
import java.time.Duration
import java.time.Instant
import java.util.Date
import kotlin.random.Random
import kotlin.concurrent.thread
import mu.KotlinLogging

private val json = Klaxon()
private val log = KotlinLogging.logger {}

class PlayerState(
  val port : Int, var expiry : Long, var state : String, val pid: Long?
)

class FileBasedStateManager(val port : Int) : StateManager {
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

  fun currentPid() : Long? {
    // NOTE: Starting in Java 9, you can just do this to get the current PID:
    //
    // ProcessHandle.current().pid()
    //
    // But we are targeting Java 8 compatibility, so we have to do this
    // craziness instead:
    val beanName = ManagementFactory.getRuntimeMXBean().getName()
    return beanName.split("@")[0].toLongOrNull()
  }

  val state = PlayerState(
    port,
    System.currentTimeMillis() + inactivityTimeoutMs,
    "starting",
    currentPid()
  )

  val stateFilesDir =
    Paths.get(projDirs.cacheDir, "state", "players").toString()

  val replServerStateFilesDir =
    Paths.get(projDirs.cacheDir, "state", "repl-servers").toString()

  val stateFilePath =
    Paths.get(stateFilesDir, playerVersion, playerId + ".json").toString()

  val stateFile = File(stateFilePath)

  fun writeStateFile() {
    stateFile.writeText(json.toJsonString(state))
  }

  fun cleanUpStaleStateFiles(dir : String) {
    // Clean up the state files directory. This is important because even though
    // we have a shutdown hook that removes a player's state file before
    // exiting, it won't run in all scenarios (e.g. OOM error, kill -9).
    log.info { "Cleaning up stale files in ${dir}..." }

    File(dir).walkTopDown().filter { it.isFile() }.forEach {
      val lastModified = Instant.ofEpochMilli(it.lastModified())
      val now = Instant.now()
      val age = Duration.between(lastModified, now)
      val maxAge = Duration.ofMinutes(2)

      if (age.compareTo(maxAge) > 0) {
        log.debug {
          "Deleting stale state file ${it.getAbsolutePath()} " +
          "(last update was ${age} ago)"
        }
        it.delete()
      }
    }

    File(dir).walkBottomUp().filter { it.isDirectory() }.forEach {
      if (it.list().size == 0) {
        log.debug { "Deleting empty directory ${it.getAbsolutePath()}" }
        it.delete()
      }
    }
  }

  init {
    stateFile.getParentFile().mkdirs()
    stateFile.createNewFile()
    stateFile.deleteOnExit()
    writeStateFile()
    cleanUpStaleStateFiles(stateFilesDir)
    cleanUpStaleStateFiles(replServerStateFilesDir)
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

  override fun delayExpiration() {
    delayExpiration(System.currentTimeMillis())
  }

  private fun setState(str : String) {
    synchronized(state) {
      if (state.state != str) {
        state.state = str
        writeStateFile()
      }
    }
  }

  override fun markReady() = setState("ready")
  override fun markActive() = setState("active")
}

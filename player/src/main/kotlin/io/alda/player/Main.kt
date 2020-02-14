package io.alda.player

import com.github.ajalt.clikt.core.CliktCommand
import com.github.ajalt.clikt.core.subcommands
import com.github.ajalt.clikt.parameters.options.flag
import com.github.ajalt.clikt.parameters.options.option
import com.github.ajalt.clikt.parameters.options.required
import com.github.ajalt.clikt.parameters.types.int
import io.github.soc.directories.ProjectDirectories
import kotlin.random.Random
import kotlin.concurrent.thread
import kotlin.streams.asSequence
import kotlin.system.exitProcess
import mu.KotlinLogging
import mu.KLogger
import org.apache.logging.log4j.core.config.Configurator
import org.apache.logging.log4j.Level

var logger : KLogger? = null

var isRunning = true

// adapted from https://stackoverflow.com/a/46944275/2338327
private fun generateId() : String {
  val source = "abcdefghijklmnopqrstuvwxyz"
  return (1..3).map { _ -> source.get(Random.nextInt(0, source.length)) }
               .joinToString("")
}

val playerId = generateId()

val projDirs = ProjectDirectories.from("io", "alda", "alda")

class Info : CliktCommand(
  help = "Print useful information including the version and log path"
) {
  override fun run() {
    println("alda-player X.X.X") // TODO: print the actual version
    println("log path: ${projDirs.cacheDir}")
  }
}

// A player process shuts down after a random length of inactivity between 15
// and 20 minutes. This helps to ensure that a bunch of old player processes
// aren't left hanging around, running idle in the background.
val inactivityTimeoutMs = Random.nextInt(15 * 60000, 20 * 60000)

var expiry = System.currentTimeMillis() + inactivityTimeoutMs

fun delayExpiration(pointInTimeMs : Long) {
  val newExpiry = pointInTimeMs + inactivityTimeoutMs
  if (newExpiry > expiry) {
    expiry = newExpiry
  }
}

fun delayExpiration() {
  delayExpiration(System.currentTimeMillis())
}

class Run : CliktCommand(
  help = "Run the Alda player process"
) {
  val port by option(
    "--port", "-p", help = "the port to listen on"
  ).int().required()

  val lazyAudio by option(
    "--lazy-audio", help = "don't immediately set up audio device resources"
  ).flag(default = false)

  override fun run() {
    val log = logger!!

    log.info { "Starting receiver, listening on port $port..." }
    val receiver = receiver(port)
    receiver.startListening()

    if (!lazyAudio) {
      midi()
    }

    val player = player()
    log.info { "Starting player..." }
    player.start()

    Runtime.getRuntime().addShutdownHook(thread(start = false) {
      log.info { "Stopping receiver..." }
      receiver.stopListening()
      log.info { "Stopping player..." }
      player.interrupt()
    })

    while (isRunning) {
      try {
        if (System.currentTimeMillis() > expiry) {
          log.info { "Shutting down due to inactivity." }
          break
        }

        Thread.sleep(100)
      } catch (iex : InterruptedException) {
        log.info { "Interrupted." }
        break
      }
    }
  }
}

class Root : CliktCommand(
  name = "alda-player",
  help = "A background process that plays Alda scores"
) {
  val verbose by option(
    "--verbose", "-v", help = "verbose output"
  ).flag(default = false)

  override fun run() {
    if (verbose) {
      Configurator.setRootLevel(Level.DEBUG)
    }
  }
}

fun main(args: Array<String>) {
  System.setProperty("playerId", playerId)
  System.setProperty("logPath", projDirs.cacheDir)
  logger = KotlinLogging.logger {}

  Root()
    .subcommands(Info(), Run())
    .main(args)

  exitProcess(0)
}


package io.alda.player

import com.github.ajalt.clikt.core.CliktCommand
import com.github.ajalt.clikt.core.subcommands
import com.github.ajalt.clikt.parameters.options.default
import com.github.ajalt.clikt.parameters.options.flag
import com.github.ajalt.clikt.parameters.options.option
import com.github.ajalt.clikt.parameters.types.int
import io.github.soc.directories.ProjectDirectories
import java.net.ServerSocket
import java.nio.file.Paths
import kotlin.random.Random
import kotlin.concurrent.thread
import kotlin.streams.asSequence
import kotlin.system.exitProcess
import mu.KotlinLogging
import mu.KLogger
import org.apache.logging.log4j.core.config.Configurator
import org.apache.logging.log4j.Level

var logger : KLogger? = null
var stateManager : FileBasedStateManager? = null

var isRunning = true

// adapted from https://stackoverflow.com/a/46944275/2338327
private fun generateId() : String {
  val source = "abcdefghijklmnopqrstuvwxyz"
  return (1..3).map { _ -> source.get(Random.nextInt(0, source.length)) }
               .joinToString("")
}

val playerId = generateId()

val projDirs = ProjectDirectories.from("", "", "alda")
val logPath = Paths.get(projDirs.cacheDir, "logs").toString()

class Info : CliktCommand(
  help = "Print useful information including the version and log path"
) {
  override fun run() {
    println("alda-player $playerVersion")
    println("log path: $logPath")
  }
}

private fun findOpenPort() : Int {
  ServerSocket(0).use { socket -> return socket.getLocalPort() }
}

class Run : CliktCommand(
  help = "Run the Alda player process"
) {
  val port by option(
    "--port",
    "-p",
    help = "the port to listen on (default: random available port)"
  ).int()
   .default(findOpenPort())

  val lazyAudio by option(
    "--lazy-audio", help = "don't immediately set up audio device resources"
  ).flag(default = false)

  override fun run() {
    val log = logger!!

    stateManager = FileBasedStateManager(port)
    stateManager!!.start()

    log.info { "Starting receiver, listening on port $port..." }
    val receiver = receiver(port)
    receiver.startListening()

    if (lazyAudio) {
      log.info {
        "Deferring audio initialization; --lazy-audio flag was provided."
      }
    } else {
      engine()
    }

    val player = player()
    player.start()

    Runtime.getRuntime().addShutdownHook(thread(start = false) {
      stateManager!!.stop()
      log.info { "Stopping receiver..." }
      receiver.stopListening()
      log.info { "Stopping player..." }
      player.interrupt()
    })

    while (isRunning) {
      try {
        if (stateManager!!.isExpired()) {
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

  val veryVerbose by option(
    "--very-verbose", "-vv", help = "very verbose output"
  ).flag(default = false)

  override fun run() {
    when {
      veryVerbose -> Configurator.setRootLevel(Level.TRACE)
      verbose     -> Configurator.setRootLevel(Level.DEBUG)
    }
  }
}

fun main(args: Array<String>) {
  System.setProperty("playerId", playerId)
  System.setProperty("logPath", logPath)
  logger = KotlinLogging.logger {}

  Root()
    .subcommands(Info(), Run())
    .main(args)

  exitProcess(0)
}


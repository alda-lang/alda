package io.alda.player

import io.github.soc.directories.ProjectDirectories
import kotlin.random.Random
import kotlin.concurrent.thread
import kotlin.streams.asSequence
import kotlin.system.exitProcess
import mu.KotlinLogging
import org.apache.logging.log4j.core.config.Configurator
import org.apache.logging.log4j.Level

var isRunning = true

// adapted from https://stackoverflow.com/a/46944275/2338327
private fun generateId() : String {
  val source = "abcdefghijklmnopqrstuvwxyz"
  return (1..3).map { _ -> source.get(Random.nextInt(0, source.length)) }
               .joinToString("")
}

val playerId = generateId()

// FIXME: only -v or -V or PORT is supported, not multiple
// TODO: proper CLI argument/options parsing
fun main(args: Array<String>) {
  System.setProperty("playerId", playerId)

  val projDirs = ProjectDirectories.from("io", "alda", "alda")
  System.setProperty("logPath", projDirs.cacheDir)
  val log = KotlinLogging.logger {}

  if (args.isEmpty()) {
    println("Args: [-v|--verbose] [-V|--version] | PORT")
    exitProcess(1)
  }

  // if (args[0] == "-v" || args[0] == "--verbose") {
    Configurator.setRootLevel(Level.DEBUG)
  // }

  if (args[0] == "-V" || args[0] == "--version") {
    println("TODO: print version information")
    return
  }

  val port = args[0].toInt()
  log.info { "Starting receiver, listening on port $port..." }
  val receiver = receiver(port)
  receiver.startListening()

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
      Thread.sleep(100)
    } catch (iex : InterruptedException) {
      log.info { "Interrupted." }
      break
    }
  }

  exitProcess(0)
}


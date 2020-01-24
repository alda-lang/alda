package io.alda.player

import kotlin.concurrent.thread
import kotlin.system.exitProcess

var isRunning = true

fun main(args: Array<String>) {
  if (args.isEmpty()) {
    println("Args: PORT")
    exitProcess(1)
  }

  val port = args[0].toInt()
  println("Starting receiver, listening on port $port...")
  val receiver = receiver(port)
  receiver.startListening()

  val player = player()
  println("Starting player...")
  player.start()

  Runtime.getRuntime().addShutdownHook(thread(start = false) {
    println("Stopping receiver...")
    receiver.stopListening()
    println("Stopping player...")
    player.interrupt()
  })

  while (isRunning) {
    try {
      Thread.sleep(100)
    } catch (iex : InterruptedException) {
      println("Interrupted.")
      break
    }
  }

  exitProcess(0)
}


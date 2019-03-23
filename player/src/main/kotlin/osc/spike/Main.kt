package osc.spike

import java.util.Scanner
import kotlin.system.exitProcess

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

  // TODO: replace this with a proper shutdown mechanism
  val scanner = Scanner(System.`in`)
  println("Press ENTER when done.")
  scanner.nextLine()

  // shutdown actions
  println("Stopping receiver...")
  receiver.stopListening()
  println("Stopping player...")
  player.interrupt()
}


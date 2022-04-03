package io.alda.player

import kotlinx.browser.window
import mu.KotlinLoggingConfiguration
import mu.KotlinLoggingLevel

fun setLogLevel(level : String) {
  KotlinLoggingConfiguration.LOG_LEVEL = when (level) {
    "trace" -> KotlinLoggingLevel.TRACE
    "debug" -> KotlinLoggingLevel.DEBUG
    "info"  -> KotlinLoggingLevel.INFO
    "warn"  -> KotlinLoggingLevel.WARN
    "error" -> KotlinLoggingLevel.ERROR
    else    -> throw IllegalArgumentException("Invalid log level: ${level}")
  }
}

fun main() {
  console.info("alda-player.js version ${playerVersion} loaded")

  window.asDynamic().AldaPlayer = jsObject {
    setLogLevel = ::setLogLevel
  }
}

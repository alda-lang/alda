package io.alda.player

import kotlinx.browser.window
import kotlin.js.Json
import mu.KotlinLoggingConfiguration
import mu.KotlinLoggingLevel

// It's absurd that Kotlin.js doesn't provide a stdlib function for this!
// Source: https://stackoverflow.com/a/47380123/2338327
inline fun jsObject(init: dynamic.() -> Unit): Json {
    val o = js("{}")
    init(o)
    return o
}

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
}

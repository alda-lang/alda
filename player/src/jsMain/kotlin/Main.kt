package io.alda.player

import kotlinx.browser.window
import kotlin.js.Json

// It's absurd that Kotlin.js doesn't provide a stdlib function for this!
// Source: https://stackoverflow.com/a/47380123/2338327
inline fun jsObject(init: dynamic.() -> Unit): Json {
    val o = js("{}")
    init(o)
    return o
}

fun main() {
  console.info("alda-player.js version ${playerVersion} loaded")

  window.asDynamic().AldaPlayer = jsObject {
    aldaJsTestFunction = fun (n: Int) : String {
      return "aldaJsTestFunction output for n: ${n}"
    }
  }
}
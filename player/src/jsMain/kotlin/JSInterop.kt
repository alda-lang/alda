package io.alda.player

import kotlin.js.Json

// It's absurd that Kotlin.js doesn't provide a stdlib function for this!
// Source: https://stackoverflow.com/a/47380123/2338327
inline fun jsObject(init: dynamic.() -> Unit): Json {
    val o = js("{}")
    init(o)
    return o
}


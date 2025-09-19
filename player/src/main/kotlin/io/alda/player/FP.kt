package io.alda.player

fun <K, V> merge(m1: Map<K, V>, m2: Map<K, V>): Map<K, V> {
  return m1 + m2
}

fun <K, V> assoc(m: Map<K, V>, k: K, v: V): Map<K, V> {
  return merge(m, mapOf(k to v))
}

fun <K, V> update(m: Map<K, V>, k: K, f: (V?) -> V): Map<K, V> {
  return assoc(m, k, f(m[k]))
}

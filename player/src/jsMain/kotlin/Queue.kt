package io.alda.player

class Queue<T> {
  private val list = arrayListOf<T>()

  override fun toString(): String = list.toString()

  fun size() = list.size

  fun peek(): T? = list.getOrNull(0)

  fun add(element: T) = list.add(element)

  fun take(): T? = if (list.size == 0) null else list.removeAt(0)
}

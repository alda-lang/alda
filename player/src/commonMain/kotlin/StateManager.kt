package io.alda.player

interface StateManager {
  fun delayExpiration()
  fun markActive()
  fun markReady()
}

class NoOpStateManager : StateManager {
  override fun delayExpiration() {}
  override fun markActive() {}
  override fun markReady() {}
}

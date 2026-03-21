package io.alda.player

interface StateManager {
  fun delayExpiration()
  fun markActive()
  fun markReady()

  // TODO: remove
  fun sleep(ms : Long)
}

class NoOpStateManager : StateManager {
  override fun delayExpiration() {}
  override fun markActive() {}
  override fun markReady() {}
  override fun sleep(ms : Long) {}
}

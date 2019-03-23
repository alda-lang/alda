package io.alda.player

import com.illposed.osc.OSCBadDataEvent
import com.illposed.osc.OSCBundle
import com.illposed.osc.OSCMessage
import com.illposed.osc.OSCPacket
import com.illposed.osc.OSCPacketEvent
import com.illposed.osc.OSCPacketListener
import com.illposed.osc.transport.udp.OSCPortIn
import com.illposed.osc.transport.udp.OSCPortInBuilder

private fun instructions(packet : OSCPacket) : List<OSCMessage> {
  if (packet is OSCMessage) {
    return listOf(packet)
  }

  return (packet as OSCBundle).getPackets().flatMap { instructions(it) }
}

fun receiver(port : Int) : OSCPortIn {
  return OSCPortInBuilder().setPort(port)
                           .setPacketListener(object : OSCPacketListener {
    override fun handlePacket(event : OSCPacketEvent) {
      playerQueue.put(instructions(event.getPacket()))
    }

    override fun handleBadData(event : OSCBadDataEvent) {
      println("bad data: $event")
    }
  }).build()
}


# The Alda player

Alda is designed so that playback is asynchronous. When you use the Alda
[client](../client) to play a score, the client sends a bunch of
[OSC][osc-intro] messages to a player process running in the background.

[osc-intro]: http://opensoundcontrol.org/introduction-osc

The player process is agnostic of the Alda language. It simply receives and
handles OSC messages containing lower-level instructions pertaining to audio
playback.

The player supports live coding in that it allows one to define, modify, and
loop patterns during playback.

## Development

> Gradle is used to build/run the player, however there is a `gradlew` wrapper
> checked into the repo that makes it so that you don't need to have Gradle
> installed. You can replace `gradle` below with `./gradlew` and it should work
> the same as if you had Gradle installed.

To run the player, listening on port 27278 (or replace with the port number of
your choosing):

```bash
gradle run --args 27278
```

You can then use the [client](../client) to send messages on the same port.

## License

Copyright © 2019 Dave Yarwood, et al

Distributed under the Eclipse Public License version 2.0.

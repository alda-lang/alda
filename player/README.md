# The Alda player

Alda is designed so that playback is asynchronous. When you use the Alda
[client](../client) to play a score, the client sends a bunch of
[OSC][osc-page] messages to a player process running in the background.

[osc-page]: http://opensoundcontrol.org

The player process is agnostic of the Alda language. It simply receives and
handles OSC messages containing lower-level instructions pertaining to audio
playback.

The player supports live coding in that it allows you to define, modify, and
loop patterns during playback.

## Development

> Once you're all set up to run the Alda player process locally, have a look at
> the [development guide](./doc/development-guide.md) for a high-level overview
> of the code base.

### Requirements

* Java 8+
  * You probably already have Java 8 or higher installed.

    To check what version of Java you have, you can run `java -version`.

* Gradle (optional)
  * There is a `gradlew` wrapper script checked into the repo that makes it so
    that you don't need to have Gradle installed.

    In the commands below, you can replace `gradle` with `./gradlew` and it
    should work the same as if you had Gradle installed.

### Running a player process

After following the instructions below to start a player process on a particular
port, you can then use the [Alda client](../client) to send messages on the same
port.

There are two ways to run the player process locally:

* **Basic**: run a script and pass it arguments as if you're running
  `alda-player` on your PATH. Takes a little bit longer sometimes, but is more
  convenient most of the time and it's exactly like running a release
  executable.

* **Gradle**: use `gradle` (or `./gradlew`) for faster builds (no need to fully
  compile the executable) and more control over build options. Useful if you're
  handy with Gradle and you know what you're doing.

#### Basic

To run the player with no arguments, which displays usage information:

```bash
# equivalent to running `alda-player`
bin/run
```

To run the player, listening on port 27278 (or replace with the port number of
your choosing):

```bash
# equivalent to running `alda-player run -p 27278`
bin/run run -p 27278
```

#### Gradle

To run the player with no arguments, which displays usage information:

```bash
# equivalent to running `alda-player`
gradle run
```

To run the player, listening on port 27278 (or replace with the port number of
your choosing):

```bash
# equivalent to running `alda-player run -p 27278`
gradle run --args "run -p 27278"
```

### Compiling and testing alda-player.js

alda-player.js is the in-browser version of the Alda player process. It depends
on alda.wasm, which is compiled from the [Go client](../client) code.

To compile the source in `src/jsMain` to alda-player.js, run:

```bash
gradlew jsBrowserProductionWebpack
```

This creates the optimized/minified JavaScript source file
`build/distributions/alda-player.js`.

Then, you can open [`test-page.html`](test-page.html), a simple HTML page with a
script tag that loads the aforementioned JavaScript file, and play around in the
JavaScript console in your browser.

For added convenience, use Gradle's `-t` flag to watch the file system for
changes and recompile alda-player.js every time you make changes to the source:

```bash
gradlew -t jsBrowserProductionWebpack
```

> NOTE: Kotlin's JS tooling also provides a `jsBrowserRun` task, which is
> supposed to do hot code reloading, but I think that might only apply to React?
> It didn't seem to work with straightforward normal script like we're using,
> which I guess makes sense, because how would it know which part of the code to
> reload?

## License

Copyright © 2019-2022 Dave Yarwood, et al

Distributed under the Eclipse Public License version 2.0.

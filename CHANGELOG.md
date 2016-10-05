# CHANGELOG

## 1.0.0-rc43 (9/25/16)

* Tuned JVM performance to use less unnecessary memory and CPU. See issue [#269](https://github.com/alda-lang/alda/issues/269) for more context, but the TL;DR version is that prior to this release, the Alda server and worker processes were each provisioned with more memory and CPU settings than they needed, causing them to wastefully use up memory/CPU resources that your computer could otherwise be using.

  Big thanks to [feldoh] for this contribution!

## 1.0.0-rc42 (9/24/16)

* Implemented an internal "worker status" system so that the Alda client has better visibility into the status of the worker process handling a request to play a score. This only affects the `alda play` command.

  `alda play` requests that take longer than 3000 milliseconds will no longer time out and result in an incorrect error about the server being "down." The way it works now is more asynchronous:

  - The client makes a request to play a large score.
  - A worker gets the request and responds immediately. The server includes a note about which worker it is so that the client can follow up with the worker for status.
  - The client repeatedly sends requests for the status of that worker, and the worker asynchronously responds to let the client know if it is parsing, playing, done, or if there was some error.
  - As the client receives updates, it prints them to the console:

  ```bash
  $ alda play -c 'bassoon: (Thread/sleep 5000) o2 d1~1~1~1~1'
  [27713] Parsing/evaluating... # immediate response
  [27713] Playing... # 5 seconds later
  ```

### Breaking Changes

* To accomplish the above, I had to make a couple of minor adjustments to the ZeroMQ message structure. If you are writing your own Alda client, server, or worker, you may need to adjust the way messages are handled slightly. The Alda [ZeroMQ Architecture](doc/zeromq-architecture.md) doc has been updated to reflect these changes to the message structure.

## 1.0.0-rc41 (9/18/16)

The focus of this release is to use less CPU when starting an Alda server and worker processes. Thanks to [0atman] for reporting [this issue](https://github.com/alda-lang/alda/issues/266)!

* Prior to this release, when you started up and Alda server, it would simultaneously start all of its worker processes, which was heavy on the CPU. It turns out that waiting 10 seconds between starting each worker decreases CPU usage significantly, so that's what we're doing now.

* Because it can take a little longer now for all workers to start (add an additional 10 seconds for each worker after 1), the server will now check on the current number of workers every 60 seconds instead of every 30 seconds.

* The default number of workers is now 2 instead of 4. This uses significantly less CPU, and should be adequate for most Alda users. Remember that if you desire more workers, you can use the `--workers` option when starting the server, e.g. `alda --workers 4 up` for 4 workers.

* Removed the temporary `--cycle-workers` option introduced in 1.0.0-rc39. Now that CPU usage will not be as much of a problem, it is safer to cycle workers by default.

## 1.0.0-rc40 (9/17/16)

* Bugfix: prior to this release, if evaluating an Alda score resulted in any error, the client would report `ERROR Invalid Alda syntax.`. Now it will report the error message. (This may have been working at some point and then regressed in a recent release. Sorry about that!)

## 1.0.0-rc39 (9/17/16)

* There is [a CPU usage issue](https://github.com/alda-lang/alda/issues/266) that can cause poor performance when cycling out workers, a feature that was implemented in the last release.

  The real solution is going to be to fix the CPU usage problem, but in the meantime, this release makes the worker-cycling behavior introduced in the last release an opt-in feature.

  To opt in, include the `--cycle-workers` option when starting the server:

  ```
  alda --cycle-workers up
  ```

  Note that even when _not_ using this option, the worker cycling will still eventually happen (within 30 seconds of your computer coming out of suspended state). When it has been 10+ seconds since the worker was last active (e.g. if your computer was in suspended state), each worker will detect this and shut itself down. Every 30 seconds, the server checks to make sure it has the correct number of workers, so it will replace the workers that went down.

  The change with this release is that the workers will not immediately be replaced upon coming out of suspended state. It would be better if they were immediately replaced, but we will have to fix the CPU usage problem first. (If you're good at profiling / determining the cause of high CPU usage, we could use your help!)

## 1.0.0-rc38 (9/14/16)

* Fixed issue [#160](https://github.com/alda-lang/alda/issues/160), re: MIDI audio being delayed after suspending and bringing back Alda processes, e.g. after closing and re-opening your laptop lid.

  Now that we have a server/workers architecture, the solution to this is for the server and workers to each detect when the system has been suspended, and act accordingly:

  - Each existing worker will shut down.
  - The server will clear out its worker queue and start up new workers.

  This is not a perfect solution in that if you close your laptop and then reopen it in under 10 seconds, the suspension may not be properly detected, so you might still have the same "bad" workers as before with delayed audio. If you ever observe this behavior, you can always fix it by restarting the server and workers via `alda downup`.

* Fixed a minor bug where workers may be sending too many heartbeats, potentially resulting in poor performance. Now they should only send heartbeats once per second.

## 1.0.0-rc37 (9/8/16)

* Continuing from the `ulimit -n`-related `alda up` issues noted in the previous release, this makes it so that we check the number of available workers every 250 ms instead of 100 ms. This does not solve the "Too many open files" issue, but as a workaround it appears to stop the error from happening when your `ulimit -n` is 256 and you're starting the default number (4) of workers.

  See issue [#261](https://github.com/alda-lang/alda/issues/261) for further discussion.

## 1.0.0-rc36 (9/5/16)

This release makes a handful of minor improvements to the way we're managing socket connections. The goal is avoid the `java.io.IOException: Too many open files` error, (issue [#261](https://github.com/alda-lang/alda/issues/261)) which can occur when Alda tries to open more socket connections than your system allows.

* Updated to JeroMQ 0.3.5, which includes some behind-the-scenes resource management improvements.

* Ensured that any time Alda forks a background process, it closes its input, output, and error streams in order to free up those resources that we aren't using.

* When you run `alda up`, the client repeatedly sends requests to get an update on when the server is up so that it can notify you, then it does the same thing with the worker processes. The worker processes can take a while to spin up, so the client ends up sending a lot of requests to the server for the status of the worker processes. Prior to this release, each of these requests created a new connection, causing the "open files" limit to be quickly met. As of this release, a single Alda CLI command (like `alda up`) will only create a single connection and reuse it.

I tested this release by setting my `ulimit -n` to 256 and saw some definite improvement, but still occasionally saw the "Too many open files" error happening, especially when trying to use 4 Alda worker processes. I'll continue to tinker with this and see if there are other improvements to be made.

> If you have a low `ulimit -n` (256 is quite low; I would recommend 10240 at least) and would like to get better performance from Alda, you can set it higher for your current terminal session by running `ulimit -n 10240`. However, this setting will go away once you close your terminal window, so if you're looking for a more permanent solution, [the Riak docs](https://docs.basho.com/riak/kv/2.1.4/using/performance/open-files-limit/) happen to have an excellent description of how to set your open file limit more permanently.

## 1.0.0-rc35 (9/4/16)

> Quick TL;DR of what you should do when updating to this release:
>
> - _Before_ updating, run `alda down` to stop your Alda server if you have one running. If you don't do this before updating, you may need to find and kill the process manually.
> - Run `alda update` to update.
> - `alda start`, `alda stop` and `alda restart` have been renamed. From now on, you will need to use the commands `alda up`, `alda down` and `alda downup` or `alda start-server`, `alda stop-server` and `alda restart-server` instead. See below for more details about this change.

* This release adds more internal architecture improvements, further leveraging ZeroMQ to break out the functionality of Alda into separate, specialized background processes.

  Prior to this release, Alda was a two-process system:

  - a **server** that runs in the background and plays scores on demand, and
  - a command-line **client** used to make requests to the server

  To help with performance issues related to playing multiple scores at once, each request is now handled by a separate **worker** process. The **client**/**server** relationship is still exactly the same; as an Alda user, you never need to communicate directly with the workers or worry about them in any way; they are solely the responsibility of the server.

  When starting Alda with `alda up`, you will now see the server come up much faster than before, followed by a more typical wait for the first worker to become available:

  ```bash
  $ alda up
  [27713] Starting Alda server...
  # short wait
  [27713] Server up ✓
  [27713] Starting worker processes...
  # longer wait
  [27713] Ready ✓
  ```

  When you see the "Ready" line, you will be able to use Alda the same as before, e.g.:

  ```bash
  $ alda play -c 'bassoon: o2 d8 e (quant 33) f+ g (quant 90) a2'
  [27713] Playing...
  ```

  You can safely leave the server running in the background and it will manage its own supply of workers, replacing any that become unresponsive. Similarly, a worker process will shut itself down if the server becomes unresponsive for any reason.

  By default, an Alda server maintains a pool of 4 workers. If you try to play more than 4 scores at the same time, the server will respond that there are no workers available and ask you to wait until the next worker becomes available:

  ```bash
  $ alda play -c 'bassoon: o2 d1~1~1~1~1'
  [27713] Playing...

  $ alda play -c 'bassoon: o2 d1~1~1~1~1'
  [27713] Playing...

  $ alda play -c 'bassoon: o2 d1~1~1~1~1'
  [27713] Playing...

  $ alda play -c 'bassoon: o2 d1~1~1~1~1'
  [27713] Playing...

  $ alda play -c 'bassoon: o2 d1~1~1~1~1'
  [27713] ERROR All worker processes are currently busy. Please wait until playback is complete and try again.
  ```

  If you desire more or fewer workers, you can use the `-w`/`--workers` option when starting the server:

  ```bash
  $ alda -w 6 up
  [27713] Starting Alda server...
  [27713] Server up ✓
  [27713] Starting worker processes...
  [27713] Ready ✓
  ```

  Be aware, however, that running a larger amount of servers uses more of your CPU. I think the default of 4 workers should be comfortable for most users.

* Running `alda status` will now show you the number of workers currently available:

  ```bash
  $ alda status
  [27713] Server up (4/4 workers available)

  $ alda play -f examples/hello_world.alda
  [27713] Playing...

  # one worker is busy playing hello_world.alda
  $ alda status
  [27713] Server up (3/4 workers available)

  # after playback completes, the worker is available again
  $ alda status
  [27713] Server up (4/4 workers available)
  ```

* Running `alda list` will now list all Alda processes, including servers and workers.

  > This command is currently only available on Unix/Linux systems.
  >
  > Please let us know if you can help us implement the same functionality for Windows!

  ```bash
  $ alda list
  [27713] Server up (4/4 workers available)
  [58678] Worker (pid: 85804)
  [58678] Worker (pid: 85805)
  [58678] Worker (pid: 85806)
  [58678] Worker (pid: 85807)
  ```

* If an Alda worker process ever encounters an error, it logs the error to `~/.alda/logs/error.log`. The logs in this folder are rotated weekly, making cleanup easy if you do not want to keep old error logs.

* Fixed [issue #243](https://github.com/alda-lang/alda/issues/243). Prior to the infrastructure improvements made in this release, the server was a single process trying to play multiple scores at once, which caused the sound to break up. Having each worker be a separate process that only handles one score at a time results in more reliable playback. Thanks to [goog] for reporting this issue!

* Fixed [issue #258](https://github.com/alda-lang/alda/issues/258). This issue was also related to the previous single-server/worker system. While busy handling one request, the server was not able to handle the next right away, causing the request to be re-submitted and handled more than once, resulting in scores playing more than once with unexpected timing. Now that the server's sole responsibility is to maintain worker processes and forward them requests, it can handle requests a lot faster, and the client no longer needs to submit them more than once. Thanks to [elyisgreat] for reporting this issue!

### Breaking Changes

* In order to enable the new server/workers architecture, we had to remove a handful of commands that relied on "stateful server" behavior. An Alda server no longer has a notion of the "current score" you are working with. Instead, the state of your score should be maintained in the form of Alda code in a file; the functionality of the removed commands is better left up to your file system and text editor.

  The following commands have been removed:

  - `alda play` (without a `--file` or `--code` argument, this command used to play the "current score" in memory on the server)
  - `alda play --append` (this used to play a file or string of code and append it to the "current score").
  - `alda append` / (same as `alda play --append`, but just appended to the "current score" without playing the new code)
  - `alda score` (used to display the "current score")
  - `alda info` (used to display information about the "current score")
  - `alda new` (used to delete the "current score" and start a new one)
  - `alda load` (used to replace the "current score" with the contents of a file or string of code)
  - `alda save` (used to save the "current score" to a file)
  - `alda edit` (used to open the "current score" in your text editor of choice)

* In anticipation of a future `alda stop` command that will [stop playback](https://github.com/alda-lang/alda/issues/69) instead of stopping the server, the previous `alda stop` command (which would stop the server) has been renamed `alda stop-server`. By analogy, `alda start` has been renamed to `alda start-server` and `alda restart` has been renamed to `alda restart-server`. I recommend using the shorter aliases `alda up`, `alda down` and `alda downup` to start, stop, and restart the server.

## 1.0.0-rc34 (8/23/16)

* This release adds a few safeguards against inadvertently starting more than one server process for the same port and ending up in a situation where you have potentially many Alda server processes hanging around, only one of which will be able to serve on that port. Thanks to [elyisgreat] for reporting this issue ([#258](https://github.com/alda-lang/alda/issues/258)).

  Safeguards include:

  * Increasing the "server startup timeout" back to 30 seconds. It turns out that I accidentally lowered it to 15 seconds in 1.0.0-rc32, and on some computers an Alda server can take longer than 15 seconds to start. 30 seconds seems like a better default.

    This timeout is the number of seconds, after running `alda up`, before the client gives up on waiting for the server to start, assumes something went wrong, and tells you that the server is down.

  * If your computer is particularly slow and it takes longer than 30 seconds to start an Alda server, you can increase the timeout by supplying a new global option `-t` / `--timeout`:

    ```bash
    alda --timeout 45 up # wait 45 seconds before giving up
    ```

    This should only be necessary on the slowest of computers, but the option is there if you need it.

  * If you do experience a timeout and you try to start the server again, a helpful message is now displayed, letting you know that there is already a server trying to start on that port:

    ```
    [27713] There is already a server trying to start on this port. Please be patient -- this can take a while.
    ```

    ...and the client does not attempt to start a duplicate server for that port.

    If you wait long enough, the existing server should be up and ready to play scores. You can check the status of the server by running `alda status`.

## 1.0.0-rc33 (8/20/16)

* Made the server a little more resilient to failure. There are a handful of tasks that are handled in parallel, like setting up the audio systems (e.g. MIDI) for a score. This involves running the tasks in parallel background threads and waiting for them to complete. This works fine when the tasks are successful, but if they fail, then the server ends up waiting forever and needs to be restarted in order to serve any more requests.

  Starting from this release, we are able to notice when the background tasks fail and re-throw the error so that the server can report the error back to the client and will be ready to handle any subsequent requests.

  One example of a background task that can fail is if you try to play an Alda score with multiple MIDI percussion instruments. There is only one MIDI percussion channel available, so this will throw a "Ran out of MIDI channels! :(" error. on the background thread that loads the instruments. Before this release, the server would just lock up when you tried to play such a score; now it will report the error back to the client.

* Added more debug logging when running a server in debug mode.

## 1.0.0-rc32 (8/18/16)

* The major change in this release is that we replaced the internal implementation of the server, previously a resource-intensive HTTP server, with a more lightweight [ZeroMQ](http://zeromq.org) REQ/RES socket. This means lower overhead for the server, which should translate to better performance.

  > We are using [the pure Java implementation of ZeroMQ](https://github.com/zeromq/jeromq), which means no native dependencies to install. You can update Alda and it'll just work™.

  This is a transparent change; you should not notice any difference in your usage of Alda via the `alda` command-line client, besides seeing better performance from the server.

* Bugfix: I just realized that the `alda play` command's options `--from` and `--to` were not actually hooked up so that they did anything useful. Oops! Sorry about that. As of this release, you can use them to specify start and end points for playing back the score, either as marker names or minute/second markings. For real, this time.

* I discovered that the `-y/--yes` auto-confirm option for the `alda play` and `alda parse` commands was never being used, so I removed the option from both commands.

  When I first wrote the Alda client, the `play` and `parse` commands would overwrite whatever score you were currently working on in memory. So, if you had unsaved changes, the client would warn you that you were about to overwrite the in-memory score and lose your unsaved changes, and ask you for confirmation. The `-y` flag was there as a convenience so that if you didn't care about the score in memory and you just wanted to play a score file, you could include the flag and skip the unsaved changes warning and prompt.

  A while back, we improved the experience so that the default behavior of `alda play` and `alda parse` is to play or parse a score separately without affecting the working score you might have in-memory. The client never needs to prompt you for confirmation anymore when using these two commands, so the `-y` flag is no longer necessary.

  > Note: The `-y` flag is still available for other commands that _can_ affect the score in memory, such as `alda load` and `alda down`.

### Breaking Changes

* If you happened to be using the HTTP server directly (i.e. via `curl` instead of the Alda client), you will find that this will no longer work, since the server does not respond to HTTP requests anymore.

  > If you still wish to have lower-level access to an Alda server, you can use ZeroMQ to make a REQ socket connection to `tcp://*:27713` (assuming 27713 is the port on which the Alda server is running), send a request with a JSON payload, and receive the response. For more information about ZeroMQ, read [the ZeroMQ guide](http://zguide.zeromq.org/page:all), it's excellent!

* Before, if you ran a command that requires the Alda server to be up, for example `alda play -f somefile.alda`, and the Alda server is not up, the client would go ahead and start it for you. We have removed this behavior; now, you will get an error message notifying you that the server is down and explaining that you can start the server by running `alda up`. We are working toward making the server failsafe enough that you can leave it running in the background forever and forget about it (making it more of a daemon), so hopefully you will not need to run `alda down` or `alda up` very often.

* When running `alda score -m` (to show the data representation of the current score) or `alda parse -m -f somefile.alda` (to show the data representation of the score in a file), the output is now JSON instead of [EDN](https://github.com/edn-format/edn).

  The reason for this is that JSON is more widely supported and there are a variety of useful command-line tools like [`jq`](https://stedolan.github.io/jq/) for working with JSON on the command line.

  For example, if you have `jq` installed, you can now run this command to pretty-print an Alda score map:

  ```
  $ alda score -m | jq .
  {
    "chord-mode": false,
    "current-instruments": [
      "flute-nUVei"
    ],
    "score-text": "(tempo! 90)\n(quant! 95)\n\npiano:\n  o5 g- > g- g-/f > e- d-4. < b-8 d-2 | c-4 e- d- d- <b-1/>g-\n\nflute:\n  r2 g-4 a- b-2. > d-32~ e-16.~8 < b-2 a- g-1\n",
    "events": [
      {
  ...

  ```

  Or you can output the offsets of every event in the score:

  ```
  $ alda score -m | jq '.events[] .offset'
  666.6666870117188
  2000.0000610351562
  5333.333465576172
  1333.3333740234375
  1333.3333740234375
  4666.666748046875
  4750.00008392334
  ```

## 1.0.0-rc31 (8/12/16)

* Fixed a bug where the Alda server was spinning its wheels for a long time trying to parse certain scores when played via the command-line.

## 1.0.0-rc30 (8/11/16)

* Removed the `--pre-buffer` and `--post-buffer` options, as I realized they weren't necessary. For details, see [this issue comment](https://github.com/alda-lang/alda/issues/26#issuecomment-239345440).

* Improved the timing of the behind-the-scenes clean-up that occurs after a score is finished playing.

* Fixed a bug where in some cases, that clean-up might not have been happening.

* Implemented a minor optimization to the way events are scheduled; earlier events are now given priority scheduling, making it less likely for events to be skipped during playback.

## 1.0.0-rc29 (7/26/16)

* Fixed a bug where using chords with implicit note duration within a CRAM expression would result in incorrect timing.

## 1.0.0-rc28 (7/25/16)

More variable-related bugfixes in this release:

* Setting a variable that includes a repeated event sequence, e.g. `foo = [c c]*3` no longer throws an error.

* Setting a variable that includes an empty Clojure expression, e.g. `foo = () c d e f` no longer throws an error.

Shout-out to [elyisgreat] for finding all these bugs!

## 1.0.0-rc27 (7/24/16)

* Behind the scenes change: simplified the code that handles getting and setting variables. See ebf2a42c78e5be3ef1cbedc4c2579a7bd72d08bb for more details.

  TL;DR: when you use a variable inside the definition of another variable, now the value of the variable is retrieved right away, instead of waiting until you try to get the value of the outer variable.

  For the most part, you should not notice any changes to the way that variables work, aside from the Alda score map (i.e. the result of using `alda parse -m`, or `:map` in the Alda REPL.) no longer containing the key `:variables` and having a greatly simplified `:env` key that is no longer a nested map, but a simple lookup of variables to their values.

* Trying to define a variable where the value includes an undefined variable now throws an error immediately, instead of waiting until you get the value of the variable. For example, the following score will now fail immediately because `bar` isn't defined:

  ```
  foo = bar
  ```

  Whereas before this release, the "undefined variable" error would not be thrown until you tried to use `foo` in an instrument part.

## 1.0.0-rc26 (7/24/16)

* Minor bugfix: in some situations, undefined variables were not being appropriately caught. Now, an error will always be thrown if you try to use a variable that hasn't been defined.

* Empty event sequences, e.g. `[]` are now supported. Hey, why not? ㄟ( ･ө･ )ㄏ

* Defining a previously undefined variable as itself, e.g. `foo = foo` used to trigger a stack overflow. Now it throws an "undefined variable: foo" error, which is more helpful.

## 1.0.0-rc25 (7/23/16)

### Breaking Changes

* There was a breaking change in one of the last few versions that may not have been documented, where it is no longer acceptable to place barlines in between notes instead of whitespace, like this:

  ```
  c|d|e
  ```

  To improve parsing flexibility, make the parser easier to develop, avoid unexpected bugs in future releases, and enforce a convention that will make Alda scores easier to read, whitespace is now required between all events, including not only notes, chords, event sequences, etc., but also barlines.

* As of this release, it is no longer valid to end a note duration with a barline, e.g.:

  ```
  c1| d
  ```

  The correct way to write the above is:

  ```
  c1 | d
  ```

  The reasoning for this is the same as the point above -- events (including barlines) must be separated by whitespace.

* Note that there is one situation that is a minor exception where it is acceptable for a barline to not be preceded/succeeded by whitespace:

  ```
  c1~|1~|1 d
  ```

  In this case, there are technically two "events" -- the C note and the D note. Alda's parser reads the C note as a single event where the pitch is C and the duration is 3 whole notes tied across 2 bar lines.

  If this is confusing, note that it is also acceptable to put spaces in between the barlines:

  ```
  c1 | ~1 | ~1 d
  ```

  The last two examples are equivalent.

## 1.0.0-rc24 (7/21/16)

* Refines the behavior of variables when used within the definitions of other variables. The "scope" of a variable is now tracked when it is defined. This makes it possible to do things like this:

  ```
  foo = c d e
  foo = foo f g

  piano: foo  # expands to "c d e f g"
  ```

  It also makes it so that you won't run into unexpected bugs when redefining a variable that another variable depends on, as in this example:

  ```
  foo = c d e
  bar = foo f g

  foo = c

  piano: bar  # still "c d e f g," because that's what it was when it was defined
  ```

### Breaking Changes

* Prior to this release, trying to use a variable that wasn't defined would result in that variable being ignored and the rest of the score processed as usual.

  Now, trying to use a variable that hasn't been defined results in an error being thrown. This will help score writers catch unintentional bugs caused by misspelling variable names, etc.

## 1.0.0-rc23 (7/18/16)

* Fixes another bug related to `>` and `<` being used back-to-back without spaces in between.

* An Alda score containing only Clojure code (i.e. no instrument parts) is now considered a valid score. For example, the following is a valid Alda score:

  ```
  (part "bassoon"
    (for [x (map (comp keyword str) "cdefgab")]
      (note (pitch x) (ms 100))))
  ```

## 1.0.0-rc22 (7/18/16)

* The previous release inadvertently made it invalid for a note to be followed immediately (no spaces) by an octave up/down operator, e.g. `c<`. This release makes it acceptable again to do that.

  (It's also still acceptable to follow an octave up/down operator immediately with a note, e.g. `>c`, and to sandwich a note between octave up/down operators, e.g. `>c<`.)

### Breaking Changes

* Because the `=` sign is used to define variables, the natural sign, which used to be `=`, has been changed to `_` to avoid confusion. If you have any scores using naturals, make sure you change the `=`'s to `_`'s to avoid parse errors.

## 1.0.0-rc21 (7/18/16)

* Variables implemented! This simple, but powerful feature allows you to define named sequences of musical events and refer to them by name. You can even define variables that refer to other variables, giving you the means to build a score out of modular parts. For details, see [the docs](doc/variables.md).

* Minor parsing performance improvements.

### Breaking Changes

* In order to avoid conflicts between variable names and multiple notes squished together (e.g. "abcdef"), the rules are now more rigid about spaces between notes. Multiple letters back-to-back without spaces in between is now read as a variable name. Multiple letters with spaces between each one (e.g. "a b c d e f") is read as multiple notes.

  Note that this does not only apply to single letters, but to anything else that the parser considers a "note" -- this includes notes that include a duration (e.g. "a4"), notes that include multiple durations tied together (e.g. "a4~4~4"), and notes that end in a final tie/slur, indicating that the note is to be played legato (e.g. "a~").

  A couple of Alda example scores contained examples that broke the "mandatory space between notes" rule, and had to be changed. For example, `awobmolg.alda` contained the following snippet representing 5 notes under a slur:

  ```
  b4.~b16~a~g~a
  ```

  The parser was trying to read this as the (legato/slurred) note `b4.~` followed immediately by another note, `b16`. This is now explicitly not allowed. The example was changed to the following (valid) syntax:

  ```
  b4.~ b16~ a~ g~ a
  ```

## 1.0.0-rc20 (6/20/16)

* Fixed a regression caused by 1.0.0-rc19, which was causing scores not to parse correctly.

## 1.0.0-rc19 (6/19/16)

* Parsing/playing Alda scores is now significantly faster, thanks to some optimizations to the parser. (Many thanks to [aengelberg] for your help with this!)

* Fixed [#235](https://github.com/alda-lang/alda/issues/235) -- when trying to parse (as a `--map`) or play a very large score, a "Method code too large!" error was occurring because of the way that scores were parsed into Clojure code as an intermediate form and then `eval`'d. Now, the parser transforms an Alda score into the score map (i.e. the output of `alda parse --map`) directly.

  Even though parsing and playing scores no longer does so by generating Clojure code, it is still possible to generate the Clojure code, if desired, by using `alda parse --lisp`.

  This should be a transparent change; both ways of parsing should still work the same as before.

### Breaking Changes

* Part of the process of optimizing the Alda parser was removing cases of ambiguity. A consequence of doing this is that the `duration` grammar rule no longer includes a `barline` or `slur` at the end. Instead, a `barline` must stand on its own (after the `note` containing the `duration`), and a `slur` must be part of a `note` instead of its `duration`.

  In other words, when writing alda.lisp code, whereas it used to be possible to do things like this:

  ```
  (note (pitch :c)
        (duration (note-length 4)
                  (barline)))

  (note (pitch :c)
        (duration (note-length 4)
                  :slur))
  ```

  Now you can only do it like this:

  ```
  (note (pitch :c)
        (duration (note-length 4)))
  (barline)

  (note (pitch :c)
        (duration (note-length 4))
        :slur)
  ```

  This is a trivial change, but I thought I'd mention it just in case anyone runs into it.

## 1.0.0-rc18 (5/28/16)

* Fixes a bug related to the fix introduced in 1.0.0-rc17. For more details, see [issue #231](https://github.com/alda-lang/alda/issues/231).

## 1.0.0-rc17 (5/21/16)

* Fixed issue #27. Setting note quantization to 100 or higher no longer causes issues with notes stopping other notes that have the same MIDI note number.

  Better-sounding slurs can now be achieved by setting quant to 100:

      bassoon: o2 (quant 100) a8 b > c+2.

## 1.0.0-rc16 (5/18/16)

* Fixed issue #228. There was a bug where repeated calls to the same voice were being treated as if they were separate voices. Hat tip to [elyisgreat] for catching this!

## 1.0.0-rc15 (5/15/16)

* This release includes numerous improvements to the Alda codebase. The primary goal was to make the code easier to understand and more predictable, which will make it possible to improve Alda and add new features at a much faster pace.

  To summarize the changes in programmer-speak: before this release, Alda evaluated a score by storing state in top-level, mutable vars, updating their values as it worked its way through the score. This code has been rewritten from the ground up to adhere much more to the functional programming philosophy. For a better explanation, read below about the breaking changes to the way scores are managed in a Clojure REPL.

* Alda score evaluation is now a self-contained process, and an Alda server (or a Clojure program using Alda as a library) can now handle multiple scores at a time without them affecting each other.

* Fixed issue #170. There was a 5-second socket timeout, causing the client to return "ERROR Read timed out" if the server took longer than 5 seconds to parse/evaluate the score. In this release, we've removed the timeout, so the client will wait until the server has parsed/evaluated the score and started playing it.

* Fixed issue #199. Local (per-instrument) attributes occurring at the same time as global attributes will now override the global attribute for the instrument(s) to which they apply.

* Using `@markerName` before `%markerName` is placed in a score now results in a explicit error, instead of throwing a different error that was difficult to understand. It turns out that this never worked to begin with. I do think it would be nice if it were possible to "forward declare" markers like this, but for the time being, I will leave this as something that (still) doesn't work, but that we could make possible in the future if there is demand for it.

### Breaking Changes

* The default behavior of `alda play -f score.alda` / `alda play -c 'piano: c d e'` is no longer to append to the current score in memory. Now, running these commands will play the Alda score file or string of code as a one-off score, not considering or affecting the current score in memory in any way. The previous behavior of appending to the current score is still possible via a new `alda play` option, `-a/--append`.

* Creating scores in a Clojure REPL now involves working with immutable data structures instead of mutating top-level dynamic vars. Whereas before, Alda event functions like `score`, `part` and `note` relied on side effects to modify the state of your score environment, now you create a new score via `score` (or the slightly lower-level `new-score`) and update it using the `continue` function. To better illustrate this, this is how you used to do it **before**:

  ```
  (score*)
  (part* "piano")
  (note (pitch :c))
  (chord (note (pitch :e)) (note (pitch :g)))
  ```

  Evaluating each S-expression would modify the top-level score environment. Evaluating `(score*)` again (or a full score wrapped in `(score ...)`) would blow away whatever score-in-progress you may have been working on.

  Here are a few different ways you can do this **now**:

  ```clojure
  ; a complete score, as a single S-expression
  (def my-score
    (score
      (part "piano"
        (note (pitch :c))
        (chord
          (note (pitch :e))
          (note (pitch :g))))))

  ; start a new score and continue it
  ; note that the original (empty) score is not modified
  (def my-score (new-score))

  (def my-score-cont
    (continue my-score
      (part "piano"
        (note (pitch :c)))))

  (def my-score-cont-cont
    (continue my-score-cont
      (chord
        (note (pitch :e))
        (note (pitch :g)))))

  ; store your score in an atom and update it atomically
  (def my-score (atom (score)))

  (continue! my-score
    (part "piano"
      (note (pitch :c))))

  (continue! my-score
    (chord
      (note (pitch :e))
      (note (pitch :g))))
  ```

  Because no shared state is being stored in top-level vars, multiple scores can now exist side-by-side in a single Alda process or Clojure REPL.

* Top-level score evaluation context vars like `*instruments*` and `*events*` no longer exist. If you were previously relying on inspecting that data, everything has now moved into keys like `:instruments` and `:events` on each separate score map.

* `(duration <number>)` no longer works as a way of manually setting the duration. To do this, use `(set-duration <number>)`, where `<number>` is a number of beats.

* The `$` syntax in alda.lisp (e.g. `($volume)`) for getting the current value of an attribute for the current instrument is no longer supported due to the way the code has been rewritten. We could probably find a way to add this feature back if there is a demand for it, but its use case is probably pretty obscure.

* Because Alda event functions no longer work via side effects, inline Clojure code works a bit differently. Basically, you'll just write code that returns one or more Alda events, instead of code that produces side effects (modifying the score) and returns nil. See [entropy.alda](examples/entropy.alda) for an example of the way inline Clojure code works starting with this release.

## 1.0.0-rc14 (4/1/16)

* Command-specific help text is now available when using the Alda command-line client. ([jgerman])

  To see a description of a command and its options, run the command with the `-h` or `--help` option.

  Example:

      $ alda play --help

      Evaluate and play Alda code
      Usage: play [options]
        Options:
          -c, --code
             Supply Alda code as a string
          -f, --file
             Read Alda code from a file
          -F, --from
             A time marking or marker from which to start playback
          -r, --replace
             Replace the existing score with new code
             Default: false
          -T, --to
             A time marking or marker at which to end playback
          -y, --yes
             Auto-respond 'y' to confirm e.g. score replacement
             Default: false

## 1.0.0-rc13 (3/10/16)

* Setting quantization to 0 now makes notes silent as expected. (#205, thanks to [elyisgreat] for reporting)

## 1.0.0-rc12 (3/10/16)

* Improve validation of attribute values to avoid buggy behavior when using invalid values like negative tempos, non-integer octaves, etc. (#195, thanks to [elyisgreat] for reporting and [jgkamat] for fixing)

## 1.0.0-rc11 (3/8/16)

* Fix parsing bugs related to ending a voice in a voice group with a certain type of event (e.g. Clojure expressions, barlines) followed by whitespace. (#196, #197 - thanks to [elyisgreat] for reporting!)

## 1.0.0-rc10 (2/28/16)

* Fix parsing bug re: placing an octave change before the slash in a chord instead of after it, e.g. `b>/d/f` (#192 - thanks to [elyisgreat] for reporting!)

## 1.0.0-rc9 (2/21/16)

* Fix parsing bug re: starting an event sequence with an event sequence. (#187 - Thanks to [heikkil] for reporting!)
* Fix similar parsing bug re: starting a cram expression with a cram expression.

## 1.0.0-rc8 (2/16/16)

* You can now update to the latest version of Alda from the command line by running `alda update`. ([jgkamat])

* This will be the last update you have to install manually :)

## 1.0.0-rc7 (2/12/16)

* Fixed a bug that was happening when using a cram expression inside of a voice. (#184 -- thanks to [jgkamat] for reporting!)

## 1.0.0-rc6 (1/27/16)

* Fixed a bug where voices were not being parsed correctly in some cases ([#177](https://github.com/alda-lang/alda/pull/177)).

## 1.0.0-rc5 (1/24/16)

* Added `midi-percussion` instrument. See [the docs](doc/list-of-instruments.md#percussion) for more info.

## 1.0.0-rc4 (1/21/16)

* Upgraded to the newly released Clojure 1.8.0 and adjusted the way we compile Alda so that we can utilize the new Clojure 1.8.0 feature [direct linking](https://github.com/clojure/clojure/blob/master/changes.md#11-direct-linking). This improves both performance and startup speed significantly.

## 1.0.0-rc3 (1/13/16)

* Support added for running Alda on systems with Java 7, whereas before it was Java 8 only.

## 1.0.0-rc2 (1/2/16)

* Alda now uses [JSyn](http://www.softsynth.com/jsyn) for higher precision of note-scheduling by doing it in realtime. This solves a handful of issues, such as [#134][issue133], [#144][issue144], and [#160][issue160]. Performance is probably noticeably better now.
* Running `alda new` now asks you for confirmation if there are unsaved changes to the score you're about to delete in order to start one.
* A heaping handful of new Alda client commands:
  * `alda info` prints useful information about a running Alda server
  * `alda list` (currently Mac/Linux only) lists Alda servers currently running on your system
  * `alda load` loads a score from a file (without playing it)
    * prompts you for confirmation if there are unsaved changes to the current score
  * `alda save` saves the current score to a file
    * prompts you for confirmation if you're saving a new score to an existing file
    * `alda new` will call this function implicitly if you give it a filename, e.g. `alda new -f my-new-score.alda`
  * `alda edit` opens the current score file in your `$EDITOR`
    * `alda edit -e <editor-command-here>` opens the score in a different editor

[issue133]: https://github.com/alda-lang/alda/issues/134
[issue144]: https://github.com/alda-lang/alda/issues/144
[issue160]: https://github.com/alda-lang/alda/issues/160


## 1.0.0-rc1 (12/25/15) :christmas_tree:

* Server/client relationship allows you to run Alda servers in the background and interact with them via a much more lightweight CLI, implemented in Java. Everything is packaged into a single uberjar containing both the server and the client. The client is able to manage/start/stop servers as well as interact with them by handing them Alda code to play, etc.
* This solves start-up time issues, making your Alda CLI experience much more lightweight and responsive. It still takes a while to start up an Alda server, but now you only have to do it once, and then you can leave the server running in the background, where it will be ready to parse/play code whenever you want, at a moment's notice.
* Re-implementing the Alda REPL on the client side is a TODO item. In the meantime, you can still access the existing Alda REPL by typing `alda repl`. This is just as slow to start as it was before, as it still has to start the Clojure run-time, load the MIDI system and initialize a score when you start the REPL. In the near future, however, the Alda REPL will be much more lightweight, as it will be re-implemented in Java, and instead of starting an Alda server every time you use it, you'll be interacting with Alda servers you already have running.
* Starting with this release, we'll be releasing Unix and Windows executables on GitHub. These are standalone programs; all you need to run them is Java. [Boot](http://boot-clj.com) is no longer a dependency to run Alda, just something we use to build it and create releases. For development builds, running `boot build -o directory_name` will generate `alda.jar`, `alda`, and `alda.exe` files which can be run directly.
* In light of the above, the `bin/alda` Boot script that we were previously using as an entrypoint to the application is no longer needed, and has been removed.
* Now that we are packaging everything together and not using Boot as a dependency, it is no longer feasible to include a MIDI soundfont with Alda. It is easy to install the FluidR3 soundfont into your Java Virtual Machine, and this is what we recommend doing. We've made this even easier (for Mac & Linux users, at least) by including a script (`scripts/install-fluid-r3`). Running it will download FluidR3 and replace `~/.gervill/soundbank-emg.sf2` (your JVM's default soundfont) with it. (If you're a Windows user and you know how to install a MIDI soundfont on a Windows system, please let us know!)

---

## 0.14.2 (11/13/15)

* Minor aesthetic fixes to the way errors are reported in the Alda REPL and when using the `alda parse` task.

## 0.14.1 (11/13/15)

* Improved parsing performance, especially noticeable for larger scores. More information [here](https://github.com/alda-lang/alda/issues/143), but the TL;DR version is that we now parse each instrument part individually using separate parsers, and we also make an initial pass of the entire score to strip out comments. This should not be a breaking change; you may notice that it takes less time to parse large scores.

* As a consequence of the above, there is no longer a single parse tree for an entire score, which means parsing errors are less informative and potentially more difficult to understand. We're viewing this as a worthwhile trade-off for the benefits of improved performance and better flexibility in parsing as Alda's syntax grows more complex.

* Minor note that will not affect most users: `alda.parser/parse-input` no longer returns an Instaparse failure object when given invalid Alda code, but instead throws an exception with the Instaparse failure output as a message.

## 0.14.0 (10/20/15)

* [Custom events](doc/inline-clojure-code.md#scheduling-custom-events) can now be scheduled via inline Clojure code.

* Added `electric-bass` alias for `midi-electric-bass-finger`.

---

## 0.13.0 (10/16/15)

* Note lengths can now be optionally specified in seconds (`c2s`) or milliseconds (`c2000ms`).

* [Repeats](doc/repeats.md) implemented.

---

## 0.12.4 (10/15/15)

* Added `:quit` to the list of commands available when you type `:help`.

## 0.12.3 (10/15/15)

* There is now a help system in the Alda REPL. Enter `:help` to see all available commands, or `:help <command>` for additional information about a command.

## 0.12.2 (10/13/15)

* Fix bug re: nested CRAM rhythms. (#124)

## 0.12.1 (10/8/15)

* Fix minor bug in Alda REPL where ConsoleReader was trying to expand `!` characters like bash does. (#125)

## 0.12.0 (10/6/15)

* [CRAM](doc/cram.md), a fun way to represent advanced rhythms ([crisptrutski]/[daveyarwood])

---

## 0.11.0 (10/5/15)

* Implemented code block literals, which don't do anything yet, but will pave the way for features like repeats.

* `alda-code` function added to the `alda.lisp` namespace, for use in inline Clojure code. This function takes a string of Alda code, parses and evaluates it in the context of the current score. This is useful because it allows you to build up a string of Alda code programmatically via Clojure, then evaluate it as if it were written in the score to begin with! More info on this in [the docs](doc/inline-clojure-code.md#evaluating-strings-of-alda-code).

---

## 0.10.4 (10/5/15)

* Bugfix (#120), don't allow negative note lengths.

* Handy `alda script` task allows you to print the latest alda script to STDOUT, so you can pipe it to wherever you keep it on your `$PATH`, e.g. `alda script > /usr/local/bin/alda`.

## 0.10.3 (10/4/15)

* Fix edge case regression caused by the 0.10.2.

## 0.10.2 (10/4/15)

* Fix bug in playback `from`/`to` options where playback would always start at offset 0, instead of whenever the first note in the playback slice comes in.

## 0.10.1 (10/4/15)

* Fix bug where playback hangs if no instruments are defined (#114)
  May have also caused lock-ups in other situations also.

## 0.10.0 (10/3/15)

* `from` and `to` arguments allow you to play from/to certain time markings (e.g. 1:02 for 1 minute, 2 seconds in) or markers. This works both from the command-line (`alda play --from 0:02 --to myMarker`) and in the Alda REPL (`:play from 0:02 to myMarker`). ([crisptrutski])

* Simplify inline Clojure expressions -- now they're just like regular Clojure expressions. No monkey business around splitting on commas and semicolons.

### Breaking changes

* The `alda` script has changed in order to pave the way for better/simpler inline Clojure code evaluation. This breaks attribute-setting if you're using an `alda` script from before 0.10.0. You will need to reinstall the latest script to `/usr/local/bin` or wherever you keep it on your `$PATH`.

* This breaks backwards compatibility with "multiple attribute changes," i.e.:

        (volume 50, tempo 100)

    This will now attempt to be read as a Clojure expression `(volume 50 tempo 100)` (since commas are whitespace in Clojure), which will fail because the `volume` function expects only one argument.

    To update your scores that contain this syntax, change the above to:

        (do (volume 50) (tempo 100))

    or just:

        (volume 50) (tempo 100)

---

## 0.9.0 (10/1/15)

* Implemented panning via the `panning` attribute.

---

## 0.8.0 (9/30/15)

* Added the ability to specify a key signature via the `key-signature` attribute. Accidentals can be left off of notes if they are in the key signature. See [the docs](doc/attributes.md#key-signature) for more info on how to use key signatures. ([FragLegs]/[daveyarwood])

* `=` after a note is now parsed as a natural, e.g. `b=` is a B natural. This can be used to override the key signature, as in traditional music notation.

---

## 0.7.1 (9/26/15)

* Fixed a couple of bugs around inline Clojure code. ([crisptrutski])

## 0.7.0 (9/25/15)

### New features

* Alda now supports inline Clojure code! Anything between parentheses is interpreted as a Clojure expression and evaluated within the context of the `alda.lisp` namespace.
To preserve backwards compatibility, attributes still work the same way -- they just happen to be function calls now -- and there is a special reader behavior that will split an S-expression into multiple S-expressions if there is a comma or semicolon, so that there is even backwards compatibility with things like this: `(volume 50, tempo! 90)` (under the hood, this is read by the Clojure compiler as `(do (volume 50) (tempo! 90))`).

### Breaking changes

* Alda no longer has a native `(* long comment syntax *)`. This syntax will now be interpreted as a Clojure S-expression, which will fail because it will try to interpret everything inside as Clojure values and multiply them all together :) The "official" way to do long comments in an Alda score now is to via Clojure's `comment` macro, or you can always just use short comments.

### Other changes

* Bugfix: The Alda REPL `:play` command was only resetting the current/last offset of all the instruments for playback, causing inconsistent playback with respect to other things like volume and octave. Now it resets all of the instruments' attributes to their initial values, so it is truly like they are starting over from the beginning of the score.

---

## 0.6.4 (9/22/15)

* Bugfix: parsing no longer fails when following a voice group with an instrument call.

## 0.6.3 (9/19/15)

* Fixed another regression caused by 0.6.1 -- tying notes across barlines was no longer working because the barlines were evaluating to `nil` and throwing a wrench in duration calculation.

* Added a `--tree` flag to the `alda parse` task, which prints the intermediate parse tree before being transformed to alda.lisp code.

## 0.6.2 (9/18/15)

* Fixed a regression caused by 0.6.1 -- the `barline` function in `alda.lisp.events.barline` wasn't actually being loaded into `alda.lisp`. Also, add debug log that this namespace was loaded into `alda.lisp`.

## 0.6.1 (9/17/15)

* Bar lines are now parsed as events (events that do nothing when evaluated) instead of comments; this is done in preparation for being able to generate visual scores.

## 0.6.0 (9/11/15)

* Alda REPL `:play` command -- plays the current score from the beginning. ([crisptrutski]/[daveyarwood])

---

## 0.5.4 (9/10/15)

* Allow quantization > 100% for overlapping notes. ([crisptrutski])

## 0.5.3 (9/10/15)

Exit with error code 1 when parsing fails for `alda play` and `alda parse` tasks. ([MadcapJake])

## 0.5.2 (9/9/15)

* Bugfix: add any pre-buffer time to the synchronous wait time -- keeps scores from ending prematurely when using the `alda play` task.
* Grammar improvement: explicit `octave-set`, `octave-up` and `octave-down` tokens instead of one catch-all `octave-change` token. ([crisptrutski][crisptrutski])

## 0.5.1 (9/8/15)

* Pretty-print the results of the `alda parse` task.

## 0.5.0 (9/7/15)

* New Alda REPL commands:
  * `:load` loads a score from a file.
  * `:map` prints the current score (as a Clojure map of data).
  * `:score` prints the current score (Alda code).

---

## 0.4.5 (9/7/15)

* Turn off debug logging by default. WARN is the new default debug level.
* Debug level can be explicitly set via the `TIMBRE_LEVEL` environment variable.

## 0.4.4 (9/6/15)

* Bugfix/backwards compatibility: don't use Clojure 1.7 `update` command.

## 0.4.3 (9/5/15)

* Don't print the score when exiting the REPL (preparing for the `:score` REPL command which will print the score whenever you want.

## 0.4.2 (9/4/15)

* `help` and `version` tasks moved to the top of help text

## 0.4.1 (9/4/15)

* `alda help` command

## 0.4.0 (9/3/15)

* `alda` executable script
* version number now stored in `alda.version`
* various other improvements/refactorings

---

## 0.3.0 (9/1/15)

* Long comment syntax changed from `#{ this }` to `(* this *)`.

---

## 0.2.1 (8/31/15)

* `alda play` task now reports parse errors.

## 0.2.0 (8/30/15)

* `alda.sound/play!` now automatically determines the audio types needed for a score, making `alda.sound/set-up! <type>` optional.

* various internal improvements / refactorings

---

## 0.1.1 (8/28/15)

* Minor bugfix, `track-volume` attribute was not being included in notes due to a typo.

## 0.1.0 (8/27/15)

* "Official" first release of Alda. Finally deployed to clojars, after ~3 years of tinkering.

[daveyarwood]: https://github.com/daveyarwood
[crisptrutski]: https://github.com/crisptrutski
[MadCapJake]: https://github.com/MadcapJake
[FragLegs]: https://github.com/FragLegs
[jgkamat]: https://github.com/jgkamat
[heikkil]: https://github.com/heikkil
[elyisgreat]: https://github.com/elyisgreat
[jgerman]: https://github.com/jgerman
[aengelberg]: https://github.com/aengelberg
[goog]: https://github.com/goog
[0atman]: https://github.com/0atman
[feldoh]: https://github.com/feldoh

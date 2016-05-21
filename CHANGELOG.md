# CHANGELOG

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


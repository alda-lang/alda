# CHANGELOG

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

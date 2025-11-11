# CHANGELOG

## 2.3.5 (2025-11-11)

* Further improvements to make REPL note timing more robust.

## 2.3.4 (2025-11-08)

* Fixed [an issue][issue-415] where an accumulating delay would
  occur in the REPL when using voices. A big thank you to
  [JakubSlacht] for their dedication in helping to solve this mystery!

* Part IDs (as seen in `:score info`) are now short, sequential, and
  human-readable (e.g. "part001") instead of memory addresses.

* The `:parts` and `:score info` REPL commands now sort parts by ID.

## 2.3.3 (2025-09-19)

* Alda now officially supports Linux on a wider range of ARM devices.
  This includes both 32-bit (e.g. Raspberry Pi 2/3) and 64-bit (e.g.
  Raspberry Pi 4 and newer) processors.

## 2.3.2 (2025-05-20)

* Fixed [an issue][issue-401] where nested repeats weren't working as expected.
  Thanks so much to [De-Alchmst] for the fix!

* Fixed spacing in the `:help play` output.

## 2.3.1 (2024-08-04)

Fixed an issue where ANSI color codes (e.g. `←[36m`) were being displayed when
running Alda in Powershell. Thanks, [Vanello1908], for the contribution!

See [this issue][issue-405] for further discussion. There is still a remaining
issue where the background color is changing unexpectedly, but as of this
release, the ANSI codes should at least be gone.

## 2.3.0 (2024-06-22)

### Improved MIDI channel assignment

> Incidental note: This release includes some **breaking changes** to the Alda
> OSC API, which is the communication layer between the Alda client and player
> processes:
>
> * Added a required channel number argument to a handful of endpoints.
>
> * Removed `/track/{number}/midi/percussion` endpoint, which is no longer
>   needed.
>
> * Removed unused mute/unmute functionality.
>
> These changes will not affect the vast majority of Alda users. The only way
> that you might run into issues is if you have written software that interfaces
> with the Alda player process directly, instead of using the Alda client.
>
> The general Alda experience still works exactly the same way as it did before,
> so the aforementioned changes probably won't affect you.

As of this release, we've re-worked the way that MIDI channels are assigned to
parts. This addresses [some shortcomings][midi-channel-assignment-discussion]
that have have come up in discussions a few times. To summarize, prior to this
release:

* It was not possible to include more than 16 instruments in a score.

* There was no visible error message when a score exceeded the number of
  available MIDI channels (16). The player process would log an error and do its
  best to play the score (leaving out some parts), but the client would just say
  "Playing..." and it was not obvious that there was a problem.

* MIDI channel assignment order was non-deterministic. This meant that when you
  exported scores to MIDI files, the instruments used in each channel were not
  necessarily in the same order that they appeared in the score.

* There was no way to control which MIDI channel was used for each part.

As of this release:

* MIDI channel assignment is now handled on the client side, and is **fully
  deterministic**. The order of parts presented in the score maps to the order
  of channels in exported MIDI files in an intuitive way.

* The Alda client provides **helpful error messages** in cases where it is not
  possible to use the 16 available MIDI channels to play your score.

* You can now have **more than 16 instruments in a score**, just as long as
  there are never more than 15 non-percussion instruments playing _at the same
  time_. I wrote a handful of fun, new example scores to show you what's
  possible:

  * [`all-instruments.alda`][all-instruments]: A demo of **all 128 General MIDI
    instruments**, played back to back

  * [`midi-channel-management.alda`][midi-channel-management]: A quick jazzy
    number featuring **31 instruments**: 1 percussion part playing throughout,
    and two groups of 15 instruments playing together at a time.

  * [`midi-channel-management-2.alda`][midi-channel-management-2]: A tiny
    example showing how the new `midi-channel` attribute works.

* There is a new [`midi-channel` attribute][midi-channel-attribute] that allows
  you to explicitly specify which MIDI channel should be used at that point in
  time for the notes in that part. Most of you will never need to use this
  attribute, because Alda does a good job of automatically assigning MIDI
  channels to parts for you. But for those of you who want more control over
  MIDI channel assignment, this attribute will give you that power.

I hope you enjoy Alda's newfound ability to handle large numbers of instruments
in a score. I'm pretty excited about it! As always, please let us know if you
notice any bugs or unexplained behavior!

### Other misc. changes

* Fixed a bug in `alda doctor` where it would hang if there was an error during
  the "Send and receive OSC messages" step. Now if there's an error during that
  step, it will print the error message.

* Client/player communication is now done via `127.0.0.1` by default, instead of
  `localhost`. These are effectively the same thing, but it's possible for
  `localhost` not to work, depending on your host configuration. `127.0.0.1` is
  more likely to work.

## 2.2.7 (2023-09-01)

Added a `pid` column to the output of `alda ps`, the value of which is the
process ID of each player and REPL server process in the list.

## 2.2.6 (2023-08-18)

Upgraded Go library dependencies to patch security vulnerabilities.

## 2.2.5 (2023-05-07)

* Corrected the casing of the word "MIDI" (it was "Midi" before) in the error
  message when a MIDI Note is outside of the 0-127 range.

* The new `alda import` command can import a MusicXML file and produce a working
  Alda score. I'm excited about this - big thanks to [Scowluga] and [alan-ma]
  for their hard work on this exciting new feature!

  See [this blog post][alda-import-blog-post] for an overview of what `alda
  import` can do and how to use it.

## 2.2.4 (2022-11-24)

* Added validation for when a MIDI note is outside of the 0-127 range.

  Thanks to [kylewilk567] for the contribution!

## 2.2.3 (2022-04-24)

* Added a new `:parts` command that can be used during an Alda REPL session. It
  prints information about the parts in the current score.

  (To display more information about the current score, you can also use the
  existing `:score info` command.)

  Thanks, [n-makim], for the contribution!

## 2.2.2 (2022-04-17)

This patch release is all about improvements to the way that player processes
are managed in an Alda REPL session.

Thanks to [elyisgreat] for [reporting the issue][issue-404] and to [ksiyuan] for
investigating and [contributing a fix][pr-418]!

* Fixed a bug causing `:stop` to sometimes not work in an Alda REPL session.

* Fixed spurious "Failed to read player state" warnings that were often
  happening briefly while a player process is starting.

* Fixed a potential edge case where, when using the Alda REPL, if a player
  process unexpectedly shuts down (not common), the Alda REPL session might
  continue to try to use the same player process.

[issue-404]: https://github.com/alda-lang/alda/issues/404
[pr-418]: https://github.com/alda-lang/alda/pull/418

## 2.2.1 (2022-04-10)

* Re-added the `pause` (i.e. rest) Lisp function that was available prior to
  Alda 2.0.0, but accidentally omitted during the rewrite.

  Thanks, [JustinLocke] for the contribution, and [UlyssesZh] for
  reporting [the issue][issue-382]! :raised_hands:

[issue-382]: https://github.com/alda-lang/alda/issues/382

## 2.2.0 (2022-01-15)

* On Mac computers, Alda now requires macOS 10.13 (High Sierra) or later.

* There is now an experimental WebAssembly build of Alda! You can't do much with
  it yet, but hopefully in the near future, we'll be able to run Alda in the
  browser! :open_mouth:

  If you're interested in following the discussion (or maybe even
  contributing!), see [this issue][alda-in-the-browser].

[alda-in-the-browser]: https://github.com/alda-lang/alda/issues/392

## 2.1.0 (2021-12-29)

* As a step to enable future work on exciting new features like automatic code
  formatting and importing from other formats like MusicXML, we have done some
  minor, under-the-hood refactoring of the Alda parser.

  Prior to this release, the Alda parser took a shortcut in that it had a step
  where it converted a list of tokens directly into a list of "score updates"
  (notes, chords, etc.). There is an important step that we were missing, which
  was producing an [AST][ast-wikipedia]. Whereas the parsing steps used to be:

  ```
  characters -> tokens -> score updates
  ```

  Now, the steps are:

  ```
  characters -> tokens -> AST -> score updates
  ```

  Although this is a minor refactor, there is some risk of breakage, so please
  [open an issue][open-an-issue] if you notice any problems!

* We've added options for outputting the parsed AST of a score. This can be
  useful for debugging potential parser errors, as well as for building tooling
  (e.g. in text editors) that depends on the AST of an Alda source file.

  * `alda parse` now has `--output ast` and `--output ast-human` options.

    `ast` output is the AST in data form, represented as a single JSON object.
    Each AST node is an object that includes the keys `type` and (if the node
    has other nodes as children) `children`:

    ```
    $ alda parse -c 'piano: c/e/g+' -o ast
    {"children":[{"children":[{"children":[{"children":[{"literal":"piano","source-context":{"column":1,"line":1},"type":"PartNameNode"}],"source-context":{"column":1,"line":1},"type":"PartNamesNode"}],"source-context":{"column":1,"line":1},"type":"PartDeclarationNode"},{"children":[{"children":[{"children":[{"children":[{"literal":"c","source-context":{"column":8,"line":1},"type":"NoteLetterNode"}],"source-context":{"column":8,"line":1},"type":"NoteLetterAndAccidentalsNode"}],"source-context":{"column":8,"line":1},"type":"NoteNode"},{"children":[{"children":[{"literal":"e","source-context":{"column":10,"line":1},"type":"NoteLetterNode"}],"source-context":{"column":10,"line":1},"type":"NoteLetterAndAccidentalsNode"}],"source-context":{"column":10,"line":1},"type":"NoteNode"},{"children":[{"children":[{"literal":"g","source-context":{"column":12,"line":1},"type":"NoteLetterNode"},{"children":[{"source-context":{"column":13,"line":1},"type":"SharpNode"}],"source-context":{"column":13,"line":1},"type":"NoteAccidentalsNode"}],"source-context":{"column":12,"line":1},"type":"NoteLetterAndAccidentalsNode"}],"source-context":{"column":12,"line":1},"type":"NoteNode"}],"source-context":{"column":8,"line":1},"type":"ChordNode"}],"source-context":{"column":8,"line":1},"type":"EventSequenceNode"}],"source-context":{"column":1,"line":1},"type":"PartNode"}],"type":"RootNode"}
    ```

    `ast-human` output is the AST in a more compact, human-readable format:

    ```
    $ alda parse -c 'piano: c/e/g+' -o ast-human
    RootNode
      PartNode [1:1]
        PartDeclarationNode [1:1]
          PartNamesNode [1:1]
            PartNameNode [1:1]: "piano"
        EventSequenceNode [1:8]
          ChordNode [1:8]
            NoteNode [1:8]
              NoteLetterAndAccidentalsNode [1:8]
                NoteLetterNode [1:8]: "c"
            NoteNode [1:10]
              NoteLetterAndAccidentalsNode [1:10]
                NoteLetterNode [1:10]: "e"
            NoteNode [1:12]
              NoteLetterAndAccidentalsNode [1:12]
                NoteLetterNode [1:12]: "g"
                NoteAccidentalsNode [1:13]
                  SharpNode [1:13]
    ```

  * In the Alda REPL, the `:score` command now has an `ast` option, which prints
    the human-readable version of the AST output:

    ```
    alda> bassoon: o2 f
    alda> :score ast
    RootNode
      PartNode [1:1]
        PartDeclarationNode [1:1]
          PartNamesNode [1:1]
            PartNameNode [1:1]: "bassoon"
        EventSequenceNode [1:10]
          OctaveSetNode [1:10]: 2
          NoteNode [1:13]
            NoteLetterAndAccidentalsNode [1:13]
              NoteLetterNode [1:13]: "f"
    ```

[ast-wikipedia]: https://en.wikipedia.org/wiki/Abstract_syntax_tree
[open-an-issue]: https://github.com/alda-lang/alda/issues/new/choose

## 2.0.8 (2021-12-20)

* Security update: upgraded log4j to version 2.17.0 to patch CVEs.

## 2.0.7 (2021-12-15)

* Security update: upgraded log4j to version 2.16.0 to patch CVEs.

## 2.0.6 (2021-10-04)

* Fixed [a bug][issue-398] where a note length of 0 (e.g. `c0`) was accepted,
  and the note would play forever. A note length of `0`, `0s` or `0ms` now
  results in a validation error.

[issue-398]: https://github.com/alda-lang/alda/issues/398

## 2.0.5 (2021-08-22)

* Fixed [a bug][issue-388] where using an octave change inside a chord inside a
  cram expression produced unexpected results.

* Fixed [a bug][issue-389] where playback was sometimes ending abruptly before
  the end of the score.

[issue-388]: https://github.com/alda-lang/alda/issues/388
[issue-389]: https://github.com/alda-lang/alda/issues/389

## 2.0.4 (2021-08-14)

* `alda shutdown` now prints a helpful message letting you know what it is
  doing (shutting down player processes).

* If `alda stop` or `alda shutdown` fails to send a "stop" or "shutdown" message
  to a player process, it will now print a warning and continue instead of
  printing an error and exiting. (This scenario is usually not a critical
  problem, and it will resolve itself within a couple of minutes.)

* `alda play` and `alda export` are now more resilient against scenarios where
  old player processes died mysteriously and left around stale state files that
  suggest they are still reachable.

  In scenarios like those, there will now be a long pause while the Alda client
  attempts to reach the dead player process, then it will print a warning saying
  it was unable to do so, and proceed to try another player process. This might
  happen a few times, but Alda will eventually recover and proceed to
  play/export your score.

  Note that this should rarely, if ever, happen! If you are seeing this happen a
  lot, then there is probably something weird going on with your player
  processes. Please have a look at the player logs (run `alda-player info` to
  learn where to find the logs) and let us know if you see any errors or
  stacktraces. This information will help us make further improvements in the
  future!

## 2.0.3 (2021-08-01)

* Fixed a bug where input like `[c1s]` (a duration in seconds at the end of an
  event sequence) was causing a parse error.

* Fixed a sporadic runtime error where this message would appear:

  `panic: runtime error: invalid memory address or nil pointer dereference`

* `alda ps` output now includes Alda REPL servers in addition to player
  processes. Example output:

  ```
  $ alda ps | column -t -s $'\t'
  id   port   state   expiry              type
  itv  33659  ready   5 minutes from now  player
  lhx  36583  active  5 minutes from now  player
  olt  34539  ready   5 minutes from now  player
  utj  40235  ready   5 minutes from now  player
  yae  35935  ready   7 minutes from now  player
  zew  40425  ready   6 minutes from now  player
  itp  33643  -       -                   repl-server
  jom  34191  -       -                   repl-server
  ```

## 2.0.2 (2021-07-31)

* Fixed a "stale state" bug where Alda would occasionally attempt to use old
  player processes that are no longer running. Whereas before, only player
  processes would clean up stale state files, now the client cleans them up too,
  to ensure that the information is up to date at the point in time when the
  client needs it.

  For more information, see [issue #369][issue-369].

* Related to the above, the `alda` client and `alda-player` processes now
  consider a state file to be "stale" if it hasn't been updated in 2 minutes,
  instead of 10 minutes.

[issue-369]: https://github.com/alda-lang/alda/issues/369

## 2.0.1 (2021-07-05)

* Alda will now attempt to detect if it's running in an environment (e.g.
  the CMD program that ships with Windows 7) that does not support ANSI escape
  codes to display colored text. If the environment does not appear to support
  ANSI escape codes, Alda will not display colored text (which is better in that
  case because otherwise you would see a bunch of weird-looking characters in
  places where there should be colored text!).

* Prior to this release, it wasn't obvious that it's incorrect to enter a
  command like:

  ```
  alda play my-score.alda
  ```

  The correct way to specify a score file to read is to use the `-f, --file`
  option:

  ```
  alda play -f my-score.alda
  ```

  Instead of silently ignoring the provided file name, the Alda CLI will now
  print a useful error message.

## 2.0.0 (2021-06-30)

Alda 2 is a from-the-ground-up rewrite, optimized for simpler architecture,
better performance, and a foundation for future work to enable fun live coding
features.

For information about what's new, what's changed, and what to expect, check out
the [Alda 2 migration guide][migration-guide]!

[migration-guide]: https://github.com/alda-lang/alda/blob/master/doc/alda-2-migration-guide.md

---

## Earlier Versions

* [1.0.0 - 1.X.X](CHANGELOG-1.X.X.md)
* [0.1.0 - 0.X.X](CHANGELOG-0.X.X.md)

[JustinLocke]: https://github.com/JustinLocke
[UlyssesZh]: https://github.com/UlyssesZh
[elyisgreat]: https://github.com/elyisgreat
[ksiyuan]: https://github.com/ksiyuan
[n-makim]: https://github.com/n-makim
[kylewilk567]: https://github.com/kylewilk567
[Scowluga]: https://github.com/Scowluga
[alan-ma]: https://github.com/alan-ma
[Vanello1908]: https://github.com/Vanello1908
[De-Alchmst]: https://github.com/De-Alchmst
[JakubSlacht]: https://github.com/JakubSlacht

[alda-import-blog-post]: https://blog.djy.io/musicxml-import-and-another-new-alda-features/
[midi-channel-assignment-discussion]: https://github.com/alda-lang/alda/discussions/447
[all-instruments]: ./examples/all-instruments.alda
[midi-channel-management]: ./examples/midi-channel-management.alda
[midi-channel-management-2]: ./examples/midi-channel-management-2.alda
[midi-channel-attribute]: ./doc/attributes.md#midi-channel
[issue-401]: https://github.com/alda-lang/alda/issues/401
[issue-405]: https://github.com/alda-lang/alda/issues/405
[issue-415]: https://github.com/alda-lang/alda/issues/415

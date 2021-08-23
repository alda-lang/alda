# CHANGELOG

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

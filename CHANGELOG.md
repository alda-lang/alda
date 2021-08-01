# CHANGELOG

## 2.0.3 (2021-08-01)

* Fixed a bug where input like `[c1s]` (a duration in seconds at the end of an
  event sequence) was causing a parse error.

* Fixed a sporadic runtime error where this message would appear:

  `panic: runtime error: invalid memory address or nil pointer dereference`

* `alda ps` output now includes Alda REPL servers in addition to player
  processes. Example output:

  ```
  $ bin/run ps | column -t -s $'\t'
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

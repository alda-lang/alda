# Repeats

[Notes](notes.md), [sequences](sequences.md) and other types of events can be **repeat**ed any number of times, by simply appending `*` and a number. Putting whitespace between the event and the `*` is optional.

For example:

```
piano:
  # repeating single notes
  c *4 c2 *2

  # repeating a sequence containing notes and an octave change
  [c8 d e >]*3
```

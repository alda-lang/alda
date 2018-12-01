# Repeats

[Notes](notes.md), [sequences](sequences.md) and other types of events can be **repeat**ed any number of times, by simply appending `*` and a number. Putting whitespace between the event and the `*` is optional, as is putting whitespace between the `*` and the number of repeats.

For example:

```
piano:
  # repeating single notes
  c *4 c2 *2

  # repeating a sequence containing notes and an octave change
  [c8 d e >]*3
```

## Variations

The Alda language has a feature that can be used to implement the ["alternate
endings"](http://dictionary.onmusic.org/terms/4798-second_ending_735) often
found in Western musical notation:

```
piano:
  [ c8 d e f
    [g f e4]'1-3
    [g a b > c4.]'4
  ]*4
```

This allows you to have repeated phrases that can differ on each iteration. In
the example above, each repeat starts with `c8 d e f`; on times 1 through 3
through the repeated phrase, the phrase ends with `g f e4`, whereas on the 4th
time through, the phrase ends with `g a b > c4.`.

Note that these "adjustments" can occur anywhere within the repeated phrase, not
necessarily at the end, making this feature of Alda more flexible than the
"alternate endings" notation seen in sheet music. To illustrate this, here is
another example where the phrase has what you might describe as an "alternate
beginning" and an "alternate middle":

```
piano:
  [ [c8 d e]'1,3 [e8 d c]'2,4
    f
    [g f e]'1-3 [g a b]'4
    > c <
  ]*4
```


# Writing music programmatically

Alda is designed to be easy to learn and use, even for those who have little to
no programming experience. The language is deliberately simple, omitting the
complex features that you see in most programming languages, like functions,
classes, and types.

For example, a C major scale played on an acoustic guitar in eighth notes,
starting on C3 _(that's the note C in octave number 3)_ is written like this:

```
acoustic-guitar:
  o3 c8 d e f g a b > c
```

At the same time, an important goal of Alda is to facilitate writing music
programmatically. You can get a lot of mileage out of the Alda language alone,
but if you happen to have a little bit of programming knowledge, you can go
beyond thinking of Alda as a **language** and start to think of it as a
**platform** for algorithmic composition and live coding.

## What is algorithmic composition?

Simply put, [algorithmic
composition](https://en.wikipedia.org/wiki/Algorithmic_composition) is when you
write music in a way that leaves some part of the process up to chance. Which
parts are left up to chance, and how exactly the "chance" parts are implemented,
is entirely open-ended.

For example, you could write a piece of music where you come up with a set of
rhythms, as well as a set of pitches, and then you write a computer program that
puts together random combinations of the rhythms and pitches to create a piece
of music.

Another example: you could write a program that fetches the 10-day weather
forecast for your area and uses some arbitrary rules to convert it into music.
For instance, maybe if the temperature is an odd number, a major scale will be
used; or maybe if the forecast calls for snow, one of the instruments will be a
cello; etc.

The possibilities are endless!

## What is live coding?

[Live coding](https://en.wikipedia.org/wiki/Live_coding) is the practice of
creating music (or other types of art) on the fly, in a live performance
setting, by programming.

Oftentimes, the performance will include a projection of the programmer's
screen, so that the audience can watch the code being written and evaluated and
observe the art being created in real time.

## Alda as a platform for algorithmic composition and live coding

As mentioned above, Alda the language does not provide facilities like functions
and random number generators that one would ordinarily use to write algorithmic
compositions.

However, the benefit of Alda's simplicity is that it can easily be used as a
foundation for more complex things.

The libraries below can be used to generate Alda code by writing a program in
another language. This technique is especially powerful when the language is one
that allows you to program interactively, e.g. in a REPL (Read-Evaluate-Print
Loop).

> If your favorite programming language is not listed here and you're interested
> in using it to generate Alda scores, please consider writing your own Alda
> live-coding library in that language and adding it to this list. It's
> surprisingly easy!

| Language | Library    | Author       |
|----------|------------|--------------|
| Clojure  | [alda-clj] | Dave Yarwood |
| Ruby     | [alda-rb]  | Ulysses Zhan |

[alda-clj]: https://github.com/daveyarwood/alda-clj
[alda-rb]: https://github.com/UlyssesZh/alda-rb

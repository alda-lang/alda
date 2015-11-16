# Variables

Repetition is an important aspect of music. In addition to repeating notes and phrases, it is often desirable to repeat larger phrases or entire sections. This can be cumbersome with Alda's [repeat](repeats.md) syntax alone. For greater flexibility and cleaner code, you can define **variables** that represent named [sequences of events](sequences.md).

## Defining a variable

To define a variable, use the syntax `variableName = events go here`, for example:

```
motif = b-8 a g f e g a4
```

This defines a variable called `motif` that can be used at any time afterward in the score as an easier way of writing the event sequence `[ b-8 a g f e g a4 ]`.

You can use multiple lines when defining a variable, if you'd prefer; the trick is to use a multi-line event sequence:

```
motif = [
  b-8 a g f
  e g a4
]
```

## Using a variable

To use a variable in your score after defining it, simply use its name inside of an instrument part:

```
piano:
  o2 motif < motif d1
```

Note that variables can be repeated in the same way as events and event sequences:

```
piano:
  motif *3
```

## Variables can be aliases

Strictly speaking, the value of a variable does not need to be an event sequence; it can be an individual event. Why would you do this? You may find it convenient to create aliases for [attribute](attributes.md) changes to specific values:

```
quiet  = (vol 25)
loud   = (vol 50)
louder = (vol 75)

notes  = c d e

piano:
  quiet notes
  loud notes
  louder notes
```

## Variables in variables

Previously defined variables can be used in the definition of other variables. This concept allows you to build up scores from smaller components.

```
notes = c d e
moreNotes = f g a b
lastOne = > c

cMajorScale = notes moreNotes lastOne

piano:
  cMajorScale
```

## Acceptable variable names

Variable names must adhere to the following rules:

* They must be at least 2 characters long.
* The first two characters must be letters (either uppercase or lowercase).
* After the first two characters, they may contain any combination of:
  * letters (upper- or lowercase)
  * digits 0-9
  * underscores `_`

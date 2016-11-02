# Instance and Group Assignment

> Note: this is a comprehensive explanation of how instance and group assignment works, for the curious. The details below are complicated, but can be simplified as a [handful of simple rules](scores-and-parts.md#how-instances-are-assigned).
>
> You may want to follow that link if you aren't concerned with the gory details!

There are four categories of **instrument calls** in Alda. Each category either _has a nickname_ or doesn't, and either _has multiple instances_ or doesn't.

### `foo:`

- If `foo` refers to a previously named instrument or group, e.g. `piano-1:`...
  - ...refers to that instrument or group.

- If `foo` is a stock instrument, e.g. `piano:`...
  - ...and we don't have a `piano` yet in the score...
    - ...creates a new instance of the stock instrument`piano`.
    - ...subsequent calls to `piano` will reference that instrument.
  - ...and we already have a NAMED `piano` in the score...
    - ...throws an ambiguity error. (Any additional `piano`s must also be named.)
  - ...and we already have exactly one `piano`, and it doesn't have a name...
    - ...refers to that `piano`.

- Else, throws an "unrecognized instrument" error.

### `foo "bar":`

- `foo` is expected to be a stock instrument. If it's not, an error will be thrown.

- If `"bar"` was already used as the nickname of another instance, throws an error.

- If there is an existing, unnamed instance of `foo` in the score, throws an error. (All instances of `foo` must be named.)

- Creates a new instance named `bar` of type `foo`, e.g.:
  - `piano "larry":` creates a new instance of the stock `piano` named `"larry"`
  - Subsequently in the score, `larry:` refers to that instance.

### `foo/bar:`

- If `foo` and `bar` are the same named instance, e.g. `foo/foo`...
  - ...throw an error because that doesn't make any sense.

- If `foo` and `bar` are the same stock instrument, e.g. `piano/piano`...
  - ...throw an error because the next time you call `piano:`, it won't be clear which one you mean.

- If both `foo` and `bar` refer to previously named instrument instances...
  - ...refers to those instances.

- If both `foo` and `bar` are stock instruments, e.g. `piano/bassoon:`
  - ...follows the `foo:` rules above to select and/or create instances.
    - (new instances will be created for any instruments that don't exist yet in the score)

- If e.g. `foo` is a named instrument and `bar` is not, or vice versa (e.g. `foo/trumpet:`):
  - ...throws an ambiguity error. (Nicknames should be used for creating new instances or grouping existing ones, not both.)

### `foo/bar "baz":`

- If `foo` and `bar` are the same named instance, e.g. `foo/foo`...
  - ...throw an error because that doesn't make any sense.

- If `foo` and `bar` are the same stock instrument, e.g. `piano/piano`...
  - ...throw an error because if you wanted to call `baz.piano:` to refer to one of the pianos, it won't be clear which one you mean.
  - So the moral of the story is, if you want a group containing two of the same instrument, you have to name both instances individually first.

- If both `foo` and `bar` refer to previously named instrument instances...
  - ...refers to those instances.
  - ...creates an alias `"baz"` which can now be used to refer to those instances as a group.
  - ...subsequently, `baz.foo` and `baz.bar` are available to reference the group members, although you could also just keep calling them `foo` and `bar`.

- If both `foo` and `bar` are stock instruments, e.g. `piano/guitar "floop":`...
  - ...creates new instances for each group member.
  - ...creates an alias `"floop"` which can now be used to refer to those instances as a group.
  - ...subsequently, `floop.piano` and `floop.guitar` are available to reference the group members.

- If `foo` is a named instrument and `bar` is a stock instrument or vice versa, e.g. `foo/trumpet "quux":`...
  - ...throws an ambiguity error. (Groups should be used for creating new instances or grouping existing ones, not both.)

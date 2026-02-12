# Implementing an Alda library

The article on [writing music programmatically][writing-music-programmatically]
explains how Alda is designed to be controlled easily by other programming
languages. The idea is that Alda is intended to be not just a simple
**language** for music composition, but also a **platform** for the programmatic
creation of music.

This simplicity in the design of Alda gives programmers the freedom to create
libraries for driving Alda in just about any programming language that they want
to use.

If you're interested in helping to extend support for Alda to users of your
programming language of choice, read on!

## Example: alda-clj

The Clojure library [alda-clj] is the canonical reference implementation of a
programming language library for Alda. Dave Yarwood (who is also the creator of
Alda) wrote alda-clj with the goal of being able to write Clojure code that will
generate interesting musical scores by taking random or "found" data (e.g.
weather forecasts) and translating it into Alda source code.

Here are a couple of basic examples showing how you can use alda-clj in a
Clojure program or REPL session:

```clojure
;; Import all of the functions from alda.core into the current context
(require '[alda.core :refer :all])

;; Play a whole note G2 on a trombone
(play!
  (part "trombone")
  (octave 2)
  (note (pitch :g) (note-length 1)))

;; Play "C D E F G" 4 times at random tempos in the range of 60-260 BPM
(play!
  (part "piano")
  (for [bpm (repeatedly 4 #(+ 60 (rand-int 200)))]
    [(tempo bpm)
     (for [letter [:c :d :e :f :g]]
       (note (pitch letter) (note-length 8)))]))
```

Every programming language is unique. Naturally, an Alda library written for
another language (Go, Java, Erlang, JavaScript, etc.) is going to look a bit
different from the code above, and each of these libraries would also look
different from each other.

Despite this, the principles remain the same. The approach to writing an Alda
library is simple enough that you can follow the same steps in just about _any_
programming language and you will end up with a library that lets you write code
in that language that lets you do fun and exciting things with Alda as a
platform for generative music.

## Features of an Alda library

At a minimum, an Alda library for a programming language needs to provide these
features:

1. A function that wraps usage of the `alda` command line client.

2. An API that takes idiomatic code written in the programming language and
   converts it into valid Alda source code.

## Step-by-step guide to writing an Alda library

To make these ideas concrete, I will be using server-side JavaScript (Node.js)
as an example to show you the desired outcome of each step in the development
process.

It's worth noting that I haven't actually written a Node.js library for Alda.
The code examples below are purely illustrative; I chose JavaScript because it's
a language that a lot of people know and understand.

### Step 1: Write an `alda` command wrapper function

Our first challenge is to write a function that will invoke the `alda` command
line client and capture the output.

This function should ideally:

* Take any number of arguments and pass them to the command line client.
* Recognize when the command failed (i.e. the exit code wasn't 0) and do
  something appropriate like print the error output and/or throw an exception.

For example, to print the version of Alda, we could execute the following code:

```javascript
let versionOutput = alda("version");
console.log(versionOutput);
```

And it would capture the output of `alda version` and print it:

```
alda 2.0.0
```

To play an Alda source file, we could run:

```javascript
// We are ignoring the output here, so we aren't printing it.
alda("play", "-f", "/home/dave/alda-scores/we-built-this-city.alda");
```

Most server-side programming languages have good functionality in the standard
library for working with subshells / child processes. With Node.js, for example,
we could use the [`child_process`][child-process] module that comes with Node.

Now that you have an `alda` CLI wrapper function, you can use it to play
arbitrary strings of Alda code, like this:

```javascript
alda("play", "-c", "harp: o5 g16 f+ d+ < a g+ > e a- > c");
```

You can even define a `play` command to make this a little more concise:

```javascript
play("harp: o5 g16 f+ d+ < a g+ > e a- > c");
```

### Step 2: Come up with an API

This still isn't much better than just typing into your terminal:

```bash
alda play -c "harp: o5 g16 f+ d+ < a g+ > e a- > c"
```

We can do better.

The problem that we're trying to solve here is that we want to be able to write
programs that operate on musical **domain objects**, not clumsy strings of text.

The meaning of "domain object" here can vary, depending on what programming
language you are working in, the sort of code that it's idiomatic to write in
that language, and your own opinions as the library author.

In alda-clj, I chose to use Clojure records, which are similar to classes in
object-oriented languages like Java, Python, or Ruby. I chose to use records
because you can use them to define different types of objects that behave
differently when you ask them to do something. We want to be able to take any of
these "domain objects" and ask them to return an Alda code string version of
themselves. Here is a Clojure REPL session that illustrates how this works in
alda-clj:

```clojure
user=> (note (pitch :c))
#alda.core.Note{
  :pitch #alda.core.LetterAndAccidentals{:letter :c, :accidentals nil},
  :duration nil,
  :slurred? nil}

user=> (->str (note (pitch :c)))
"c"

user=> (chord
         (note (pitch :c))
         (note (pitch :e))
         (note (pitch :g)))
#alda.core.Chord{
  :events (
    #alda.core.Note{
      :pitch #alda.core.LetterAndAccidentals{:letter :c, :accidentals nil},
      :duration nil,
      :slurred? nil},
    #alda.core.Note{
      :pitch #alda.core.LetterAndAccidentals{:letter :e, :accidentals nil},
      :duration nil,
      :slurred? nil},
    #alda.core.Note{
      :pitch #alda.core.LetterAndAccidentals{:letter :g, :accidentals nil},
      :duration nil,
      :slurred? nil})}

user=> (->str (chord
                (note (pitch :c))
                (note (pitch :e))
                (note (pitch :g))))
"c / e / g"
```

The alda-clj API is a domain-specific language (DSL) consisting of a bunch of
functions with names like `note`, `pitch` and `chord`. You can compose them
together to create the musical "thing" that you have in mind (a note, a chord,
etc.), and that "thing" is implemented as a Clojure record. When you invoke
alda-clj's `->str` function on one of these things, the return value is a string
of valid Alda code.

Why is this useful? Because now we're in a programming environment, and we can
do all sorts of interesting things that will dynamically generate Alda code. For
example, you could write a function that returns a chord with a random number of
notes in it, and the notes themselves are randomly selected from a list of
possible choices:

```clojure
(defn random-note
  []
  (let [letter     (rand-nth [:c :d :e :f :g :a :b :c])
        accidental (rand-nth [:sharp :flat nil])]
    (note (if accidental
            (pitch letter accidental)
            (pitch letter)))))

(defn random-chord
  []
  (apply chord (repeatedly (rand-int 6) random-note)))

(->str (random-chord)) ;=> "d+ / c- / c+ / a+"
(->str (random-chord)) ;=> "a / c+ / a / a-"
(->str (random-chord)) ;=> "d- / c+"
(->str (random-chord)) ;=> "c+"
(->str (random-chord)) ;=> "c+ / e+"
(->str (random-chord)) ;=> "d- / d / f- / g+"

;; Play 3 randomly generated chords in whole notes.
(play!
  (part "piano")
  (set-note-length 1)
  (random-chord)
  (random-chord)
  (random-chord))
```

By providing an API that allows users to create the basic domain objects (like
`note` and `chord`), we're enabling users to put them together in all kinds of
interesting ways, limited only by their creativity.

If our JavaScript Alda API was written in a functional style similar to
alda-clj, then basic usage of the library might look something like this:

```javascript
play(
  part("piano"),
  note(pitch("c"), noteLength(8)),
  note(pitch("d")),
  note(pitch("e"))
  chord(
    note(pitch("g"), noteLength(1)),
    note(pitch("b"))
  )
);
```

Or, if you prefer a more object-oriented style, you might create an API that
looks something like this:

```javascript
let score = new AldaScore();
score.addPart("piano");
score.addNote("c", 8);
score.addNote("d");
score.addNote("e");
// ... etc. ...

score.play();
```

There is no right or wrong answer when it comes to what the API should look like
or how it's implemented. It's up to you, the library author!

### Step 3: Generate Alda code

The key idea behind a successful Alda library is that we separate the concerns
of

1. working with musical domain objects, and
2. generating Alda source code.

You can see this idea at work in the Clojure code example above. alda-clj
provides two separate API functions, `play!` and `->str`, both of which take
musical domain objects as arguments. Under the hood, all `play!` is doing is
it's calling `->str` on its arguments to turn them into Alda source code, and
then using that string of Alda code as an argument to the `alda play` command.

Our JavaScript equivalent to `play!` might look something like this:

```javascript
function stringify(object) {
  // return a string form of the object, which will vary depending on the type
  // of the object (note, chord, rest, cram expression, etc.)
}

function play(...objects) {
  let aldaCode = objects.reduce((code, object) => {
    return code + stringify(object) + " ";
  }, "");

  alda("play", "-c", aldaCode);
}
```

The `stringify` function above is responsible for translating one of our domain
objects into a string of valid Alda source code. For example:

```javascript
let note1 = note(pitch("c"));
let note2 = note(pitch("e-flat"));
let myChord = chord(note1, note2);

console.log(stringify(note1)); // c
console.log(stringify(note2)); // e-
console.log(stringify(myChord)); // c / e-
```

A simple way to do this in JavaScript is to just write a function that includes
a big `switch` statement that checks the type of the object and acts
accordingly:

```javascript
function stringify(object) {
  switch(true) {
    case object instanceof Note:
      return object.letter + (object ?? "");

    case object instanceof Chord:
      return object.notes.map(stringify).join(" / ");

    // ... etc. for all the other types ...

    default:
      throw "Unsupported type: " + typeof object
  }
}
```

Another approach is to use "duck typing" and implement a `.stringify()` method
on each of the domain objects:

```javascript
note("c", noteLength(8)).stringify() // => "c8"
chord(note1, note2, note3).stringify() // => "c/e/g"
```

Like I said before in Step 2, there is no right or wrong way to do this. As the
library author, you have the freedom to do it any way you'd like!

### _(optional)_ Step 4: REPL integration

By now, if you've been following along at home, you have created a good Alda
library for your programming language of choice. You can use it as a CLI wrapper
to issue arbitrary Alda commands like `alda ps` and `alda doctor`. More
importantly, you can use the functions provided by your library to build scores
in a programmatic way and have some fun creating algorithmic compositions. Nice
work! :tada:

If you'd like to go a step further, you can give your library the ability to
integrate with Alda REPL servers. This would make your library suitable for
**live coding**. The idea there is that you can start to play a fragment of a
musical score (a musical idea of some sort) and then add more fragments onto the
end _while the score is playing_.

> Another benefit is that multiple users of your library can connect to the
> _same_ Alda REPL server and compose music together in real time!

This is similar to the interactive REPL experience that you get when you run
`alda repl` at the command line. You can build up your score incrementally in
small pieces.

In an Alda REPL session, you can evaluate the following Alda code:

```alda
harmonica: o5 c d e f g a
```

And then, if you immediately type the following in and press enter:

```alda
b > c
```

You will hear those last two notes played in time after the first ones, because
the second line that you entered is actually just a continuation of the score
that you started on the first line.

You can see the full score text by typing `:score text` into the REPL prompt. It
will output something like:

```
harmonica: o5 c d e f g a
b > c
```

What's happening here is that there is an Alda REPL server running in the
background, and it's keeping track of the details of our score, including
important facts like:

* The current instrument is a harmonica.
* It's playing in the 5th octave.
* It's playing quarter notes.

If you're interested in **repl**icating (_I'm sorry, I had to_) this behavior in
your library, you're in luck. The Alda REPL server has a simple JSON API and you
can send it messages via the Alda CLI.

To see this workflow in action, open two terminals. In one terminal, start an
Alda REPL server by running the following command:

```bash
alda repl --server
```

The output of this command tells you which port the Alda REPL server is
listening on:

```
nREPL server started on port 34223 on host localhost - nrepl://localhost:34223
```

The REPL server process also writes the port number into a file in the current
directory:

```
$ cat .alda-nrepl-port
34223
```

With this port number in hand, you can use the Alda CLI to send messages to the
Alda REPL server:

```
$ alda repl --client --port 34223 --message '{"op": "eval-and-play", "code": "harmonica: c8 d e f"}'
{"id":"17aa14c2-fa23-4bde-af4d-85839a85fc5d","session":"2d9ca3d1-2bcb-4e8b-9344-957ee4e4bdd9","status":["done"]}

$ alda repl --client --port 34223 --message '{"op": "eval-and-play", "code": "g f e d c2"}'
{"id":"27612a45-7d4a-4f7e-b65d-be6cf5107abf","session":"60f70325-efd6-492c-9e8f-279eba76a4d7","status":["done"]}

$ alda repl --client --port 34223 --message '{"op": "score-text"}'
{"id":"f1f45d28-d7f3-4fc2-b605-0fbc2b8d4ae1","session":"3da40081-1007-4630-bb38-b19abeaeef0a","status":["done"],"text":"harmonica: c8 d e f\ng f e d c2\n"}
```

> For a comprehensive list of what other operations are available and
> information about the parameters that they take, see the [Alda REPL server
> API][alda-repl-server-api] documentation.

The library that we built in the previous steps already has an `alda` function
that can shell out to the Alda CLI. We can build a few more functions on top of
that to achieve something analogous to the REPL workflow above:

```javascript
const fs = require('fs');

let replPort = null;

// connect(12345);  <- connect to the Alda server on port 12345
// connect();       <- read the port number from the .alda-nrepl-port file
function connect(port) {
  if (port) {
    replPort = port;
  } else {
    replPort = fs.readFileSync(".alda-nrepl-port", "utf8");
  }
}

function disconnect() {
  replPort = null;
}

function sendReplMessage(message) {
  let output = alda(
    "repl",
    "--client",
    "--port", replPort,
    "--message", JSON.stringify(message)
  );

  return JSON.parse(output);
}

function play(...objects) {
  let aldaCode = objects.reduce((code, object) => {
    return code + stringify(object) + " ";
  }, "");

  return sendReplMessage({"op": "eval-and-play", "code": aldaCode});
}
```

Now, you should be able to evaluate the following `play` expressions one at a
time and hear the notes played back to back, as part of the same score:

```javascript
play(
  part("harmonica"),
  octave(5),
  note(pitch("c"), noteLength(8)),
  note(pitch("d")),
  note(pitch("e")),
  note(pitch("f")),
  note(pitch("g")),
  note(pitch("a"))
);

play(
  note(pitch("b")),
  octaveUp(),
  note(pitch("c"))
);
```

## That's it!

If you've gotten this far, I hope you were able to lay the groundwork for a
super fun new Alda library for `<insert programming language here>`!

What did you think of this article? Was it helpful? Is it missing something?
Come chat with us in the [Alda Slack group][alda-slack] and let us know what you
think. We're happy to help!

Oh, and if you'd like to share what you've made, please consider adding it to
the list of libraries at the bottom of the [writing music
programmatically][writing-music-programmatically] article!

[writing-music-programmatically]: writing-music-programmatically.md
[alda-clj]: https://github.com/daveyarwood/alda-clj
[child-process]: https://nodejs.org/api/child_process.html
[alda-repl-server-api]: alda-repl-server-api.adoc
[alda-slack]: http://slack.alda.io/

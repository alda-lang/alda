# Alda 2 migration guide

Alda 2.0.0 was released in June 2021. Whereas Alda 1 had been written mostly in
Clojure (with a thin Java client for faster command line interactions), Alda 2
is a from-scratch rewrite in Go and Kotlin.

> If you're curious why Dave decided to rewrite Alda in Go and Kotlin, have a
> read through [this blog post][why-the-rewrite] that he wrote about it!

Alda 2 is mostly backwards compatible with Alda 1, to the extent that most of
the scores that you may have written with Alda 1 should work with Alda 2 and
sound exactly the same. The implementation of Alda has been rewritten from the
ground up, but Alda the language remains almost identical.

There is one important change to the language in Alda 2: **inline Clojure code
is no longer supported**. This is for obvious reasons: The Alda client is now
written in Go, so we can't evaluate arbitrary Clojure code inside an Alda score,
the way that we used to. (Despite this, Alda remains as powerful as ever as a
tool for algorithmic composition! See "Programmatic composition" below.)

Read on for a run-down of some things that you should be aware of when you
upgrade from Alda 1 to Alda 2.

## No more running `alda up`! Introducing `alda-player`

Alda 1 required you to start an Alda server (by running `alda up`) before you
could play a score.

In Alda 2, there is no Alda server. You can simply run a command like `alda play
-c "flute: o5 c8 < b16 a g f e d c2"`, without needing to run `alda up` first.

There is a new background process called `alda-player` that handles the audio
playback. Alda will start one for you automatically each time you play a score.
You'll need to have both `alda` and `alda-player` installed and available on
your PATH in order for this to work.

The Alda CLI will help make sure that you have the same version of `alda` and
`alda-player` installed, and it will even offer to install the correct version
of `alda-player` for you if they happen to become out of sync.

When you run `alda update`, both `alda` and `alda-player` are updated to the
latest version.

## Better troubleshooting with `alda doctor`

`alda doctor` is a new command that runs some basic health checks and looks for
signs that your system might not be set up for Alda to work properly. If all
goes well, you should see output like this:

```
OK  Parse source code
OK  Generate score model
OK  Find an open port
OK  Send and receive OSC messages
OK  Locate alda-player executable on PATH
OK  Check alda-player version
OK  Spawn a player process
OK  Ping player process
OK  Play score
OK  Export score as MIDI
OK  Locate player logs
OK  Player logs show the ping was received
OK  Shut down player process
OK  Spawn a player on an unknown port
OK  Discover the player
OK  Ping the player
OK  Shut the player down
OK  Start a REPL server
nREPL server started on port 36099 on host localhost - nrepl://localhost:36099
OK  Interact with the REPL server
```

If you run into any unexpected problems, the output of `alda doctor` can help
you pinpoint the issue and help Alda's maintainers find and fix bugs.

## New and improved `alda repl`

The REPL (**R**ead-**E**val-**P**lay **L**oop, a variation on the
"read-eval-print loop" from Lisp tradition) experience that you know and love
from Alda 1 is preserved in Alda 2. Simply run `alda repl` to begin an
interactive REPL session. Then you can play around and experiment with Alda code
and hear how each line of input sounds. (Try typing in something like
`midi-woodblock: c8. c c8 r c c` and see what happens.)

Just like before, you can enter `:help` for an overview of what REPL commands
are available, and find out more about a command by entering e.g. `:help play`.

So, what's new in the Alda 2 REPL? One superpower that we've given the new REPL
is that it can be run in either client or server mode. By default, `alda repl`
will start both a server and a client session. But if you already have a REPL
server running (or if a friend does, somewhere else in the world... :bulb:), you
can connect to it by running `alda repl --client --host example.com --port
12345` (or the shorter version: `alda repl -c -H example.com -p 12345`). This
has the potential to be a whole lot of fun, because multiple score-writers can
connect to the same REPL server and collaborate in real time!

> If you're interested in the technical details behind Alda's new super-REPL,
> check out Dave's blog post about it, [Alda and the nREPL
> protocol][alda-nrepl].

Some other REPL-related things that have changed since Alda 1:

* Server/worker management commands are no longer present because there are no
  longer server and worker processes to be managed! The following commands have
  been removed:
  * `:down`
  * `:downup`
  * `:list`
  * `:status`
  * `:up`

* The commands related to printing information about the score have been renamed
  so that things are a little more organized:
  * (v1) `:score` => (v2) `:score text` or just `:score`
  * (v1) `:info` => (v2) `:score info`
  * (v1) `:map` => (v2) `:score data`
  * (available in v2 only) `:score events`

## Attribute syntax has changed, in some cases

You might not have realized this, but in Alda 1, attributes like `(volume 42)`
were actually Clojure function calls that were evaluated at runtime. In fact,
the entire Clojure language was available to use within an Alda score. You
could, for example, generate a random number between 0 and 100 and set the
volume to that value with `(volume (rand-int 100))`.

You can't do this kind of thing anymore in Alda 2, because Alda is no longer
written in Clojure. (However, if you're interested in doing this kind of thing,
you'll be relieved to know that you still can! see "Programmatic composition"
below.)

Clojure is a [Lisp][lisp] programming language. If you don't know what that is,
here is a simple explanation: Lisp languages have a syntax that mostly consists
of parentheses. An "S-expression" is a list of elements inside of parentheses,
`(like this list)`. The first item in the list is an _operator_, and the
remaining items are _arguments_. S-expressions are nestable; for example, an
arithmetic expression like `(1 + 2) * (3 + 4)` is written in Lisp like this: `(*
(+ 1 2) (+ 3 4))`.

Alda 2 includes a simple, built-in Lisp language ("alda-lisp") that provides
just enough to support Alda's attribute operations. However, it lacks a lot of
the syntax of Clojure. Clojure has a variety of additional syntax that you may
have seen in Alda scores, including `:keywords`, `[vectors]` and `{hash maps}`.
alda-lisp does not have these features, so some Alda scores will not be playable
in Alda 2 if they make use of these features of Clojure.

The following attributes are affected by syntax changes in Alda 2:

<table>
  <thead>
    <tr>
      <th>Example</th>
      <th>Alda 1</th>
      <th>Alda 2</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Key signature</td>
      <td>
        <pre><code>(key-sig! "f+ c+ g+")
(key-sig! [:a :major])
(key-sig! {:f [:sharp] :c [:sharp] :g [:sharp]})</code></pre>
      </td>
      <td>
        <pre><code>(key-sig! "f+ c+ g+")
(key-sig! '(a major))
(key-sig! '(f (sharp) c (sharp) g (sharp)))</code></pre>
      </td>
    </tr>
    <tr>
      <td>Octave up/down</td>
      <td>
        <pre><code>>
<
(octave :up)
(octave :down)</code></pre>
      </td>
      <td>
        <pre><code>>
<
(octave 'up)
(octave 'down)</code></pre>
      </td>
    </tr>
  </tbody>
</table>

All other attributes should work just fine, but please [let us
know][open-an-issue] if you run into any other backwards compatibility issues
with your existing Alda 1 scores!

## Score starting volumes 

Alda 1 started all scores at an Alda volume of 100 corresponding to a MIDI 
velocity of 127. This is the maximum value. With Alda 2, you can now specify 
volumes with dynamic markings such as `(mp)` or `(ff)`. 

With this addition, all scores now default to a dynamic volume of `(mf)` 
equivalent to `(vol 54)`. This means that if you previously relied on scores 
starting from a volume of 100 in Alda 1, you now have to specify this attribute 
at the beginning of your Alda 2 score. 

## Programmatic composition

As noted above, inline Clojure code is no longer supported as a feature of Alda.

However, if you're interested in using Clojure to write algorithmic music,
you're in luck! In 2018, Dave created [alda-clj], a Clojure library for
live-coding music with Alda. The library provides a Clojure DSL for writing Alda
scores, and the DSL is equivalent to the one that was available in Alda 1.

Here is an [example score][entropy] that shows how a Clojure programmer can use
alda-clj to compose algorithmic music.

## `alda parse` output

The `alda parse` command parses an Alda score and produces JSON output that
represents the score data. This can be useful for debugging purposes, or for
building tooling on top of Alda.

The output of `alda parse` in Alda 2 is different from that of Alda 1 in a
number of ways. For example, here is the output of `alda parse -c "guitar: e" -o
events` in Alda 1:

```json
[
  {
    "event-type": "part",
    "instrument-call": {
      "names": [
        "guitar"
      ]
    },
    "events": null
  },
  {
    "event-type": "note",
    "letter": "e",
    "accidentals": [],
    "midi-note": null,
    "beats": null,
    "ms": null,
    "slur?": null
  }
]
```

And in Alda 2:

```json
[
  {
    "type": "part-declaration",
    "value": {
      "names": [
        "guitar"
      ]
    }
  },
  {
    "type": "note",
    "value": {
      "pitch": {
        "accidentals": [],
        "letter": "E"
      }
    }
  }
]
```

As you can see, Alda 1 and Alda 2 present the same information in very different
ways! If you happen to have built any tooling or workflows that rely on the Alda
1 `alda parse` output, you will likely need to make adjustments after upgrading
to Alda 2.

## That's it!

We hope you enjoy Alda 2! Please feel free to join our [Slack group][alda-slack]
and let us know what you think of it. You can also [open an
issue][open-an-issue] if you run into a bug or any other sort of weird behavior.
We'll be happy to help you get it sorted out!

[why-the-rewrite]: https://blog.djy.io/why-im-rewriting-alda-in-go-and-kotlin/
[lisp]: https://en.wikipedia.org/wiki/Lisp_(programming_language)
[alda-clj]: https://github.com/daveyarwood/alda-clj
[entropy]: https://github.com/daveyarwood/alda-clj/blob/master/examples/entropy
[alda-nrepl]: https://blog.djy.io/alda-and-the-nrepl-protocol/
[open-an-issue]: https://github.com/alda-lang/alda/issues/new/choose
[alda-slack]: http://slack.alda.io

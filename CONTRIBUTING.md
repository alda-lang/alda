# Contributing to Alda

The Alda project is composed of a number of subprojects, each of which has its
own GitHub repo:

  - [**alda**](https://github.com/alda-lang/alda) is the "main" repository.

    It includes a `build.boot` file that we use to pull together all of the
    subprojects and build the `alda` (or `alda.exe`, for Windows users)
    command-line executable.

    This repo also includes the Alda documentation, installation instructions,
    and downloadable releases. It serves as a landing page for newcomers to
    Alda.

  - [**alda-client-java**](https://github.com/alda-lang/alda-client-java) is the
    Alda command-line client, written in Java.

  - [**alda-server-clj**](https://github.com/alda-lang/alda-server-clj) is the
    Alda server process, which runs in the background and handles commands from
    the client. The server is implemented in Clojure.

  - [**alda-core**](https://github.com/alda-lang/alda-core) is the core
    implementation of Alda in the form of a Clojure library.

    This library includes the code that parses and compiles an Alda score into a
    data format that is "ready to play" by the server.

  - [**alda-sound-engine-clj**](https://github.com/alda-lang/alda-sound-engine-clj)
    is the part that interprets and plays the fully-realized score.

  - [**alda-repl-clj**](https://github.com/alda-lang/alda-repl-clj) is the Alda
    **R**ead-**E**val-**P**lay **L**oop. It is currently implemented in Clojure,
    but will eventually be [rewritten as part of the Java
    client](https://github.com/alda-lang/alda/issues/154).

  - [**alda.io**](https://github.com/alda-lang/alda.io) is the source code for
    the [official Alda website](http://alda.io).

    The website is currently under construction. We could use your help to make
    it look awesome and make sure it has all the information it needs!

Pull requests to any of these repos are warmly welcomed. Please feel free to
take on any open issue that interests you.

For a syrupy visual of what we have on our plate, check out our [waffle.io
board](https://waffle.io/alda-lang/alda).

## Instructions

- Fork the repository and make changes on your fork.
- See the README of each Alda subproject repo for details useful for developing
  that component.  You will find information like how to run and test that
  component locally and how to run the unit tests.
- Test your changes and make sure everything is working. Please add to the unit
  tests whenever it is appropriate.
- Submit a Pull Request.
- Your Pull Request should get the Dave Yarwood Seal of Approval™ before being
  merged. (Don't worry, he's not hard to win over.)

If you're confused about how any aspect of the code works (Clojure questions,
"what does this piece of code do," "can you walk me through how this works,"
etc.), don't hesitate to ask questions on the issue you're working on, or pop
into the [Alda Slack group](http://slack.alda.io) -- we'll be more than happy to
help!

## Building the `alda` (or `alda.exe`) executable

Dave is responsible for deploying subprojects to
[Clojars](https://clojars.org/groups/alda) and building and releasing the latest
builds of `alda` and `alda.exe`.

However, if you'd like to build a custom `alda` executable yourself (e.g. for
funsies or as an experiment), the process is documented
[here](doc/building-the-alda-executable.md).


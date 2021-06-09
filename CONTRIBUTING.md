# Contributing to Alda

_NOTE: As of December 31, 2020, Alda has been reformatted to encompass a single 
repository: "alda". As such, much of the below information is no longer applicable.
The previously-incorporated repositories ("alda-client-java", "alda-server-clj",
"alda-core", and "alda-sound-engine-clj") have been archived and are no longer
available to contribute to. Alda v2 will be released in the near future to reflect
these changes._

_You can read the official post about this topic [**here**](https://github.com/alda-lang/alda/issues/293#issuecomment-753245126)._

##
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

  - [**alda.io**](https://github.com/alda-lang/alda.io) is the source code for
    the [official Alda website](http://alda.io).

    The website is currently under construction. We could use your help to make
    it look awesome and make sure it has all the information it needs!

Pull requests to any of these repos are warmly welcomed. Please feel free to
take on any open issue that interests you.

For a top-level overview of things we're talking about and working on across all
of these repos, check out the [Alda GitHub Project board][gh-project].

[gh-project]: https://github.com/orgs/alda-lang/projects/1

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
into the #development channel in the [Alda Slack group](http://slack.alda.io) --
we'll be more than happy to help!

## Building the `alda` (or `alda.exe`) executable

Dave is responsible for deploying subprojects to
[Clojars](https://clojars.org/groups/alda) and building and releasing the latest
builds of `alda` and `alda.exe`.

However, if you'd like to build a custom `alda` executable yourself (e.g. for
funsies or as an experiment), the process is documented
[here](doc/building-the-alda-executable.md).


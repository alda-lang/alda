# Contributing to Alda

We're working on the following projects:

- The Alda [**client**](#client), written in Go.
- The Alda [**player**](#player), written in Kotlin.
- The official Alda website, [**alda.io**](https://alda.io).

The source code for the Alda client and player lives in this repo.

The website has [its own repo][alda-site-repo]. It's currently under
construction and we could use your help making it look awesome and making sure
it has all the information it needs!

Pull requests to contribute to any of these projects are warmly welcomed. Please
feel free to take on any open issue that interests you, and let us know if you
need any help!

For a top-level overview of things we're talking about and working on, check out
the [Alda GitHub Project board][gh-project].

[gh-project]: https://github.com/orgs/alda-lang/projects/1
[alda-site-repo]: https://github.com/alda-lang/alda.io

## General Instructions

- Fork the repository and make changes on your fork.
- Test your changes and make sure everything is working. Please add to the unit
  tests whenever it is appropriate.
- Submit a Pull Request.
- Your Pull Request should get the Dave Yarwood Seal of Approval™ before being
  merged. (Don't worry, he's not hard to win over!)

If you need help understanding how the code works or you have any other
questions, stop by the `#development` channel in the [Alda Slack
group](http://slack.alda.io) and say hi. We'll be happy to help!

## Development Guide

### Client

The Alda client is a Go program that parses input (Alda code), turns it into a
data representation of a musical score, and sends instructions to an Alda player
process to perform the score.

For more info, see the [README](client) in the `client/` folder of this repo.

### Player

Alda is designed so that playback is asynchronous. When you use the Alda client
to play a score, the client sends a bunch of [OSC][osc-intro] messages to a
player process running in the background.

The player process is agnostic of the Alda language. It simply receives and
handles OSC messages containing lower-level instructions pertaining to audio
playback.

The player supports live coding in that it allows one to define, modify, and
loop patterns during playback.

For more info, see the [README](player) in the `player/` folder of this repo.

[osc-intro]: http://opensoundcontrol.org/introduction-osc


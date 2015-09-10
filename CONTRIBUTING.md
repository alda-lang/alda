# Contributing to Alda

Pull requests are warmly welcomed. Please feel free to take on whatever [issue](https://github.com/alda-lang/alda/issues) interests you. 

## Instructions

- Fork this repository and make changes on your fork.
- Submit a Pull Request.
- Your Pull Request should get the Dave Yarwood Seal of Approval™ before being merged. (Don't worry, he's not hard to win over.)

If you're confused about how some aspect of the code works (Clojure questions, "what does this piece of code do," etc.), don't hesitate to ask questions on the issue you're working on -- we'll be more than happy to help.

## Development Guide

*TODO: more information here about how the codebase is laid out -- we want to make it easy for new contributors to jump right in*

### Testing changes

There are a couple of [Boot](http://boot-clj.com) tasks provided to help test changes.

#### `boot test`

You should run `boot test` prior to submitting a Pull Request. This will run automated tests that live in the `test` directory.

##### Adding tests

It is a good idea in general to add to the existing tests wherever it makes sense, i.e. if there is a new test case that Alda needs to consider. [Test-driven development](https://en.wikipedia.org/wiki/Test-driven_development) is a good idea.

If you find yourself adding a new file to the tests, be sure to add its namespace to the `test` task option in `build.boot` so that it will be included when you run the tests via `boot test`.

#### `boot alda`

When you run the `alda` executable, it uses the most recent *released* version of Alda. So, if you make any changes locally, they will not be included when you run `alda repl`, `alda play`, etc.

For testing local changes, you can use the `boot alda` task, which uses the current state of the repository, including any local changes you have made.

##### Example usage

    boot alda -x repl

    boot alda -x "play --code 'piano: c d e f g'"

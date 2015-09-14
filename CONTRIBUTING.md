# Contributing to Alda

Pull requests are warmly welcomed. Please feel free to take on whatever [issue](https://github.com/alda-lang/alda/issues) interests you. 

## Instructions

- Fork this repository and make changes on your fork.
- Submit a Pull Request.
- Your Pull Request should get the Dave Yarwood Seal of Approval™ before being merged. (Don't worry, he's not hard to win over.)

If you're confused about how some aspect of the code works (Clojure questions, "what does this piece of code do," etc.), don't hesitate to ask questions on the issue you're working on -- we'll be more than happy to help.

## Development Guide

* [alda.parser](#parsing-phase)
* [alda.lisp](#aldalisp)
* [alda.now](#aldanow)
* [alda.repl](#aldarepl)

Alda is a program that takes a string of code written in Alda syntax, parses it into executable Clojure code that will create a score, and then plays the score.

### Parsing phase

Parsing begins with the `parse-input` function in the [`alda.parser`](https://github.com/alda-lang/alda/blob/master/src/alda/parser.clj) namespace. This function uses a parser built using [Instaparse](https://github.com/Engelberg/instaparse), an excellent parser-generator library for Clojure. The grammar for Alda is [a single file written in BNF](https://github.com/alda-lang/alda/blob/master/grammar/alda.bnf) (with some Instaparse-specific sugar); if you find
yourself editing this file, it may be helpful to read up on Instaparse. [The tutorial in the README](https://github.com/Engelberg/instaparse) is comprehensive and excellent.

Code is given to the parser, resulting in a parse tree:

```
alda.parser=> (alda-parser "piano: c8 e g c1/f/a")

[:score 
  [:part 
    [:calls [:name "piano"]] 
    [:note 
      [:pitch "c"] 
      [:duration 
        [:note-length [:number "8"]]]] 
    [:note 
      [:pitch "e"]] 
    [:note 
      [:pitch "g"]] 
    [:chord 
      [:note 
        [:pitch "c"] 
        [:duration [:note-length [:number "1"]]]] 
      [:note 
        [:pitch "f"]] 
      [:note 
        [:pitch "a"]]]]]
```

The parse tree is then [transformed](https://github.com/Engelberg/instaparse#transforming-the-tree) into Clojure code which, when run, will produce a data representation of a musical score.

Clojure is a Lisp; in Lisp, code is data and data is code. This powerful concept allows us to represent a morsel of code as a list of elements. The first element in the list is a function, and every subsequent element is an argument to that function. These code morsels can even be nested, just like our parse tree. Alda's parser's transformation phase translates each type of node in the parse tree into a Clojure expression that can be evaluated with the help of the `alda.lisp` namespace.

```
alda.parser=> (parse-input "piano: c8 e g c1/f/a")

(alda.lisp/score 
  (alda.lisp/part {:names ["piano"]} 
    (alda.lisp/note 
      (alda.lisp/pitch :c) 
      (alda.lisp/duration (alda.lisp/note-length 8))) 
    (alda.lisp/note 
      (alda.lisp/pitch :e)) 
    (alda.lisp/note 
      (alda.lisp/pitch :g)) 
    (alda.lisp/chord 
      (alda.lisp/note 
        (alda.lisp/pitch :c) 
        (alda.lisp/duration (alda.lisp/note-length 1))) 
      (alda.lisp/note 
        (alda.lisp/pitch :f)) 
      (alda.lisp/note 
        (alda.lisp/pitch :a)))))
```

### alda.lisp

TODO

### alda.now

TODO

### alda.repl

TODO

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

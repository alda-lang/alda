# Installation

## Mac OS X / Linux

The executable file `alda` in the `bin` directory of this repository is a standalone executable script that can be run from anywhere. It will retrieve the latest release version of Alda and run it, passing along any command-line arguments you give it.

* To install Alda, simply copy the `alda` script from this repo into any directory in your `$PATH`, e.g. `/bin` or `/usr/local/bin`:

        curl https://raw.githubusercontent.com/alda-lang/alda/master/bin/alda -o /usr/local/bin/alda && chmod +x /usr/local/bin/alda

* This script requires the Clojure build tool [Boot](http://www.boot-clj.com), so you will need to have that installed as well. Mac OS X users with [Homebrew](https://github.com/homebrew/homebrew) can run `brew install boot-clj` to install Boot. Otherwise, [see here](https://github.com/boot-clj/boot#install) for more details about installing Boot.

Once you've completed the steps above, you'll be able to run `alda` from any working directory. Running the command `alda` by itself will display the help text.

## Windows

The `alda` script doesn't seem to work for Windows users. If you're a Windows power user, [please feel free to weigh in on this issue](https://github.com/alda-lang/alda/issues/48). Until we have that sorted out, there is a workaround:

1. Install [Boot](https://github.com/boot-clj/boot#install).
2. Clone this repo and `cd` into it.
3. You can now run `boot alda -x "<cmd> <args>"` while you are in this directory.

Examples:

* `boot alda -x repl` to start the Alda REPL
* `boot alda -x "play --code 'piano: c d e f g'"`

Caveats:

* It's more typing.
* It only works if you're in the Alda repo folder.
* Unlike the `alda` script, running the `boot alda` task will not automatically update Alda; you will have to do so manually by running `git pull`.
* If the command you're running is longer than one word, you must wrap it in double quotes -- see the examples above.


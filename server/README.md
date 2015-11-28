# alda-server

The core machinery of Alda, implemented in Clojure.

## Components

* **alda.parser** (reads Alda code and transforms it into Clojure code in the context of the `alda.lisp` namespace)

* **alda.lisp** (a Clojure DSL which provides the context for evaluating an Alda score, in its Clojure code form)

* **alda.sound** (generates sound based on the data map produced by parsing and evaluating Alda code)

* **alda.now** (an entrypoint for using Alda as a Clojure library)

* **alda.repl** (an interactive **R**ead-**E**val-**P**lay **L**oop for Alda code)

## Usage

The standard way to use Alda is via the client.

However, the server can also be used directly, if desired. The downside of this (and the reason why the client exists) is that it can take several seconds or more to run tasks from the command line due to Clojure/JVM startup time. For information about available tasks, run `bin/alda help`.

## Building/Running

### Future

We are currently revising the build procedure for the server part of Alda -- in the near future, there will be a single Boot task that will build standalone `alda-server` executables for both Unix/Linux and Windows.

### Present

#### `bin/alda` script

As we're hammering out the kinks, the current "official" way to build the latest version of Alda for development purposes is to install it to your local Maven repository and run the `alda` script included in the `bin/` directory of this repo:

* Install [Boot](http://boot-clj.com).
* For the latest release version of Alda, simply run `bin/alda`. This will pull down the latest release of Alda from Clojars and install it to your local Maven repo.
* To run Alda in its current state on your computer (e.g. to test changes), install it to your local Maven repo by running `boot pom jar install`. Then run `bin/alda`, which will use your local version.

##### Example usage

```
bin/alda help
bin/alda repl
bin/alda play --file ../examples/hello_world.alda
```

#### `boot alda` task

As a convenience, you can also run Alda in its current state (without installing it to your local Maven repo) via the `boot alda` task.

##### Example usage

```
boot alda -x help
boot alda -x repl
boot alda -x "play --file ../examples/hello_world.alda"
```


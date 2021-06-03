# Building the `alda` (and/or `alda.exe`) Executable

Dave does this process himself for every release, but it doesn't hurt to have it documented here!

Knowledge is Power. :bulb:

## What You'll Need

* The Alda project uses [Boot](https://boot-clj.github.io/) to build releases from the Java/Clojure source, as well as to perform useful development tasks like running tests. You will need Boot in order to test any changes you make to the source code.

  > The `boot` commands described in this guide need to be run while in the root directory of this project, which contains the project file `build.boot`.

* [Launch4j](http://launch4j.sourceforge.net) is needed in order to build the Windows executable.

  If you're on a Mac with [Homebrew](http://brew.sh) installed, you can install Launch4j by running:

      brew install launch4j

  Java Development Kit (JDK) 8 is needed.

## The `boot build` task

All of the Alda subprojects are packaged together into the same uberjar. You can build the project by running:

    boot build -o /path/to/output-dir/

This will build the `alda` and `alda.exe` executables and place them in the output directory of your choice.

Note that this will use the versions of the Alda subprojects specified in the `build.boot` in this repository. For release builds, the process is to update and deploy new versions of the changed subproject(s) to [Clojars](https://clojars.org/groups/alda), then update the subproject dependencies (at the top of the `build.boot` in this repo) to the new versions and run `boot build` to build the executables.

If you are creating an executable that uses custom subproject code, you will need to first build the subproject(s) you changed by running `boot package install` in the appropriate subproject repo folder. This will install the subproject package into your local Maven repository on your computer, and this local version will be used instead of the version deployed to Clojars when the executables are built.


# alda (client)

A command-line client for Alda, written in Go.

## Building

Go is a quirky language. It looks like there are roughly 4.2 billion ways to organize and build Go projects, many of which rely on fumbling with the `$GOPATH` environment variable. `gb` seems to be the most sensible way to do project-based builds.

To build the latest version of the `alda` client for development purposes:

1. Install [`gb`](https://getgb.io/).

2. Fetch dependencies by running `gb vendor restore`.

3. Build the project by running `gb build alda`. This builds an executable for your platform, located at `bin/alda`.

## Usage

For information about available tasks, run `alda help`.

# alda.server

`alda.server` is the Clojure namespace containing the code for the background server process started by the Alda command-line [client](alda-client.md).

The command `alda up` starts a server in the background, whereas `alda server` starts a server in the foreground, which can be useful for development/debugging purposes.

Note that the actual "work" of parsing, evaluating and playing Alda scores is done by separate [worker](alda-worker.md) processes. As a user, you do not need to worry about these worker processes, as the server manages them for you in the background. When you run a command using the Alda command line client (such as `alda play -f some-file.alda`), the client makes a request to the server, and the server delegates the work to the first available worker process.

## See Also

* [ZeroMQ Architecture](zeromq-architecture.md)

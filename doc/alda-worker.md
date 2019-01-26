# alda.worker

`alda.worker` is the Clojure namespace containing the code run by each
background Alda worker process. These processes are maintained by the [Alda
server](alda-server.md) in such a way that the end user does not necessarily
need to know about them. The client communicates with the server and the server
utilizes its workers to respond to the client's requests.

Worker processes are spawned in the background when a server starts.

To start a worker in the foreground for development/debugging purposes, you can
run `alda -p 12345 worker`, where `12345` is the port number of the backend port
used by the server to communicate with its workers.

Each worker process has its own Alda environment, including its own MIDI
Synthesizer instance and JSyn SynthesisEngine instance.


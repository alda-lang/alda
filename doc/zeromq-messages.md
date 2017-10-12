# ZeroMQ Messages

## Sent by Client

### `parse`

#### Description

Asks the server to parse a string of Alda code.

The result returned by the server is, by default, a JSON object representing the score data.

If the request includes `"options": {"output": "events"}`, the result returned
is a JSON array of objects representing Alda events.

#### Example

```json
{
  "command": "parse",
  "body": "piano: c8 d e f g2",
  "options": {
    "output": "events"
  }
}
```

---

### `ping`

#### Description

Pings the server to see if it's up.

#### Example

```json
{"command": "ping"}
```

---

### `play`

#### Description

Asks the server to play a string of Alda code.

Options include `from` and `to` strings, representing minute/second markings or score marker names directing the server where in the score to start/end. (When omitted, the score will default to the beginning and end of the score.)

Another option is to supply a `history` string of Alda code, representing the score so far up until the current moment in time. The `body` then represents any new notes/events, starting now. This can be useful for implementing an alternate Alda client, for example to implement a text editor plugin. When `history` is supplied, the entire score (i.e. `history` + `body`) is parsed and evaluated for context, but only the `body` is played.

The `jobId` option is used internally to differentiate one score from another
and coordinate with the server that the client and server are both talking about
the same score.

#### Examples

```json
{
  "command": "play",
  "body": "piano: c8 d e f g2",
  "options": {
    "from": "chorus",
    "to": "5:55",
    "jobId": "ba2a6924-45aa-11e7-a919-92ebcb67fe33"
  }
}
```

```json
{
  "command": "play",
  "body": "g a b > c",
  "options": {
    "history": "piano: (tempo 200) c8 d e f",
    "jobId": "c2d43f32-45aa-11e7-a919-92ebcb67fe33"
  }
}
```

---

### `play-status`

#### Description

Asks a worker for its current status, e.g. parsing a score, playing a score, done.

The `jobId` option is used internally to differentiate one score from another
and coordinate with the server that the client and server are both talking about
the same score. The client should use the same `jobId` that it supplied when it
submitted the `play` request for the score.

The `body` of the response is a string like "parsing," "available," etc.

There is also a `pending` key in the response, which is `true` or `false` to
indicate whether there is still more work to be done. As playback is
asynchronous, the playback process is effectively "done" (meaning the client can
stop querying for status) when playback has started, even if playback hasn't
finished.

In addition, the response may contain a `score` key, whose value is a JSON
object representing the score data resulting from parsing the Alda code.

> Note that this message is sent not to the server, but to a specific worker. To do this, you add an extra frame with the worker's address. For more details, see [ZeroMQ Architecture#Message Structure](zeromq-architecture.md#message-structure).

#### Example

```json
{
  "command": "play-status",
  "options": {
    "jobId": "c2d43f32-45aa-11e7-a919-92ebcb67fe33"
  }
}
```

---

### `stop-server`

#### Description

Tells the server to shut down.

#### Example

```json
{"command": "stop-server"}
```

---

### `status`

#### Description

Asks the server for its current status.

The response from the server is a string like `Server up (2/2 workers available)`.

#### Example

```json
{"command": "status"}
```

---

### `version`

#### Description

Asks the server for its Alda version number.

#### Example

```json
{"command": "version"}
```

## Sent by Server

> The server sends 3 kinds of messages:
>
> - Messages forwarded to/from clients and workers.
> - One-frame worker control messages like `KILL` and `HEARTBEAT`.
> - Direct JSON responses to client requests that don't need to be handled by workers.
>
> Only the last type is covered here.

### Example Response

```json
{"success": true, "body": "1.0.0-rc42", "noWorker": true}
```

## Sent by Worker

### Example Response

```json
{"success": true, "pending": false, "body": "parsing"}
```

## See Also

* [ZeroMQ Architecture](zeromq-architecture.md)

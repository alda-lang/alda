# Alda REPL

Alda comes with an interactive REPL (**R**ead-**E**val-**P**lay **L**oop) that
you can use to play around with its syntax. After each line of code that you
enter into the REPL prompt, you will hear the result.

To start the Alda REPL, run:

```bash
alda repl
```

At the REPL prompt, you can either enter Alda code, or use one of the available
built-in commands, which start with a colon.

For a list of available commands, enter `:help`.

## REPL clients and servers

When you run `alda repl`, it is actually starting both a REPL server and a REPL
client session. The client sends your input to the server (which happens to be
running inside the same process, in this case) and prints output based on the
server's response.

You can also start an Alda REPL client or server independently:

```bash
# Start server only, running on a random available port.
alda repl --server

# Start server only, running on port 12345.
alda repl --server --port 12345

# Start client only, communicating with the server running on port 12345.
alda repl --client --port 12345
```

You can even run the REPL server on a different computer, and connect to it by
specifying the host or IP address:

```bash
# On another computer:
alda repl --server --port 12345

# Start a client, communicating with the Alda REPL server running on the other
# computer (assuming that the external IP address is 11.22.33.44, for the sake of
# example):
alda repl --client --host 11.22.33.44 --port 12345
```

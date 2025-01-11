# Troubleshooting

## General tips

### `alda doctor`

If you have any issues with Alda, you can run `alda doctor` to run a series of
quick health checks:

```
$ alda doctor
OK  Parse source code
OK  Generate score model
OK  Ensure that there are no stale player processes
OK  Find an open port
OK  Send and receive OSC messages
OK  Locate alda-player executable on PATH
OK  Check alda-player version
OK  Spawn a player process
OK  Ping player process
OK  Play score
OK  Export score as MIDI
OK  Locate player logs
OK  Player logs show the ping was received
OK  Shut down player process
OK  Spawn a player on an unknown port
OK  Discover the player
OK  Ping the player
OK  Shut the player down
OK  Start a REPL server
nREPL server started on port 39879 on host 127.0.0.1 - nrepl://127.0.0.1:39879
OK  Find the REPL server
OK  Interact with the REPL server
OK  Shut down the REPL server
```

This can help you narrow down where the specific issue may be.

### Player processes

Normally, the `alda` client operates by launching `alda-player` processes
automatically in the background. These processes are designed to be short-lived,
handling audio playback and eventually shutting themselves down after a period
of inactivity.

You can list information all of the player processes that are currently running
with `alda ps`:

```
$ alda ps
id      port    state   expiry  type    pid
aar     42175   ready   6 minutes from now      player  387371
hgf     40093   ready   7 minutes from now      player  387370
oau     44559   ready   8 minutes from now      player  387372
```

Each player process saves log messages into files. These log files can be useful
for troubleshooting if you're having issues with playback. To find the location
of these log files, run `alda-player info`:

```
$ alda-player info
alda-player 2.3.0
log path: /home/dave/.cache/alda/logs
```

Another thing you can try, if you're having issues, is to run a player process
in the foreground, with the `-vv` option for "very verbose" logging:

```
$ alda-player -vv run -p 27705
ykt INFO  2024-07-12 13:10:43 StateManager.cleanUpStaleStateFiles:88 - Cleaning up stale files in /home/dave/.cache/alda/state/players...
ykt INFO  2024-07-12 13:10:43 StateManager.cleanUpStaleStateFiles:88 - Cleaning up stale files in /home/dave/.cache/alda/state/repl-servers...
ykt INFO  2024-07-12 13:10:43 Main.run:77 - Starting receiver, listening on port 27705...
ykt INFO  2024-07-12 13:10:43 MidiEngine.info:242 - [0] Initializing MIDI sequencer...
ykt INFO  2024-07-12 13:10:43 MidiEngine.info:242 - [0] Initializing MIDI synthesizer...
ykt INFO  2024-07-12 13:10:47 MidiEngine.info:242 - [0] Player ready
```

Then, in a separate terminal, you can use the port number (in this case, we
chose port 27705 when we started the player process) to use that player
specifically in an `alda play` command:

```
$ alda play -p 27705 -c 'piano: c d e f g'
Playing...
```

You should see logs in the terminal where you're running the player process.

## IcedTea-related issues

[IcedTea] is a variant of OpenJDK that is distributed with Linux distributions
such as Fedora, Gentoo, and Debian. If your distribution includes IcedTea, you
might encounter an issue where playback does not work with Alda.

One user reported seeing this error in the player logs:

```
Exception in thread "main" java.lang.UnsatisfiedLinkError: no icedtea-sound in java.library.path
```

They were able to fix it by running `apt-cache search icedtea` to find IcedTea
related packages, and then running `sudo apt install icedtea-netx libpulse-java
libpulse-jni` to install the missing packages.

## Windows

### RegCreatKeyEx errors

When running `alda` in the terminal on Windows, you might run into the following error:

```
WARNING: Could not open/create prefs root node Software\JavaSoft\Prefs at root 0x80000002.
Windows RegCreateKeyEx(...) returned error code 5.
```

This error means that Windows is missing the `JavaSoft\Prefs` registry key, but this is easily solved: simple add one.

1. Run `regedit` to start the Windows registry editor
2. Expand `HKEY_LOCAL_MACHINE` and find the `Software` key,
3. Inside `Software`, find the `JavaSoft` key.

Right click on "JavaSoft", and select "new" -> "Key". Call this key `Prefs`, hit enter, and you're done.

[IcedTea]: https://openjdk.org/projects/icedtea/

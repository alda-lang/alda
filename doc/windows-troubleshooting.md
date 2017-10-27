# Troubleshooting on Windows

## RegCreatKeyEx errors

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

## Workers not starting

When running `alda repl` in the terminal on Windows, you might experience worker processes not starting:

```
E:\Workspace\Temp>alda repl
... 

          1.0.0-rc75
         repl session

Type :help for a list of available commands.

> piano: c d e f
[27713] ERROR No worker processes are ready yet. Please wait a minute.
>
```

Waiting might not solve the problem. As a workaround you can start a
worker process manually: First find out the backend port on the server:

```
> :status
[27713] Server up (0/2 workers available, backend port: 55521)
```

Here the backend port is 55521. Now open another terminal window and run the command

```
alda -p 55521 worker
```

Back in the REPL the server should be connected to the worker now:

```
> :status
[27713] Server up (1/2 workers available, backend port: 55521)
```

Good to go!

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

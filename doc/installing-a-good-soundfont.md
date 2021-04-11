# Installing a good soundfont

Default JVM soundfonts usually are of low quality. We recommend installing a good freeware soundfont like FluidR3 to make your MIDI instruments sound a lot nicer.

A variety of popular freeware soundfonts, including FluidR3, are available for
download [here](https://musescore.org/en/handbook/soundfonts#list).

## Mac / Linux

For your convenience, there is a script in this repo that will install the
FluidR3 soundfont for Mac and Linux users.

Grab a copy of the script:

```bash
curl \
  https://raw.githubusercontent.com/alda-lang/alda/master/scripts/install-fluidr3 \
  -o /tmp/install-fluidr3

chmod +x /tmp/install-fluidr3
```

Take an opportunity to review the script, if you'd like, before you run it:

```bash
/tmp/install-fluidr3
```

This will download FluidR3 and replace `~/.gervill/soundbank-emg.sf2` (your
JVM's default soundfont) with it.

## Windows

<img src="windows_jre_soundfont.png"
     alt="Replacing the JVM soundfont on Windows">

To replace the default soundfont on a Windows OS:

1. Locate your Java Runtime (JRE) folder and navigate into the `lib` folder.
   * If you have JDK 8 or earlier installed, locate your JDK folder instead and navigate into the `jre\lib` folder.
2. Make a new folder named `audio`.
3. Copy any `.sf2` file into this folder.

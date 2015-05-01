#### Feature Ideas

- alternate syntax for changing attributes over time (ex. cresc/decresc with time parameters, instead of having to change the volume manually in steps)
- instrument specific commands/modes (guitar, winds, etc.)
- drums / percussion
- waveform synthesis
- a "sampler" instrument
- harness the power of randomness

##### alda.repl

- support for inputting notes via a MIDI keyboard (or maybe even STDIN with keys mapped to MIDI notes, like a pseudo-MIDI-keyboard): the REPL would guess, based on the current tempo, what note lengths you meant. You could then edit the notes before submitting them. Alternatively (and more simply), it could just read in notes one at a time and print them, without note-lengths attached, and you could add in the durations as you go.

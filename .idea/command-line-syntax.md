Play a file:

    alda demo.alda

Save/export the audio output as a .wav file:

    alda demo.alda file.wav

Optionally choose only a certain part of your score to be played / exported, by using -s (start) and -e (end) markers. These markers can be minute/second marks or custom markers defined in the score. Whenever -s and -e are not specified at the command line, default to the very beginning and the very end of the score.

    alda -s 0:30 demo.alda         <-- plays the score starting at 30 seconds in
    alda -e 1:00 demo.alda         <-- plays the score from the beginning and stops at 1 minute in
    alda -s 0:30 -e 1:00 demo.alda <-- plays the score just from 0:30 - 1:00

    alda -s chorus demo.alda           <-- plays the score starting from custom marker "chorus" as defined in the score
    alda -e chorus demo.alda           <-- plays the score up until the "chorus" marker, then stops
    alda -s chorus -e bridge demo.alda <-- plays the score starting at the "chorus" marker and ending at the "bridge" marker

# List of Instruments

Currently, only General MIDI instruments are supported. In the future, we plan to add [waveform synthesis](https://github.com/alda-lang/alda/issues/100) so that you will be able to use sine/square/triangle/sawtooth waves as well as complex synthesizers built from waveforms.

Any of the instrument names below, as well as their aliases, can be used as instruments in an Alda score, e.g.:

```alda
midi-harpsichord: c8 d e f g a b > c
```

## MIDI Instruments

These directly correspond to the instruments in the [General MIDI sound set](http://www.midi.org/techspecs/gm1sound.php). They are grouped here by patch group according to the MIDI spec.

Aliases are in parentheses after the instrument's name.

> Note that some of these aliases may be replaced in the future with non-MIDI instruments, e.g. sampled or waveform instruments. To ensure that your scores will always use specifically MIDI instruments, you can use the `midi-` prefixed names.

### Piano

* midi-acoustic-grand-piano (midi-piano, piano)
* midi-bright-acoustic-piano
* midi-electric-grand-piano
* midi-honky-tonk-piano
* midi-electric-piano-1
* midi-electric-piano-2
* midi-harpsichord (harpsichord)
* midi-clavi (midi-clavinet, clavinet)

### Chromatic Percussion

* midi-celesta (celesta, celeste, midi-celeste)
* midi-glockenspiel (glockenspiel)
* midi-music-box (music-box)
* midi-vibraphone (vibraphone, vibes, midi-vibes)
* midi-marimba (marimba)
* midi-xylophone (xylophone)
* midi-tubular-bells (tubular-bells)
* midi-dulcimer (dulcimer)

### Organ

* midi-drawbar-organ
* midi-percussive-organ
* midi-rock-organ
* midi-church-organ (organ)
* midi-reed-organ
* midi-accordion (accordion)
* midi-harmonica (harmonica)
* midi-tango-accordion

### Guitar

* midi-acoustic-guitar-nylon (midi-acoustic-guitar, acoustic-guitar, guitar)
* midi-acoustic-guitar-steel
* midi-electric-guitar-jazz
* midi-electric-guitar-clean (electric-guitar-clean)
* midi-electric-guitar-palm-muted
* midi-electric-guitar-overdrive (electric-guitar-overdrive)
* midi-electric-guitar-distorted (electric-guitar-distorted)
* midi-electric-guitar-harmonics (electric-guitar-harmonics)

### Bass

* midi-acoustic-bass (acoustic-bass, upright-bass)
* midi-electric-bass-finger (electric-bass-finger, electric-bass)
* midi-electric-bass-pick (electric-bass-pick)
* midi-fretless-bass (fretless-bass)
* midi-bass-slap
* midi-bass-pop
* midi-synth-bass-1
* midi-synth-bass-2

### Strings (and Timpani, for some reason)

* midi-violin (violin)
* midi-viola (viola)
* midi-cello (cello)
* midi-contrabass (string-bass, arco-bass, double-bass, contrabass, midi-string-bass, midi-arco-bass, midi-double-bass)
* midi-tremolo-strings
* midi-pizzicato-strings
* midi-orchestral-harp (harp, orchestral-harp, midi-harp)
* midi-timpani (timpani)

### Ensemble

* midi-string-ensemble-1
* midi-string-ensemble-2
* midi-synth-strings-1
* midi-synth-strings-2
* midi-choir-aahs
* midi-voice-oohs
* midi-synth-voice
* midi-orchestra-hit

### Brass

* midi-trumpet (trumpet)
* midi-trombone (trombone
* midi-tuba (tuba)
* midi-muted-trumpet
* midi-french-horn (french-horn)
* midi-brass-section
* midi-synth-brass-1
* midi-synth-brass-2

### Reed
* midi-soprano-saxophone (midi-soprano-sax, soprano-saxophone, soprano-sax)
* midi-alto-saxophone (midi-alto-sax, alto-saxophone, alto-sax)
* midi-tenor-saxophone (midi-tenor-sax, tenor-saxophone, tenor-sax)
* midi-baritone-saxophone (midi-baritone-sax, midi-bari-sax, baritone-saxophone, baritone-sax, bari-sax)
* midi-oboe (oboe)
* midi-english-horn (english-horn)
* midi-bassoon (bassoon)
* midi-clarinet (clarinet)

### Pipe

* midi-piccolo (piccolo)
* midi-flute (flute)
* midi-recorder (recorder)
* midi-pan-flute (pan-flute)
* midi-bottle (bottle)
* midi-shakuhachi (shakuhachi)
* midi-whistle (whistle)
* midi-ocarina (ocarina)

### Synth Lead

* midi-square-lead (square, square-wave, square-lead, midi-square, midi-square-wave)
* midi-saw-wave (sawtooth, saw-wave, saw-lead, midi-sawtooth, midi-saw-lead)
* midi-calliope-lead (calliope-lead, calliope, midi-calliope)
* midi-chiffer-lead (chiffer-lead, chiffer, chiff, midi-chiffer, midi-chiff)
* midi-charang (charang)
* midi-solo-vox
* midi-fifths (midi-sawtooth-fifths)
* midi-bass-and-lead (midi-bass+lead)

### Synth Pad

* midi-synth-pad-new-age (midi-pad-new-age, midi-new-age-pad)
* midi-synth-pad-warm (midi-pad-warm, midi-warm-pad)
* midi-synth-pad-polysynth (midi-pad-polysynth, midi-polysynth-pad)
* midi-synth-pad-choir (midi-pad-choir, midi-choir-pad)
* midi-synth-pad-bowed (midi-pad-bowed, midi-bowed-pad, midi-pad-bowed-glass, midi-bowed-glass-pad)
* midi-synth-pad-metallic (midi-pad-metallic, midi-metallic-pad, midi-pad-metal, midi-metal-pad)
* midi-synth-pad-halo (midi-pad-halo, midi-halo-pad)
* midi-synth-pad-sweep (midi-pad-sweep, midi-sweep-pad)

### Synth Effects

* midi-fx-rain (midi-fx-ice-rain, midi-rain, midi-ice-rain)
* midi-fx-soundtrack (midi-soundtrack)
* midi-fx-crystal (midi-crystal)
* midi-fx-atmosphere (midi-atmosphere)
* midi-fx-brightness (midi-brightness)
* midi-fx-goblins (midi-fx-goblin, midi-goblins, midi-goblin)
* midi-fx-echoes (midi-fx-echo-drops, midi-echoes, midi-echo-drops)
* midi-fx-sci-fi (midi-sci-fi)

### Ethnic

* midi-sitar (sitar)
* midi-banjo (banjo)
* midi-shamisen (shamisen)
* midi-koto (koto)
* midi-kalimba (kalimba)
* midi-bagpipes (bagpipes)
* midi-fiddle
* midi-shehnai (shehnai, shahnai, shenai, shanai, midi-shahnai, midi-shenai, midi-shanai)

### Percussive

* midi-tinkle-bell (midi-tinker-bell)
* midi-agogo
* midi-steel-drums (midi-steel-drum, steel-drums, steel-drum)
* midi-woodblock
* midi-taiko-drum
* midi-melodic-tom
* midi-synth-drum
* midi-reverse-cymbal

### Sound Effects

* midi-guitar-fret-noise
* midi-breath-noise
* midi-seashore
* midi-bird-tweet
* midi-telephone-ring
* midi-helicopter
* midi-applause
* midi-gunshot (midi-gun-shot)

### Percussion

There is a special `midi-percussion` instrument (alias: `percussion`) which provides a variety of percussion sounds, each mapped to a different note. Each note corresponds to a unique percussive instrument, but the sound's pitch is not relative to the pitch of the note. (See [here](https://en.wikipedia.org/wiki/General_MIDI#Percussion) for more information about MIDI percussion.)

Drum set example:

```alda
(tempo! 150)

midi-percussion:
  V1: # bass and snare
    o2 c4 e8 c r c e c
  V2: # cymbals (hi-hat, crash, ride bell, another crash)
    o2 f+8 f+ r o3 c+8~8 f16 f r8 a
```


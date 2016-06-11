(ns alda.lisp.instruments.midi
  (:require [alda.lisp.model.instrument :refer (definstrument)]))

; NOTE: For the time being, some of these instruments have non-prefixed names
;       like "piano" and "trumpet" as aliases. I eventually want to re-map
;       those names to more realistic-sounding, sampled instruments instead of
;       MIDI. Once that happens, we'll need to remove the aliases.

; reference: http://www.jimmenard.com/midi_ref.html#General_MIDI

;; 1-8: PIANO ;;

(definstrument midi-acoustic-grand-piano
  :aliases ["midi-piano" "piano"]
  :config  {:type :midi
            :patch 1})

(definstrument midi-bright-acoustic-piano
  :aliases []
  :config  {:type :midi
            :patch 2})

(definstrument midi-electric-grand-piano
  :aliases []
  :config  {:type :midi
            :patch 3})

(definstrument midi-honky-tonk-piano
  :aliases []
  :config  {:type :midi
            :patch 4})

(definstrument midi-electric-piano-1
  :aliases []
  :config  {:type :midi
            :patch 5})

(definstrument midi-electric-piano-2
  :aliases []
  :config  {:type :midi
            :patch 6})

(definstrument midi-harpsichord
  :aliases ["harpsichord"]
  :config  {:type :midi
            :patch 7})

(definstrument midi-clavi
  :aliases ["midi-clavinet" "clavinet"]
  :config  {:type :midi
            :patch 8})

;; 9-16: CHROMATIC PERCUSSION ;;

(definstrument midi-celesta
  :aliases ["celesta" "celeste" "midi-celeste"]
  :config  {:type :midi
            :patch 9})

(definstrument midi-glockenspiel
  :aliases ["glockenspiel"]
  :config  {:type :midi
            :patch 10})

(definstrument midi-music-box
  :aliases ["music-box"]
  :config  {:type :midi
            :patch 11})

(definstrument midi-vibraphone
  :aliases ["vibraphone" "vibes" "midi-vibes"]
  :config  {:type :midi
            :patch 12})

(definstrument midi-marimba
  :aliases ["marimba"]
  :config  {:type :midi
            :patch 13})

(definstrument midi-xylophone
  :aliases ["xylophone"]
  :config  {:type :midi
            :patch 14})

(definstrument midi-tubular-bells
  :aliases ["tubular-bells"]
  :config  {:type :midi
            :patch 15})

(definstrument midi-dulcimer
  :aliases ["dulcimer"]
  :config  {:type :midi
            :patch 16})

;; 17-24: ORGAN ;;

(definstrument midi-drawbar-organ
  :aliases []
  :config  {:type :midi
            :patch 17})

(definstrument midi-percussive-organ
  :aliases []
  :config  {:type :midi
            :patch 18})

(definstrument midi-rock-organ
  :aliases []
  :config  {:type :midi
            :patch 19})

(definstrument midi-church-organ
  :aliases ["organ"]
  :config  {:type :midi
            :patch 20})

(definstrument midi-reed-organ
  :aliases []
  :config  {:type :midi
            :patch 21})

(definstrument midi-accordion
  :aliases ["accordion"]
  :config  {:type :midi
            :patch 22})

(definstrument midi-harmonica
  :aliases ["harmonica"]
  :config  {:type :midi
            :patch 23})

(definstrument midi-tango-accordion
  :aliases []
  :config  {:type :midi
            :patch 24})

;; 25-32: GUITAR ;;

(definstrument midi-acoustic-guitar-nylon
  :aliases ["midi-acoustic-guitar" "acoustic-guitar" "guitar"]
  :config  {:type :midi
            :patch 25})

(definstrument midi-acoustic-guitar-steel
  :aliases []
  :config  {:type :midi
            :patch 26})

(definstrument midi-electric-guitar-jazz
  :aliases []
  :config  {:type :midi
            :patch 27})

(definstrument midi-electric-guitar-clean
  :aliases ["electric-guitar-clean"]
  :config  {:type :midi
            :patch 28})

(definstrument midi-electric-guitar-palm-muted
  :aliases []
  :config  {:type :midi
            :patch 29})

(definstrument midi-electric-guitar-overdrive
  :aliases ["electric-guitar-overdrive"]
  :config  {:type :midi
            :patch 30})

(definstrument midi-electric-guitar-distorted
  :aliases ["electric-guitar-distorted"]
  :config  {:type :midi
            :patch 31})

(definstrument midi-electric-guitar-harmonics
  :aliases ["electric-guitar-harmonics"]
  :config  {:type :midi
            :patch 32})

;; 33-40: BASS ;;

(definstrument midi-acoustic-bass
  :aliases ["acoustic-bass" "upright-bass"]
  :config  {:type :midi
            :patch 33})

(definstrument midi-electric-bass-finger
  :aliases ["electric-bass-finger" "electric-bass"]
  :config  {:type :midi
            :patch 34})

(definstrument midi-electric-bass-pick
  :aliases ["electric-bass-pick"]
  :config  {:type :midi
            :patch 35})

(definstrument midi-fretless-bass
  :aliases ["fretless-bass"]
  :config  {:type :midi
            :patch 36})

(definstrument midi-bass-slap
  :aliases []
  :config  {:type :midi
            :patch 37})

(definstrument midi-bass-pop
  :aliases []
  :config  {:type :midi
            :patch 38})

(definstrument midi-synth-bass-1
  :aliases []
  :config  {:type :midi
            :patch 39})

(definstrument midi-synth-bass-2
  :aliases []
  :config  {:type :midi
            :patch 40})

;; 41-48: STRINGS ;;

(definstrument midi-violin
  :aliases ["violin"]
  :config  {:type :midi
            :patch 41})

(definstrument midi-viola
  :aliases ["viola"]
  :config  {:type :midi
            :patch 42})

(definstrument midi-cello
  :aliases ["cello"]
  :config  {:type :midi
            :patch 43})

(definstrument midi-contrabass
  :aliases ["string-bass" "arco-bass" "double-bass" "contrabass"
            "midi-string-bass" "midi-arco-bass" "midi-double-bass"]
  :config  {:type :midi
            :patch 44})

(definstrument midi-tremolo-strings
  :aliases []
  :config  {:type :midi
            :patch 45})

(definstrument midi-pizzicato-strings
  :aliases []
  :config  {:type :midi
            :patch 46})

(definstrument midi-orchestral-harp
  :aliases ["harp" "orchestral-harp" "midi-harp"]
  :config  {:type :midi
            :patch 47})

; no idea why this is in strings, but ok! ¯\_(ツ)_/¯
(definstrument midi-timpani
  :aliases ["timpani"]
  :config  {:type :midi
            :patch 48})

;; 49-56: ENSEMBLE ;;

(definstrument midi-string-ensemble-1
  :aliases []
  :config  {:type :midi
            :patch 49})

(definstrument midi-string-ensemble-2
  :aliases []
  :config  {:type :midi
            :patch 50})

(definstrument midi-synth-strings-1
  :aliases []
  :config  {:type :midi
            :patch 51})

(definstrument midi-synth-strings-2
  :aliases []
  :config  {:type :midi
            :patch 52})

(definstrument midi-choir-aahs
  :aliases []
  :config  {:type :midi
            :patch 53})

(definstrument midi-voice-oohs
  :aliases []
  :config  {:type :midi
            :patch 54})

(definstrument midi-synth-voice
  :aliases []
  :config  {:type :midi
            :patch 55})

(definstrument midi-orchestra-hit
  :aliases []
  :config  {:type :midi
            :patch 56})

;; 57-64: BRASS ;;

(definstrument midi-trumpet
  :aliases ["trumpet"]
  :config  {:type :midi
            :patch 57})

(definstrument midi-trombone
  :aliases ["trombone"]
  :config  {:type :midi
            :patch 58})

(definstrument midi-tuba
  :aliases ["tuba"]
  :config  {:type :midi
            :patch 59})

(definstrument midi-muted-trumpet
  :aliases []
  :config  {:type :midi
            :patch 60})

(definstrument midi-french-horn
  :aliases ["french-horn"]
  :config  {:type :midi
            :patch 61})

(definstrument midi-brass-section
  :aliases []
  :config  {:type :midi
            :patch 62})

(definstrument midi-synth-brass-1
  :aliases []
  :config  {:type :midi
            :patch 63})

(definstrument midi-synth-brass-2
  :aliases []
  :config  {:type :midi
            :patch 64})

;; 65-72: REED ;;

(definstrument midi-soprano-saxophone
  :aliases ["midi-soprano-sax"
            "soprano-saxophone" "soprano-sax"]
  :config  {:type :midi
            :patch 65})

(definstrument midi-alto-saxophone
  :aliases ["midi-alto-sax"
            "alto-saxophone" "alto-sax"]
  :config  {:type :midi
            :patch 66})

(definstrument midi-tenor-saxophone
  :aliases ["midi-tenor-sax"
            "tenor-saxophone" "tenor-sax"]
  :config  {:type :midi
            :patch 67})

(definstrument midi-baritone-saxophone
  :aliases ["midi-baritone-sax" "midi-bari-sax"
            "baritone-saxophone" "baritone-sax" "bari-sax"]
  :config  {:type :midi
            :patch 68})

(definstrument midi-oboe
  :aliases ["oboe"]
  :config  {:type :midi
            :patch 69})

(definstrument midi-english-horn
  :aliases ["english-horn"]
  :config  {:type :midi
            :patch 70})

(definstrument midi-bassoon
  :aliases ["bassoon"]
  :config  {:type :midi
            :patch 71})

(definstrument midi-clarinet
  :aliases ["clarinet"]
  :config  {:type :midi
            :patch 72})

;; 73-80: PIPE ;;

(definstrument midi-piccolo
  :aliases ["piccolo"]
  :config  {:type :midi
            :patch 73})

(definstrument midi-flute
  :aliases ["flute"]
  :config  {:type :midi
            :patch 74})

(definstrument midi-recorder
  :aliases ["recorder"]
  :config  {:type :midi
            :patch 75})

(definstrument midi-pan-flute
  :aliases ["pan-flute"]
  :config  {:type :midi
            :patch 76})

(definstrument midi-bottle
  :aliases ["bottle"]
  :config  {:type :midi
            :patch 77})

(definstrument midi-shakuhachi
  :aliases ["shakuhachi"]
  :config  {:type :midi
            :patch 78})

(definstrument midi-whistle
  :aliases ["whistle"]
  :config  {:type :midi
            :patch 79})

(definstrument midi-ocarina
  :aliases ["ocarina"]
  :config  {:type :midi
            :patch 80})

;; 81-88: SYNTH LEAD ;;

(definstrument midi-square-lead
  :aliases ["square" "square-wave" "square-lead"
            "midi-square" "midi-square-wave"]
  :config  {:type :midi
            :patch 81})

(definstrument midi-saw-wave
  :aliases ["sawtooth" "saw-wave" "saw-lead"
            "midi-sawtooth" "midi-saw-lead"]
  :config  {:type :midi
            :patch 82})

(definstrument midi-calliope-lead
  :aliases ["calliope-lead" "calliope"
            "midi-calliope"]
  :config  {:type :midi
            :patch 83})

(definstrument midi-chiffer-lead
  :aliases ["chiffer-lead" "chiffer" "chiff"
            "midi-chiffer" "midi-chiff"]
  :config  {:type :midi
            :patch 84})

(definstrument midi-charang
  :aliases ["charang"]
  :config  {:type :midi
            :patch 85})

(definstrument midi-solo-vox
  :aliases []
  :config  {:type :midi
            :patch 86})

(definstrument midi-fifths
  :aliases ["midi-sawtooth-fifths"]
  :config  {:type :midi
            :patch 87})

(definstrument midi-bass-and-lead
  :aliases ["midi-bass+lead"]
  :config  {:type :midi
            :patch 88})

;; 89-96: SYNTH PAD ;;

(definstrument midi-synth-pad-new-age
  :aliases ["midi-pad-new-age" "midi-new-age-pad"]
  :config  {:type :midi
            :patch 89})

(definstrument midi-synth-pad-warm
  :aliases ["midi-pad-warm" "midi-warm-pad"]
  :config  {:type :midi
            :patch 90})

(definstrument midi-synth-pad-polysynth
  :aliases ["midi-pad-polysynth" "midi-polysynth-pad"]
  :config  {:type :midi
            :patch 91})

(definstrument midi-synth-pad-choir
  :aliases ["midi-pad-choir" "midi-choir-pad"]
  :config  {:type :midi
            :patch 92})

(definstrument midi-synth-pad-bowed
  :aliases ["midi-pad-bowed" "midi-bowed-pad"
            "midi-pad-bowed-glass" "midi-bowed-glass-pad"]
  :config  {:type :midi
            :patch 93})

(definstrument midi-synth-pad-metallic
  :aliases ["midi-pad-metallic" "midi-metallic-pad"
            "midi-pad-metal" "midi-metal-pad"]
  :config  {:type :midi
            :patch 94})

(definstrument midi-synth-pad-halo
  :aliases ["midi-pad-halo" "midi-halo-pad"]
  :config  {:type :midi
            :patch 95})

(definstrument midi-synth-pad-sweep
  :aliases ["midi-pad-sweep" "midi-sweep-pad"]
  :config  {:type :midi
            :patch 96})

;; 97-104: SYNTH EFFECTS ;;

(definstrument midi-fx-rain
  :aliases ["midi-fx-ice-rain" "midi-rain" "midi-ice-rain"]
  :config  {:type :midi
            :patch 97})

(definstrument midi-fx-soundtrack
  :aliases ["midi-soundtrack"]
  :config  {:type :midi
            :patch 98})

(definstrument midi-fx-crystal
  :aliases ["midi-crystal"]
  :config  {:type :midi
            :patch 99})

(definstrument midi-fx-atmosphere
  :aliases ["midi-atmosphere"]
  :config  {:type :midi
            :patch 100})

(definstrument midi-fx-brightness
  :aliases ["midi-brightness"]
  :config  {:type :midi
            :patch 101})

(definstrument midi-fx-goblins
  :aliases ["midi-fx-goblin" "midi-goblins" "midi-goblin"]
  :config  {:type :midi
            :patch 102})

(definstrument midi-fx-echoes
  :aliases ["midi-fx-echo-drops" "midi-echoes" "midi-echo-drops"]
  :config  {:type :midi
            :patch 103})

(definstrument midi-fx-sci-fi
  :aliases ["midi-sci-fi"]
  :config  {:type :midi
            :patch 104})

;; 105-112: ETHNIC (das racist) ;;

(definstrument midi-sitar
  :aliases ["sitar"]
  :config  {:type :midi
            :patch 105})

(definstrument midi-banjo
  :aliases ["banjo"]
  :config  {:type :midi
            :patch 106})

(definstrument midi-shamisen
  :aliases ["shamisen"]
  :config  {:type :midi
            :patch 107})

(definstrument midi-koto
  :aliases ["koto"]
  :config  {:type :midi
            :patch 108})

(definstrument midi-kalimba
  :aliases ["kalimba"]
  :config  {:type :midi
            :patch 109})

(definstrument midi-bagpipes
  :aliases ["bagpipes"]
  :config  {:type :midi
            :patch 110})

(definstrument midi-fiddle
  :aliases []
  :config  {:type :midi
            :patch 111})

(definstrument midi-shehnai
  :aliases ["shehnai" "shahnai" "shenai" "shanai"
            "midi-shahnai" "midi-shenai" "midi-shanai"]
  :config  {:type :midi
            :patch 112})

;; 113-120: PERCUSSIVE ;;

(definstrument midi-tinkle-bell
  :aliases ["midi-tinker-bell"]
  :config  {:type :midi
            :patch 113})

(definstrument midi-agogo
  :aliases []
  :config  {:type :midi
            :patch 114})

(definstrument midi-steel-drums
  :aliases ["midi-steel-drum"
            "steel-drums" "steel-drum"]
  :config  {:type :midi
            :patch 115})

(definstrument midi-woodblock
  :aliases []
  :config  {:type :midi
            :patch 116})

(definstrument midi-taiko-drum
  :aliases []
  :config  {:type :midi
            :patch 117})

(definstrument midi-melodic-tom
  :aliases []
  :config  {:type :midi
            :patch 118})

(definstrument midi-synth-drum
  :aliases []
  :config  {:type :midi
            :patch 119})

(definstrument midi-reverse-cymbal
  :aliases []
  :config  {:type :midi
            :patch 120})

;; 121-128: SOUND EFFECTS ;;

(definstrument midi-guitar-fret-noise
  :aliases []
  :config  {:type :midi
            :patch 121})

(definstrument midi-breath-noise
  :aliases []
  :config  {:type :midi
            :patch 122})

(definstrument midi-seashore
  :aliases []
  :config  {:type :midi
            :patch 123})

(definstrument midi-bird-tweet
  :aliases []
  :config  {:type :midi
            :patch 124})

(definstrument midi-telephone-ring
  :aliases []
  :config  {:type :midi
            :patch 125})

(definstrument midi-helicopter
  :aliases []
  :config  {:type :midi
            :patch 126})

(definstrument midi-applause
  :aliases []
  :config  {:type :midi
            :patch 127})

(definstrument midi-gunshot
  :aliases ["midi-gun-shot"]
  :config  {:type :midi
            :patch 128})

;; Percussion

(definstrument midi-percussion
  :aliases ["percussion"]
  :config  {:type :midi
            :percussion? true})

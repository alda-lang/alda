(ns alda.lisp.instruments.midi)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.instruments.midi...")

; NOTE: For the time being, some of these instruments have non-prefixed names
;       like "piano" and "trumpet" as aliases. I eventually want to re-map
;       those names to more realistic-sounding, sampled instruments instead of
;       MIDI. Once that happens, we'll need to remove the aliases.

; reference: http://www.jimmenard.com/midi_ref.html#General_MIDI

;; 1-8: PIANO ;;

(definstrument midi-acoustic-grand-piano
  :aliases ["midi-piano" "piano"]
  :config {:type :midi
           :patch 1})

(definstrument midi-bright-acoustic-piano
  :aliases []
  :config {:type :midi
           :patch 2})

(definstrument midi-electric-grand-piano
  :aliases []
  :config {:type :midi
           :patch 3})

(definstrument midi-honky-tonk-piano
  :aliases []
  :config {:type :midi
           :patch 4})

(definstrument midi-electric-piano-1
  :aliases []
  :config {:type :midi
           :patch 5})

(definstrument midi-electric-piano-2
  :aliases []
  :config {:type :midi
           :patch 6})

(definstrument midi-harpsichord
  :aliases ["harpsichord"]
  :config {:type :midi
           :patch 7})

(definstrument midi-clavi
  :aliases ["midi-clavinet" "clavinet"]
  :config {:type :midi
           :patch 8})

;; 9-16: CHROMATIC PERCUSSION ;;

(comment "TODO")

;; 17-24: ORGAN ;;

(comment "TODO")

;; 25-32: GUITAR ;;

(definstrument midi-acoustic-guitar-nylon
  :aliases ["midi-acoustic-guitar" "acoustic-guitar" "guitar"]
  :config {:type :midi
           :patch 25})

(definstrument midi-acoustic-guitar-steel
  :aliases []
  :config {:type :midi
           :patch 26})

(definstrument midi-electric-guitar-jazz
  :aliases []
  :config {:type :midi
           :patch 27})

(definstrument midi-electric-guitar-clean
  :aliases ["electric-guitar-clean"]
  :config {:type :midi
           :patch 28})

(definstrument midi-electric-guitar-palm-muted
  :aliases []
  :config {:type :midi
           :patch 29})

(definstrument midi-electric-guitar-overdrive
  :aliases ["electric-guitar-overdrive"]
  :config {:type :midi
           :patch 30})

(definstrument midi-electric-guitar-distorted
  :aliases ["electric-guitar-distorted"]
  :config {:type :midi
           :patch 31})

(definstrument midi-electric-guitar-harmonics
  :aliases ["electric-guitar-harmonics"]
  :config {:type :midi
           :patch 32})


;; 33-40: BASS ;;

(comment "TODO")

;; 41-48: STRINGS ;;

(definstrument midi-violin
  :aliases ["violin"]
  :config {:type :midi
           :patch 41})

(definstrument midi-viola
  :aliases ["viola"]
  :config {:type :midi
           :patch 42})

(definstrument midi-cello
  :aliases ["cello"]
  :config {:type :midi
           :patch 43})

(definstrument midi-contrabass
  :aliases ["string-bass" "arco-bass" "double-bass" "contrabass"
            "midi-string-bass" "midi-arco-bass" "midi-double-bass"]
  :config {:type :midi
           :patch 44})

(definstrument midi-tremolo-strings
  :aliases []
  :config {:type :midi
           :patch 45})

(definstrument midi-pizzicato-strings
  :aliases []
  :config {:type :midi
           :patch 46})

(definstrument midi-orchestral-harp
  :aliases ["harp" "orchestral-harp" "midi-harp"]
  :config {:type :midi
           :patch 47})

; no idea why this is in strings, but ok!
(definstrument midi-timpani
  :aliases ["timpani"]
  :config {:type :midi
           :patch 48})

;; 49-56: ENSEMBLE ;;

(comment "TODO")

;; 57-64: BRASS ;;

(definstrument midi-trumpet
  :aliases ["trumpet"]
  :config {:type :midi
           :patch 57})

(definstrument midi-trombone
  :aliases ["trombone"]
  :config {:type :midi
           :patch 58})

(definstrument midi-tuba
  :aliases ["tuba"]
  :config {:type :midi
           :patch 59})

(definstrument midi-muted-trumpet
  :aliases []
  :config {:type :midi
           :patch 60})

(definstrument midi-french-horn
  :aliases ["french-horn"]
  :config {:type :midi
           :patch 61})

(definstrument midi-brass-section
  :aliases []
  :config {:type :midi
           :patch 62})

(definstrument midi-synth-brass-1
  :aliases []
  :config {:type :midi
           :patch 63})

(definstrument midi-synth-brass-2
  :aliases []
  :config {:type :midi
           :patch 64})

;; 65-72: REED ;;

(comment "TODO")

;; 73-80: PIPE ;;

(comment "TODO")

;; 81-88: SYNTH LEAD ;;

(comment "TODO")

;; 89-96: SYNTH PAD ;;

(comment "TODO")

;; 97-104: SYNTH EFFECTS ;;

(comment "TODO")

;; 105-112: ETHNIC (das racist) ;;

(comment "TODO")

;; 113-120: PERCUSSIVE ;;

(comment "TODO")

;; 121-128: SOUND EFFECTS ;;

(comment "TODO")

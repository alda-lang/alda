(ns alda.lisp.instruments.midi)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.instruments.midi...")

; TODO: define MIDI instruments here

(definstrument piano
  :aliases ["midi-piano"]
  :type :midi)

(definstrument trumpet
  :aliases ["midi-trumpet"]
  :type :midi)

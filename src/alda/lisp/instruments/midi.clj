(ns alda.lisp.instruments.midi)
(in-ns 'alda.lisp)

(log/debug "Loading alda.lisp.instruments.midi...")

; TODO: define MIDI instruments here

(definstrument piano
  :aliases ["midi-piano"]
  :config {:type :midi})

(definstrument trumpet
  :aliases ["midi-trumpet"]
  :config {:type :midi})

(definstrument cello
  :aliases ["midi-cello"]
  :config {:type :midi})

(definstrument violin
  :aliases ["midi-violin"]
  :config {:type :midi})

(definstrument viola
  :aliases ["midi-viola"]
  :config {:type :midi})

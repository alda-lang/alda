# General ideas

- Entering an Alda file line-by-line should be essentially the same as evaluating the file.

# REPL commands

- [ ] **:new** / **:new score** -- invokes `(score*)`
- [ ] **:new part (instrument) (opt. nickname)** -- switches to a new instrument part, creating it if applicable. Should also print a confirmation message like "piano-aXg4k (midi-acoustic-grand-piano) was created." *(Note: as an alternative to this, the user can just enter 'piano "my-piano":' like they would in an Alda file.)* Omitting any args should cue an interactive prompt asking for the instrument(s), then the nickname (enter through if no nickname).
- [ ] **:part** or **:parts** or **:instruments** shows current instruments, formatted in a friendly way that shows you if the instruments are in a group
- [ ] **:info** should print a bunch of information about the score, such as the instruments, current instruments, markers, user-readable length of the piece, etc.
- [ ] **:info (inst-id)** should print info about an instrument instance (offset, octave, stock instrument, etc.)
- [ ] **:map** prints the data representation of the score in progress (score-map)
- [ ] **:graph** should display an ASCII graph of the score. I'm envisioning sort of a horizontal bar graph with a different color bar for each instrument instance, showing you when the instrument is playing. At the bottom is a legend showing which color is which instrument, and labels for the time markings at start and end of the piece. Maybe include markers, too.
- [ ] The user can save the score in progress to a file at any time by entering **:save (filename)** or just **:save** (which will re-save it if there's already a file specified, otherwise prompt for a filename). There should be a **:save-as (filename)** too -- or maybe that functionality can just be baked into the :save command -- e.g. if you already saved it as "a.alda," entering ":save b.alda" will do a save-as, and future :save's will re-save b.alda.
- [ ] **:export** should save the audio, and take the same options as the CLI command for saving a script as wav, mp3, etc. Maybe we should infer from the file ending what type of encoding the user wants.
- [ ] **:load (filename)** should load a score into memory and leave it open for editing
- [ ] TODO: figure out some sensible way to edit the score in progress... Maybe just an **:edit** command that opens the temp file in an $EDITOR subprocess
- [ ] **:play** should play the score from the top. Should take optional args to start/end at a certain marker or minute/second mark.
- [ ] Ideally, when any audio is playing, it should happen in the background and the user should be able to start typing the next command. Subsequent commands that want to play audio should stop any currently playing audio. User should also be able to **:stop** and maybe even **:pause** explicitly.
- [ ] **:help** should display available commands with descriptions. :help <command> should display options.

# Features

- [ ] Show custom prompts depending on context. Show the current instrument(s), current voice if applicable, maybe the current marker/offset (leaning toward offset)
- [ ] Tab completion (for instance ids/names, stock instruments, maybe also for filenames when loading a file)
- [ ] When a user enters notes, play them and also print them to the console in a friendly format. Maybe even print each note as it plays!
- [ ] It might be cool if, when playing a score in interactive mode, it prints all the notes in real-time too. Notes for different instruments could be printed in different colors. This could be potentially annoying for large scores, so should maybe make it an option/setting.

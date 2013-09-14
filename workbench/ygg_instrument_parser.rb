require 'treetop'
Treetop.load 'instrument_grammar'

class MidiTrack

  # This is a hash of { name => [MidiTrack_instances] }.
  # The parser uses it to look up if an instrument by a given name exists,
  # and to look up specific MidiTrack instance(s), given the (group) name.
  @@instruments = {} 

  attr_accessor :midi_instrument, :note_data, :name

  def initialize(midi_instrument, note_data, name=instrument)
    @midi_instrument = midi_instrument
    @note_data = note_data
    @name = name

    @@instruments[@name] = [self]
  end

  def add_note_data(note_data)
    # to do
  end

end

class TrackGroup

  attr_accessor :name, :tracks

  def initialize(name, tracks)
    @name = name
    @tracks = tracks   # array of tracks/groups
  end

  =begin
  This function works differently depending on if it's being
  called by a MidiTrack instance or a TrackGroup instance. 
  This is for recursion through an arbitrary number of levels
  of groups containing groups containing groups (etc.)
  containing tracks.
  =end  
  def add_note_data(note_data)
    @tracks.each do |instrument|
      instrument.add_note_data(note_data)
    end
  end

end

  

end

=begin

For each "block" (instrument_def + note_data):
  Does the instrument_def specify a name?
     NO: 
         First, check all instruments to make sure they are valid MIDI 
         instruments as defined in Yggdrasil's master MIDI instrument list.
         If not, raise error.

         One instrument (or group) only?
         Does MidiTrack.instruments.include? the instrument (=name)
            YES: 
               MidiTrack.instruments[name].each do |instrument| 
                 instrument.add_note_data(note_data)
               end   DONE
            NO: algorithm for naming MidiTrack object (piano_1, etc.)
                said name = MidiTrack.new(midipatch, note_data)  DONE

         Multiple instruments?
            Go through the "one instrument" instructions for each instrument
            (or group) in succession.  DONE

     YES: (name or group name specified)
         First, check all instruments to make sure they either exist in
         MidiTracks.instruments (i.e. the names already exist), or otherwise
         they are valid MIDI instruments so that new tracks can be created for
         them. If neither of these conditions are met, raise error.

         One instrument only?
            Does MidiTrack.instruments.include? "name"
               YES: Raise error (name already exists - cannot reuse)
               NO: algorithm for naming MidiTrack object
                   said name = MidiTrack.new(midipatch, note_data, name) DONE

         Multiple instruments? (meaning this is a group name)
            Does MidiTrack.instruments.include? "name"
               YES: Raise error (cannot reuse name)
               NO: 
                  1) MidiTracks.instruments["groupname"] = []

                  2) for each instrument, do:
                       MidiTracks.instruments[name].each do |instrument|
                         instrument.add_note_data(note_data)
                       end   DONE     (works for TrackGroups too,
                                        as they have their own add_note_data
                                          function)

=end

parser = InstrumentParser.new
parser.parse(pass ygg code in here)
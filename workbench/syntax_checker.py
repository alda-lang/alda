#! /usr/bin/python2

from sys import argv
from sys import exit

script, filename = argv

class InstrumentError(Exception):
    """Base class for errors related to defining
    and calling instruments in yggdrasil code."""
    
def get_instruments():

    instruments = []
    txt = open(filename)

    for line in txt:
        if "add" in line:
            line_items = line.split()
            x = line_items.index("add")
            instruments.append(line_items[x+1])
            
    return instruments

    
def check_syntax():
    txt = open(filename)
    
    # checks to see if there is any note data in the file
    # for which the instrument is unclear / not defined

    okay = True
    error_lines = []
    bad_instruments = []
    line_number = 1
    instrument_not_defined = True
            
    for line in txt:
        if not line.strip():   # if line is empty,
            line_number += 1
            continue           # skips the line

        if ":" in line:
            line_type = "call instrument"
            instrument_not_defined = False
        
            y = line.index(":")     # checks to make sure the instrument being called 
            temp_line = line[:y]    # is defined somewhere in the file
            temp_words = temp_line.split('/')
            for w in temp_words:
                if w not in get_instruments():
                    okay = False  # no such instrument(s) error
                    bad_instruments.append(w)
                else:
                    okay = True
            if okay == False:
                if len(bad_instruments) == 1:
                    new_error = "Error in line %d: instrument '%s' not defined" % (line_number, bad_instruments[0])
                elif len(bad_instruments) == 2:
                    new_error = "Error in line %d: instruments '%s' and '%s' not defined" % (line_number, bad_instruments[0], bad_instruments[1])
                elif len(bad_instruments) > 2:
                    last_instrument = bad_instruments.pop()
                    other_instruments = "', '".join(bad_instruments)
                    bad_instrument_string = "'%s' and '%s'" % (other_instruments, last_instrument)
                    new_error = "Error in line %d: instruments %s not defined" % (line_number, bad_instrument_string)
                error_lines.append(new_error)
                bad_instruments = []
        elif "add" in line:
            line_type = "add instrument"
            okay = True
            instrument_not_defined = True
        # to do: check to make sure the instrument 
        # being added is a valid instrument
            
        else: 
            line_type = "note data"
            if instrument_not_defined:
                okay = False  # undefined instrument error
                new_error = "Error in line %d: instrument unclear" % line_number
                error_lines.append(new_error)
                        
        line_number += 1
        last_line = line_type

    txt.close()
    
    if error_lines:
        for line in error_lines:
            print line
        exit(1)

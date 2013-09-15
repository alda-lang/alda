#! /usr/bin/python2

from sys import argv
from sys import exit
from ygg_lex import tokens
from ygg_parse import parser
import syntax_checker
import ply.yacc as yacc

script, filename = argv

def print_file():

    txt = open(filename)
    print "Contents of %r:" % filename
    print "-------------" + "-" * len(filename) + "-"

    for line in txt:
        print line,
        
    txt.close()
    
def get_instruments():

    instruments = []
    txt = open(filename)

    for line in txt:
        if "add" in line:
            line_items = line.split()
            x = line_items.index("add")
            instruments.append(line_items[x+1])
            
    return instruments
    
def get_notes(instrument):
    txt = open(filename)
    start_tag = instrument + ":"
    start_group = instrument + "/"
    in_group = "/" + instrument + "/"
    last_in_group = "/" + instrument + ":"
    reading = False    
    note_data = []

    # returns a [] list of lines of note data belonging to
    # each instrument, including "\n" for carriage returns
    for line in txt:
        if line.startswith(start_tag):
            reading = True
        if line.startswith(start_group):
            reading = True
        if in_group in line:
            reading = True
        if last_in_group in line:
            reading = True
        if "add" in line:
            reading = False
        if ":" in line and not instrument in line:
            reading = False
            
        if reading == True:
            if line.strip():  # if line is not empty,
                line = line.strip(' \n')  # removes "\n" and
                note_data.append(line)    # adds to note data

    txt.close()

    note_data = " ".join(note_data)

### TO DO: fix this so that voices (V1:) don't count
###        as instruments...
    
    #removes the instrument names (anything ending in ":")
    notes_split = note_data.split()
    for x in notes_split:
        if ":" in x:
            notes_split.remove(x)
    note_data = " ".join(notes_split)
            
    return note_data

syntax_checker.check_syntax()

for i in get_instruments():
    print i.upper() + ":"
    parser.parse(get_notes(i))
    print "\n\n"

#! /usr/bin/python2

from ygg_lex import tokens
import ply.yacc as yacc
import mingus.core.value as value

# MIDI "working variables":
currentOctave = 4
currentDuration = 4.0
currentVoice = 0

# yggdrasil functions:
def yggChange(attribute, amount):
    'Changes x attribute by y amount; called by att_event'
    print "%s value changed to %s." % (attribute, amount)

# parser functions:
def p_error(p):
    print "Syntax error! %s" % p
    
def p_events(p):
    '''events : event
              | events event'''
    pass
    
def p_event(p):
    '''event : note_event
             | rest_event
             | chord_event
             | att_event
             | oct_change
             | voice'''
    p[0] = p[1]

def p_att_event(p):
    'att_event : LPAREN att_change RPAREN'
    for ac in p[2]:
        yggChange(ac[0], ac[1])

def p_att_change(p):
    '''att_change : attribute NUMBER
                  | attribute NUMBER SEMICOLON att_change'''
    if len(p) == 3:
        p[0] = [(p[1], p[2])]
    elif len(p) > 3:
        p[0] = p[4].insert(0, (p[1], p[2]))

def p_attribute(p):
    '''attribute : VOLUME
                 | PAN 
                 | TEMPO'''
    p[0] = p[1]

def p_oct_change(p):
    '''oct_change : OCTAVE NUMBER 
                  | OCT_UP 
                  | OCT_DOWN'''
    global currentOctave
    if p[1] == 'o':
        currentOctave = p[2]
    elif p[1] == '<':
        currentOctave -= 1
    elif p[1] == '>':
        currentOctave += 1
        
def p_chord_event(p):
    '''chord_event : chord'''
    print "chord:"
    print p[1]

def p_chord(p):
    '''chord : note SLASH note
             | note SLASH rest
             | rest SLASH note
             | rest SLASH rest
             | note SLASH chord
             | rest SLASH chord'''
    p[0] = []
    for x in p[1]: p[0].append(x)
    for x in p[3]: p[0].append(x)
    
def p_note_event(p):
    '''note_event : note'''
    pi = p[1][0].upper().replace('+','#').replace('-','b')    
    print "Note: %s%d, duration %d" % (pi, p[1][1], p[1][2])

def p_note(p):
    '''note : PITCH
            | PITCH duration'''
    global currentDuration
    pi = p[1]
    oc = currentOctave
    if len(p) == 2:
        #print "2 args:", p[0], p[1]
        du = currentDuration
    elif len(p) == 3:
        #print "3 args:", p[0], p[1], p[2]
        du = p[2]
        currentDuration = du
    p[0] = (pi, oc, du)

def p_rest_event(p):
    '''rest_event : rest'''
    print "rest-%d" % p[1]

def p_rest(p):
    '''rest : REST
            | REST duration'''
    global currentDuration
    if len(p) == 2:
        p[0] = currentDuration
    elif len(p) == 3:
        p[0] = p[2]
        currentDuration = p[2]

def p_duration(p):
    '''duration : notelength
                | notelength TIE notelength'''
    if len(p) == 2:
        p[0] = p[1]
    elif len(p) == 4:
        p[0] = value.add(p[1], p[3])

def p_notelength(p):
    '''notelength : NUMBER
                  | NUMBER dots'''
    if len(p) == 2:
        p[0] = p[1]
    elif len(p) == 3:
        p[0] = value.dots(p[1], p[2])

def p_dots(p):
    '''dots : DOT
            | DOT dots'''
    if len(p) == 2:
        p[0] = 1
    elif len(p) == 3:
        p[0] += 1
    print "dots: %d" % p[0]
    
def p_voice(p):
    '''voice : VOICE'''
    pass
            
parser = yacc.yacc()

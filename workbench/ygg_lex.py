#! /usr/bin/python2

# tokenizer for yggdrasil code. First, main.py takes care of the
# general setup, which instruments are called, etc., checks for 
# general syntax errors (calling invalid instruments, etc.), and 
# creates one string of note data per instrument. lexer.py works
# on these strings of note data.

import ply.lex as lex

tokens = (
    'PITCH',
    'DOT',
    'SEMICOLON',
    'TIE',
    'REST',
    'SLASH',
    'LPAREN',
    'RPAREN',
    'OCTAVE',
    'OCT_UP',
    'OCT_DOWN',
    'VOLUME',
    'PAN',
    'TEMPO',
    'VOICE',
    'NUMBER',
)

t_PITCH         = r'[a-g][+-]*'
t_DOT           = r'\.'
t_SEMICOLON     = r';'
t_TIE           = r'~'
t_REST          = r'r'
t_SLASH         = r'/'
t_LPAREN        = r'\('
t_RPAREN        = r'\)'
t_OCTAVE        = r'o' 
t_OCT_UP        = r'>'
t_OCT_DOWN      = r'<'
t_VOLUME        = r'volume'
t_PAN           = r'pan'
t_TEMPO         = r'tempo'
t_VOICE         = r'V[0-9]+:'

def t_NUMBER(t):
    r'-?[0-9]+'
    t.value = float(t.value)
    return t

t_ignore = ' \t'

def t_error(t):
    print "Illegal unit '%s'" % t.value[0]
    t.lexer.skip(1)

lexer = lex.lex()

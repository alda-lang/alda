# Instance and Group Assignment

The way Alda assigns instances is slightly complicated, but ultimately lends itself to intuitive score-writing for the end user.

1. Just the name of the instrument, e.g. `piano:` -- this is a call to piano, which works because piano is a stock instrument that Alda knows about. The first time this call is made, the instance is registered as "piano 1." 

        piano: c c c c c   # piano 1
        cello: g g g g g   # cello 1
        piano: e e e e e   # still piano 1
        cello: c c c c c   # still cello 1

    If you then call a different instrument, say `cello:`, a new instrument instance is created ("cello 1"). Then, if you switch back to `piano:`, Alda will see that piano is already registered, and you'll be appending music events to the same instance, piano 1. **A new instance of an instrument is created whenever a stock instrument is called *that is not already in use*.**

2. Multiple instruments can be combined into groups like this: `trumpet/trombone/tuba:`. Any music events following such a call will be applied to each instrument in the group. **The instances are assigned just like with single instruments -- new instances will be created for any instrument that does not already exist, and if an instance of the instrument does exist, the music data will be appended to that instance.** (It's worth noting that the instrument instances in a group will have the same music events, but they don't necessarily have to start at the same time -- each instance will start the music events whenever it's finished with its own preceding events.)

        trumpet: c4 c8 c c2   # trumpet 1
        trumpet/trombone/tuba: c d e f g1   # trumpet 1 (still), trombone 1 and tuba 1

3. Instruments can be nicknamed, and in fact this is necessary if you want to have two or more instances of the same instrument:

        flute "bill": c d e f g2. # flute 1
        flute "bob":  e f g a b2. # flute 2

    **A new instance of an instrument is created whenever a stock instrument is called with any nickname, either as part of a group or not.**

        flute: g a b > c d2.       # flute 1
        flute "bill": c d e f g2.  # flute 2
        flute "bob":  e f g a b2.  # flute 3

    Another example: If there is already a clarinet 1 and there is already a cello 1 nicknamed 'thor', then the call `thor/clarinet 'band':` will refer to the same instance of cello (cello 1, 'thor'), but a *new* clarinet instance (clarinet 2) because a nickname, 'band' is being given to this group, and 'clarinet' refers to the stock instrument, not any particular named instance of clarinet. On the other hand, `thor/clarinet:` in the same scenario will refer to cello 1 and clarinet 1, the same instances that were already in use.

    Example 1:

        clarinet: g g g g2.                # clarinet 1
        cello "thor": g b d > g2.          # cello 1
        thor/clarinet "band": g d < b g2.  # cello 1 (still) and clarinet 2

    Example 2:

        clarinet: g g g g2.                # clarinet 1
        cello "thor": g b d > g2.          # cello 1
        thor/clarinet: g < d b g2.         # cello 1 (still) and clarinet 1 (still)

    Again, the key thing to remember is that **a new instance of an instrument is created whenever a stock instrument is called with any nickname, either as part of a group or not.**

Generally, I would recommend that score writers use the names of the stock instruments instead of nicknames if there is only one instance of each instrument -- and if there is more than one, assign nicknames to each instrument the first time it is called. (If you do it this way, you don't have to understand how any of the above works :smiley_cat:)

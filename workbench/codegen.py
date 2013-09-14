import string
import random

codes = []
possChars = [x for x in string.uppercase] + [str(x) for x in xrange(10)]

def generateCode():
    'Generates one code.'
    newCode = ""
    for x in xrange(3): 
        newCode += random.choice(possChars)
    lets, nums = 0, 0
    for c in newCode:
        if c in string.uppercase: 
            lets += 1
        if c in map(str, xrange(10)): 
            nums += 1
    if lets < 3 and nums < 3 and newCode not in codes:
        codes.append(newCode)

def generateCodes(n):
    'Generates n codes.'
    for x in xrange(n): generateCode()

generateCodes(10)
generateCodes(10)

for code in codes: print code
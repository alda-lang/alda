# Comments

You can add **comments** to your Alda scores by placing a `#` character to the
left of each line that you'd like to comment out.

```alda
# This is a comment.
piano: c d e f
```

Alda will ignore everything following a `#` character on each line.

```alda
# trumpet: c c c c   <- you will NOT hear that
piano: c d e f     # <- you WILL hear that
```

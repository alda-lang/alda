# 声部

*此页面翻译自[voice.md](../voice.md)*

**Voices** provide a way to subdivide an instrument into its own separate parts,
which play simultaneously. This can be useful for polyphonic instruments, that
is, instruments that can play more than one [note](notes.md) at a time, e.g.
guitar, piano.
**声部**提供了一个...的方法 

乐器可以同时奏响更多的音符 比如吉他 钢琴

### 例子

```alda
piano:
  V1: c d e f g1
  V2: e f g a b1
  V3: g a b > c d1

  V0: c4 e g > c2.
```

Each voice is its own sequence of note events. The first note/[rest](rests.md)
in each voice starts at the same time, like the notes in a [chord](chords.md).
Whereas a chord bumps forward the current [offset](offset.md) by the shortest
note duration in the chord, after a group of voices, the current offset is that
of the longest voice in the group. `V0:` signals the end of a voice grouping and
a return to using a single voice -- the first note placed after `V0:` will
happen after all voices in the group have finished.


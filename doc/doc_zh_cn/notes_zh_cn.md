# 音符

*本页面翻译自[Notes](../notes.md)*

Alda的**音符**深受[MML](https://en.wikipedia.org/wiki/Music_Macro_Language#Modern_MML)的启发

## 组件

### 八度组

西方音乐理论将音高分为12个音符的重复组(十二平均律) 例如: (升序的) `c c# d d# e f f# g g# a a# b (高八度) c c# d`等等

音名(如C#)与八度的组合决定了音符的频率 八度组用数字表示 通常在1到7之间 它们对应于[科学音高记号](https://zh.wikipedia.org/wiki/%E7%A7%91%E5%AD%A6%E9%9F%B3%E9%AB%98%E8%AE%B0%E5%8F%B7)(也称国际音高记号)

例如 中央C与A440都在第4个八度组中 这也是Alda中默认的八度组

与MML一样 八度组与音符本身是分开设置的--八度组不是"附加"于音符 而是每个音符都查看当前的八度组来确定自己的音高

您可以用这两种方式设置八度组:

`o5` 将八度组设置为5 任何整数都可以跟在`o`后面

`<` 将当前八度组减少1. `>`将当前八度组增加1.

### 时值

在Alda中 音符的时值通常以标准音乐符号中的音符长度表示 与MML一样

它以数字形式表示 4是四分音符 2是二分音符 1是全音符 等等

音符后面可以添加任意数量的点`.` 它代表附点 与标准音乐符号相同--它会将音符的时值延长到其原先时值的1.5倍(也就是延长其一半的时值)

例如

```
2 = 二分音符 2拍
2. = 附点二分音符 3拍(2 * 1.5)
2.. = 复附点二分音符 3.5拍(2 * 1.5 + (2 * 0.5 * 0.5))
```

*译者注: 当一个音符有复附点时 每次附的点所延长的时值是上一个点所延长的时值的一半*

音符时值也可以用连音线语法`~`相加 例如: `4~4` = 两个四分音符加在一起 总共2拍

Alda keeps track of both the current octave and the current default note
duration as notes are processed sequentially in a score. Each time a note
duration is specified, that duration becomes the new default note duration. Each
note that follows, when no note duration is specified, will have the default
note duration. At the beginning of each instrument part, the default octave is 4
and the default note duration is 4 (i.e. a quarter note, 1 beat).

#### Advanced Rhythms

* A special feature of Alda is that you can use non-standard numbers as note
  durations. For example, 6 is a note that lasts 1/6 of a measure in 4/4 time.
  In standard notation, there is no such thing as a "sixth note," but this note
  length would be commonly expressed as one note in a quarter note triplet; in
  Alda, a "6th note" doesn't necessarily need to be part of a triplet, however,
  which offers interesting rhythmic possibilities.

* Extending this concept, Alda allows for non-integer decimal note lengths.

  For example, `c0.5` (or a double whole note, in Western classical notation) is
  twice the length of `c1` (a whole note).

  The numbers do not need to be powers of 2. For example, `c2.4` is valid.

* Alda also has an alternate way of specifying rhythms called a [cram
  expression](cram-expressions.md).

* Note lengths can also be expressed in milliseconds and seconds, which can
  optionally be mixed and matched with standard note lengths:

  ```alda
  c350ms    # a C note lasting 350 milliseconds
  d2s       # a D note lasting 2 seconds
  e2s~200ms # an E note lasting 2 seconds + 200 milliseconds
  f300ms~4. # an F note lasting 300 milliseconds + a dotted quarter note
  ```

### Letter pitch

A note in Alda is expressed as a letter from a-g, any number of accidentals
(optional), and a note duration (also optional).

Flats and sharps will decrease/increase the pitch by one half step, e.g. C + 1/2
step = C#. Flats and sharps are expressed in Alda as `-` and `+`, and you can
have multiple sharps or multiple flats, or even combine them, if you'd like.
e.g. `c++` = C double-sharp = D.

As an alternative to placing flats and sharps on every note that needs them, you
may prefer to set the [key signature](attributes.md#key-signature), which will
add the necessary sharps/flats to any note that needs them in order to match the
key. See below for an example of using a key signature.

To overwrite the flat/sharp specified by a key signature, you can include an
accidental, i.e. `-` or `+` to make the note flat or sharp. You can also
override the key signature and force a note to be natural with `_`, i.e. `c_` is
a C natural regardless of what key you are in.

## Example

The following is a 1-octave B major scale, ascending and descending, starting in
octave 4:

```alda
o4 b4 > c+8 d+ e f+ g+ a+ b4
a+8 g+ f+ e d+ c+ < b2.
```

Here is the same example, using a key signature in order to avoid having to
include all of the sharps:

```alda
(key-signature "f+ c+ g+ d+ a+")
o4 b4 > c8 d e f g a b4
a8 g f e d c < b2.
```


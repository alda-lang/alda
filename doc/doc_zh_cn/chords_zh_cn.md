# 和弦

*此页面翻译自[Chords](../chords.md)*

**和弦**是一组具有相同[偏移量](offset_zh_cn.md)的[音符](notes_zh_cn.md) 即它们开始奏响的时间是相同的 在Alda中表示和弦 可以在组成和弦的音符之间加上斜杠`/` 例如 `c/e/g`

和弦的音符之间可以有八度变化 这允许和弦跨越多个八度: `c/g/>c/e/g`

和弦中的音符可以有不同的长度 在这种情况下 和弦之后的下一个音符事件会发生在**和弦中最短的音符** 之后这样可以轻松获得带有变调的和弦 例如: `c1~1/>c/<e4 f g f e1` (另外请注意 在定义了新的音符时值后 每个音符的时值都会成为后续所有音符的默认值 这与连续音符一样 -- 所以 刚刚的和弦中的两个C音符的时值都是2个全音符)

在Alda中 您还可以在和弦中使用[休止符](rests_zh_cn.md) 因为和弦后的下一个音符事件将在和弦中最短的音符/休止符之后开始 这个特性对于编写与和弦交织在一起的旋律很有用 例如 `c1/e/g/r4 b e g`

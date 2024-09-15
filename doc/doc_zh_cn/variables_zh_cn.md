# 变量(Variables)

*本页面翻译自[Variables.md](../variables.md)*

反复是音乐的一个重要部分 除了对音符和乐句做反复 还需要对较大的乐句或乐章做反复 如果只使用Alda的[反复](./repeats_zh_cn.md)语法 可能会很麻烦 为了获得更大的灵活性 并使代码更清晰 您可以用**变量**为[序列](./sequences_zh_cn.md)命名

## 定义变量

您可以使用`variableName = events go here`这样的语法来定义变量 例如:

```alda
motif = b-8 a g f e g a4
```

在这个代码中 定义了一个名为`motif`的变量 之后可以在乐谱中任何地方使用 这里面存放了`[b-8 a g f e g a4]`这一事件序列

您也可以将多行代码放在一个变量中 只需要将它们写在一对中括号`[]`中 这称之为多行事件序列:

```alda
motif = [
b-8 a g f
e g a4
]
```

## 使用变量

在定义变量后 我们可以使用这个变量 只需在乐器部分写出变量的名称即可:

```alda
piano:
o2 motif < motif d1
```

变量也可以像事件序列一样做反复:

```alda
piano:
motif *3
```

## 变量作为别名

严格来说 变量不一定是事件序列 它也可以是单个的事件 比如可以给特定的[属性](./attributes_zh_cn.md)创建别名 这样会非常方便:

```alda
quiet = (vol 25)
loud = (vol 50)
louder = (vol 75)

notes = c d e

piano:
quiet notes
Loud notes
Louder notes
```

## 变量嵌套

变量中可以包含先前定义的其他变量 这样您就可以从较小的组件写起 模块化地构建乐谱

```alda
notes = c d e
moreNotes = f g a b
lastOne = > c

cMajorScale = notes moreNotes lastOne

piano:
cMajorScale
```

### 命名规则

变量的名称有以下这些规则:

* 至少有2个字符长
* 前两个字符必须是字母(大小写均可)
* 在前两个字符之后 可以包含以下任意组合:
  * 大小写字母
  * 数字
  * 以下任意符号: `_ - + ' ( )`


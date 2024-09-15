# 乐谱与部分(Scores and Parts)

*此页面翻译自[scores-and-parts.md](../scores-and-parts.md)*

用Alda编写的乐曲的最高级别是**乐谱(score)** 乐谱由一个或多个乐器**部分(part)**组成 每个部分都有自己的[音符](notes_zh_cn.md)事件 这些事件会同时发生

Alda的设计很灵活 您可以灵活地组织乐谱 对于同一首曲子 作曲家可以从头到尾编写每个乐器部分的音符(例1) 或者交替地编写各乐器部分(例2)

> 注意 Alda中可用的乐器可以在此[乐器列表](list-of-instruments_zh_cn.md)查看

**例1**

```alda
trumpet:
o4 c d e f g a b > c d e f g a b > c

trombone:
o2 e f g a b > c d e f g a b > c d e
```

**例2**

```alda
trumpet: o4 c d e f g a b > c
trombone: o2 e f g a b > c d e

trumpet: d e f g a b > c
trombone: f g a b > c d e
```

在底层 Alda会按顺序处理乐谱 跟踪每种乐器的信息 包括音量 节奏 时值 偏移量和八度组 在您切换另一种乐器又切换回来后 您不必手动将各属性改成您上次使用该乐器时的状态--Alda会为您跟踪这些属性

## 乐器组

在Alda中 可以通过对多个乐器进行分组来使它们使用相同的音符事件 例如

```alda
trumpet/trombone: c d e f g f e d c
```

需要注意的是 Alda仍然分别跟踪每种乐器的各属性 这意味着作曲家需要确保各乐器的演奏同步 在例3中 trumpet在乐谱开头重复演奏一些D音符 然后演奏上升的D小调音阶 trombone也演奏D小调音阶 但是它是从乐谱的开头开始的 所以它会先演奏上升的音阶而不是与trumpet同步

例4中演示了一种解决方法 使两种乐器同时演奏D音阶 例5中演示了另一种方法--使用[标记](markers_zh_cn.md)实现相同的效果

**例3**

```alda
trumpet: d d d d d d d d

# 不同步 trombone更早开始
trumpet/trombone: d e f g a b- > c d
```

**例4**

```alda
trumpet: d d d d d d d d
trombone: r r r r r r r r # (休止8拍)

# 同步
trumpet/trombone: d e f g a b- > c d
```

**例5**

```alda
trumpet:
d d d d d d d d %scaleTime

trumpet/trombone:
@scaleTime d e f g a b- > c d
```

在乐器组中 Alda不会强制各乐器部分同步 作曲家可以自由地尝试用多种乐器以不同的方式演奏不同的音符 例如 您可以为各乐器指定不同的节奏和/或不同的音符时值 然后让它们演奏相同的音符:

**例6**

```alda
violin: (tempo 100)
viola: (tempo 112)
cello: (tempo 124)

violin/viola/cello: e f g e f g e f g e f g e f g
```

## 别名

现在我们已经知道如何使用不同类型的乐器 但是我们想要*同时使用同一种乐器*怎么办? 假如我们在用两个双簧管写一个曲子 显然不能将它们都称为"oboe" 在这样的情况下 就可以用**别名(aliases)**来区分它们

您可以将别名放在乐器名后面的双引号中 来为乐器赋予别名

```alda
oboe "oboe-1":
  c8 d e f g2
```

现在`oboe-1`指的是第一个双簧管 如果要编排它的音符 我们要使用`oboe-1` *而不是*`oboe`

现在可以用`oboe`创建第二个双簧管乐器:

```alda
oboe "oboe-2":
  e8 f g a b2
```

您也可以为乐器组设置别名:

```alda
oboe-1/oboe-2 "oboes":
  > c1
```

当为乐器组设置别名时 可以通过点(`.`)运算符来访问到单个乐器

当您创建了一组未命名的实例 并在后面想要使用单个乐器时 这会很有用:

```alda
violin/viola/cello "strings": g1~1~1
strings.cello: < c1~1~1
```

### 命名规则

乐器的名称和别名有以下这些规则:

* 至少有2个字符长
* 前两个字符必须是字母(大小写均可)
* 在前两个字符之后 可以包含以下任意组合:
  * 大小写字母
  * 数字
  * 以下任意符号: `_ - + ' ( )`

### 如何分配实例

Alda创建和分配乐器实例的细节很[复杂](instance-and-group-assignment_zh_cn.md) 但在实际操作中 为了避免错误 您可以遵循以下的简单的规则以避免错误:

- 如果您为乐器的一个实例分配了别名 则在当前乐谱中 该乐器的其他任何实例都必须具有别名

  ```alda
  # ERROR
  piano "foo": c8 d e f g2
  piano: e8 f g a b2

  # ERROR
  piano: c8 d e f g2
  piano "bar": e8 f g a b2

  # OK
  piano "foo": c8 d e f g2
  piano "bar": e8 f g a b2
  ```

- 一旦实例拥有了别名 您就不能再给它赋予新的别名

  ```alda
  # ERROR
  piano "foo": c8 d e f g
  foo "bar": a b > c

  # OK
  piano "foo": c8 d e f g
  foo: a b > c
  ```

- 您不能将别名重新分配给另一个实例

  ```alda
  # ERROR
  piano "foo": c8 d e f g2
  clarinet "foo": e8 f g a b2

  # OK
  piano "foo": c8 d e f g2
  clarinet "bar": e8 f g a b2
  ```

- 创建一个乐器组时 组中的成员必须是: a) 乐器的新实例 或 b)_带有别名_的现有乐器

  不允许两者混用 因为这会导致Alda不清楚应该用哪个乐器实例

  ```alda
  # ERROR
  piano "foo": c8 d e
  foo/trumpet: g1

  # OK
  piano "foo": c8 d e
  trumpet "bar": r8~8~8
  foo/bar: g1

  # OK
  piano/trumpet: c8 d e g1
  ```


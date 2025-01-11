# 实例与组的分配(Instance and Group Assignment)

*此文档翻译自[Instance and Group Assignment](../instance-and-group-assignment.md)*

> 注意: 这是关于Alda中实例和组分配的详细说明 适合对此感兴趣的用户 以下内容较为复杂 但可以简化为[几个简单规则](scores-and-parts_zh_cn.md#如何分配实例)
> 如果您不关心具体细节 可以直接点击链接查看简化版 

Alda中有四种**乐器调用**的方式 每种方式可以**有别名**或**没有别名** 也可以**有多个实例**或**没有多个实例**

### `foo:`

- 如果`foo`是一个已命名的乐器或组 比如`piano-1:`...
  - 则表示该乐器或组

- 如果`foo`是一个预设乐器 比如`piano:`...
  - 如果乐谱中还没有`piano` 则：
    - 创建新的`piano`实例
    - 后续的`piano`调用将引用这个实例
  - 如果乐谱中已经有一个命名的`piano` 则会报错(多个`piano`实例必须命名)
  - 如果乐谱中已有一个未命名的`piano`实例 调用将引用该实例

- 否则 会报"未识别的乐器"错误

### `foo "bar":`

- `foo`必须是一个预设乐器 否则会报错
- 如果`"bar"`已被用作其他实例的别名 会报错
- 如果乐谱中已有未命名的`foo`实例 也会报错(`foo`的所有实例必须命名)
- 创建一个名为`bar`的`foo`实例 如:
  - `piano "larry":` 创建一个名为`"larry"`的`piano`实例
  - 后续乐谱中 `larry:` 将引用该实例

### `foo/bar:`

- 如果`foo`和`bar`是同一命名实例 比如`foo/foo` 会报错
- 如果`foo`和`bar`是同一预设乐器 比如`piano/piano` 也会报错
- 如果`foo`和`bar`都引用之前命名的乐器实例 表示这些实例
- 如果`foo`和`bar`是预设乐器 如`piano/bassoon:` 则按`foo:`的规则选择或创建实例

### `foo/bar "baz":`

- 如果`foo`和`bar`是同一命名实例 比如`foo/foo` 会报错 
- 如果`foo`和`bar`是同一预设乐器 如`piano/piano` 会报错 必须分别命名
- 如果`foo`和`bar`是命名的乐器实例 表示这些实例 并创建一个名为`"baz"`的别名
- 如果`foo`和`bar`是预设乐器 如`piano/guitar "floop":` 则创建新实例 并为组创建别名`"floop"`


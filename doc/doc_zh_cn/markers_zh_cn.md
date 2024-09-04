# Markers

**标记(Markers)**可以放置在乐谱中的任何位置 也可以放在任何乐器部分 例如 `%chorus`将在当前偏移量处放置一个名为"chorus"的标记 然后在任何位置使用`@chorus`会将当前[偏移量](offset_zh_cn.md)设置为"chorus"标记处的偏移量

标记在放置前不能被引用 -- 例如 以下的乐谱将导致错误

```alda
piano:
  @someMarkerThatDoesntExistYet
  c8 d e f g2

guitar:
  r1
  %someMarkerThatDoesntExistYet
```

在引用标记之前 它必须被放置:

```alda
guitar:
  r1
  %existingMarker

piano:
  @existingMarker
  c8 d e f g2
```

## 合法的标记名

标记的命名规则与[乐器名](scores-and-parts_zh_cn.md#合法的名称)相同的规则


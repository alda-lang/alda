# 以编程的方式编写音乐

*此页面翻译自[Writing Music Programmatically](../writing-music-programmatically.md)*

Alda的设计旨在更易于学习和使用 对于没有或很少编程经验的用户来说也是如此 所以该语言省略了大多数编程语言中的复杂功能 如函数 类和数据类型

例如 使用原声吉他用八分音符演奏C大调音阶 从C3开始_(即第3个八度组中的C)_ 其书写方式如下:

```alda
acoustic-guitar:
  o3 c8 d e f g a b > c
```

同时 Alda的一个重要目标是促进以编程的方式创作音乐 仅仅在Alda语言中 您就能获得很多好处 但是如果您恰好有一点编程知识 您就可以将Alda视作算法作曲和实时编码的一个**平台** 而不只是一种**语言**

## 什么是算法作曲

简单来说 [算法作曲](https://zh.wikipedia.org/wiki/%E7%AE%97%E6%B3%95%E4%BD%9C%E6%9B%B2)就是一种创作音乐的方式 您可以将部分创作过程留给随机的机会 哪些部分留给"机会" 以及"机会"的部分究竟如何实现 都是完全开放的

例如 您可以写一首曲子 想出一组节奏和音调 然后写一个程序 将节奏和音调随机组合在一起 创作出一首音乐

另一个例子: 您可以写一个程序 获取您当地10天的天气预报 并使用任意的规则将其转换成音乐 比如 如果温度是奇数 可能会使用大调音阶 或者如果天气预报说会下雪 其中一种乐器将是大提琴 等等...

可能性无穷无尽!

## 什么是实时编码(live coding)

[live coding](https://en.wikipedia.org/wiki/live_coding)是在现场表演环境中通过编程即时创作音乐(或其他类型艺术)的实践行为
*译者注: 本维基词条没有中文版*

通常 这种表演包括程序员屏幕的演示 以便观众可以看到正在编写和评估的代码 并实时观察正在创作的作品

## Alda作为算法创作的实时编码的平台

如上所述 Alda语言没有提供函数和随机数生成器这种常用于编写算法作品的机制 然而 Alda由于其简单性 可以轻松地作为更复杂结构的基础

利用下面这些库 您可以通过另一种编程语言编写程序来生成Alda代码 当宿主语言提供交互模式时(例如在repl(read-evaluate-print loop)中) 这种技术尤其强大

> 如果您喜欢的语言没有在这里列出 并且您有兴趣使用它来生成Alda乐谱
> 请考虑使用该语言[编写您自己的Alda][write-your-own-alda-library]并将其添加到这个列表中

| 语言 | 库       | 作者         |
|----------|---------------|----------------|
| clojure  | [alda-clj]    | dave yarwood   |
| ruby     | [alda-rb]     | ulysses zhan   |
| julia    | [alda.jl]     | ismael venegas |
| python   | [alda-python] | nicola vitucci |

[alda-clj]: https://github.com/daveyarwood/alda-clj
[alda-rb]: https://github.com/ulysseszh/alda-rb
[alda.jl]: https://github.com/salchipapa/alda.jl
[alda-python]: https://github.com/nvitucci/alda-python
[write-your-own-alda-library]: implementing-an-alda-library_zh_cn.md


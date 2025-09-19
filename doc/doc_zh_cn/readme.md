<p align="center">
  <a href="http://alda.io">
    <img src="../../alda-logo-horizontal.svg"
         alt="alda logo"
         width=360
         height=128>
  </a>

  <p align="center">
  <b><a href="#安装">安装</a></b>
  |
  <b><a href="./index_zh_cn.md">文档</a></b>
  |
  <b><a href="../../CHANGELOG.md">更新日志</a></b>
  |
  <b><a href="#contributing">贡献</a></b>

  <br>
  <br>

  <a href="http://slack.alda.io">
    Slack上的Alda频道
  </a>
  </p>
</p>

Alda是一种基于文本的音乐编程语言 用于音乐创作 您可以在只使用文本编辑器和命令行的情况下用它来创作和播放音乐

```alda
piano:
  o3
  g8 a b > c d e f+ g | a b > c d e f+ g4
  g8 f+ e d c < b a g | f+ e d c < b a g4
  << g1/>g/>g/b/>d/g
```

> 有关更多示例 请参阅这些[示例乐谱](../../examples/)

该语言的设计同样注重美观 灵活性和易用性

(为什么用纯文本的方式创作音乐 而不使用基于GUI的编曲软件 请阅读[这篇博文][alda-blog-post] 了解简要的历史和理由)

[alda-blog-post]: https://blog.djy.io/alda-a-manifesto-and-gentle-introduction/

## 功能

* 易于理解 类似标记的语法
* 专为不懂编程的音乐家和不懂音乐的程序员设计
* 乐谱是纯文本文件 可以用`alda`命令行工具播放
* 提供[交互模式(REPL)](./rests_zh_cn.md) 您可以实时输入Alda代码并听到结果
* 支持[以编程方式编写音乐](./writing-music-programmatically_zh_cn.md)(用于算法作曲 实时编码等)
* 使用[通用MIDI音效集][gm-sound-set]中的乐器创建MIDI音乐

[gm-sound-set]: http://www.midi.org/techspecs/gm1sound.php

### 计划

> 如果您愿意提供帮助 请加入 -- [the water's fine](#贡献)!

* [在浏览器中运行Alda](https://github.com/alda-lang/alda/discussions/455)
* [定制和使用波形合成乐器](https://github.com/alda-lang/alda/discussions/435)
* [导出为MusicXML](https://github.com/alda-lang/alda/discussions/424)以供其他编曲软件编辑
* [改进对树莓派的支持](https://github.com/alda-lang/alda/discussions/456)

## 安装

请查阅[官网][alda-install]以获取安装说明 安装最新版本

[alda-install]: https://alda.io/install

## Demo

可用命令和选项的概述:

    alda --help

播放包含Alda代码的文件:

    alda play --file examples/bach_cello_suite_no_1.alda

在命令行播放任意代码:

    alda play --code "piano: c6 d12 e6 g12~4"

启动[交互模式](./alda-repl_zh_cn.md):

    alda repl

## 文档

您可以在[此处](./index_zh_cn.md)找到Alda的文档

## 贡献

我们非常乐意得到您的帮助 -- 欢迎提交pr!

想要了解我们正在讨论和开展的工作 请访问[Alda GitHub项目板][gh-project]

关于如何做贡献 请参阅[贡献文档](../../CONTRIBUTING.md)

> 另一种贡献方式是[赞助Dave][gh-sponsor] 参与Alda未来的开发

[gh-org]: https://github.com/alda-lang
[gh-project]: https://github.com/orgs/alda-lang/projects/1
[gh-sponsor]: https://github.com/sponsors/daveyarwood

## 支持 讨论与联系

**Slack**: 加入[Alda Slack group](http://slack.alda.io) 快捷轻松 来打个招呼吧!

**Reddit**: 订阅[/r/alda](https://www.reddit.com/r/alda/)在此讨论关于Alda的事情 还可以分享您的乐谱!

## 许可证

Copyright © 2012-2025 Dave Yarwood et al

Distributed under the Eclipse Public License version 2.0.

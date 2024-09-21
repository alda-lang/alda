# Alda 2 迁移指南

*此文档翻译自 [Alda 2 Migration Guide](../alda-2-migration-guide.md)*

*此文档由[cewno](https://github.com/cewno)机翻并校对 非人工翻译 故有极大可能会有错误(尤其是命令相关的文本 标点和字母大小写会出问题) 我(欧阳闻奕/OWALabuy)或许没有完全验证此文档中命令文本的正确性 在此段文字移除前 请各位自行验证输入的命令的准确性(参考原文档)*

Alda 2.0.0 于 2021 年 6 月发布。Alda 1 主要用 Clojure 编写（带有优化 Java 客户端，以实现更快的命令行交互），而 Alda 2 是用 Go 和 Kotlin 从头开始重写的

> 如果你好奇为什么 Dave 决定用 Go 和 Kotlin 重写 Alda，请阅读 [这篇说明][why-the-rewrite]！

Alda 2 基本上与 Alda 1 向后兼容，在大部分情况下，你用 Alda 1 编写的乐谱都应该与 Alda 2 兼容，并且听起来完全一样。虽然 Alda 的实现已经从头开始重写，但 Alda 的语法几乎保持不变

在 Alda 2 中，语言有一个重要的变化：**不再支持内联 Clojure 代码**。原因很明显：Alda 客户端现在是用 Go 编写的，所以我们不能像过去那样在 Alda score 中计算任意 Clojure 代码。(尽管如此，Alda 仍然是一个强大的算法合成工具！参见下面的“[Programmatic composition](#programmatic-composition)”)

下面是从 Alda 1 升级到 Alda 2 时应该注意的一些事项

## 无需再运行 `alda up` ！使用 `alda-player`

在Alda 1中 您在播放乐谱之前必须通过`alda up`命令启动Alda服务器

在 Alda 2 中，没有 Alda 服务器。你可以简单地运行一个命令，比如 `alda play -c "flute: o5 c8 < b16 a g f e d c2"`，而不需要先运行 `alda up`

有一个名为 `alda-player` 的新后台进程来处理音频播放。每次您播放乐谱时，Alda 都会自动为您启动一个。 您需要在 PATH 环境变量上包含可用的 `alda` 和 `alda-player` 才能正常工作

Alda CLI 将帮助确保您安装了相同版本的 `alda` 和 `alda-player`，如果它们来自不同的版本，它还会为您安装正确版本的`alda-player`

当您运行 `alda update` 时，会将 `alda` 和 `alda-player` 更新到最新版本

## 使用 `alda doctor` 更好地排除故障

`alda doctor` 是一个新命令，它运行一些基本的检查，并检查与 alda 有关的设置。如果一切顺利，您应该会看到如下输出：

```
OK  Parse source code
OK  Generate score model
OK  Find an open port
OK  Send and receive OSC messages
OK  Locate alda-player executable on PATH
OK  Check alda-player version
OK  Spawn a player process
OK  Ping player process
OK  Play score
OK  Export score as MIDI
OK  Locate player logs
OK  Player logs show the ping was received
OK  Shut down player process
OK  Spawn a player on an unknown port
OK  Discover the player
OK  Ping the player
OK  Shut the player down
OK  Start a REPL server
nREPL server started on port 36099 on host localhost - nrepl://localhost:36099
OK  Interact with the REPL server
```

如果您遇到意料之外的问题，`alda doctor` 的输出可以帮助您查明问题并帮助 Alda 的维护者查找和修复错误

## 全新改进的 `alda repl` (交互模式)

你在 Alda 1 中所了解和喜爱的 REPL (**R** read- **E**val-**P**lay **L** loop，一种来自 Lisp 传统的“读-求值-打印循环”的变体)经验在 Alda 2 中得到了保留。只需运行 `alda repl` 即可开始交互式repl会话。然后你可以尝试使用Alda代码，听听每一行输入的声音。(试着输入`midi-woodblock: c8. c c8 r c c`，看看会发生什么

就像以前一样，你可以输入 `:help` 来了解可用的 REPL 命令，然后通过输入 `:help play` 来了解更多关于命令的信息

那么，Alda 2 REPL 有什么新功能呢？我们新赋予 REPL 的一个强大功能是它可以在客户端或服务器模式下运行。默认情况下，`alda repl` 将同时启动服务器和客户端会话。但是如果你已经有了一个正在运行的 REPL 服务器(或者如果你的朋友有，在世界的其他地方…:bulb:)，你可以通过运行 `alda repl --client --host example.com --port 12345` (或者更短的版本: `alda repl -c -H example.com -p 12345`)来连接它。这可能会带来很多乐趣，因为多个客户端可以连接到同一个 REPL 服务器并实时协作！

> 如果您对 Alda 新 super-REPL 背后的技术细节感兴趣，
> 查看 Dave 的博客文章，[Alda 和 nREPL 协议][alda-nrepl]

自 Alda 1 以来更改一些与 REPL 相关的内容：

* 服务器 / 工作进程管理命令不再存在，因为不再需要管理服务器和工作进程！已删除以下命令：
  * `:down`
  * `:downup`
  * `:list`
  * `:status`
  * `:up`

* 与打印乐谱信息相关的命令已重命名，以便使用：
  * (v1) `:score` => (v2) `:score text` 或 `:score`
  * (v1) `:info` => (v2) `:score info`
  * (v1) `:map` => (v2) `:score data`
  * (仅在 v2 中可用) `:score events`

## 在某些情况下，属性语法已更改

你可能没有意识到这一点，但在 Alda 1 中，像 `(volume 42)` 这样的属性实际上是在运行时转换的 Clojure 函数调用。事实上，整个 Clojure 语言都可以在 Alda scores 中使用。例如，您可以生成一个介于 0 和 100 之间的随机数，并使用 `(volume (rand-int 100))` 将音量设置为该值

在 Alda 2 中，你不能再做这种事情了，因为 Alda 不再用 Clojure 编写的。（但是，如果您对做这种事情感兴趣，您不必担心，因为您仍然可以这么做！请参阅下面的“[Programmatic composition](#programmatic-composition)”

Clojure 是一种 [Lisp][lisp] 编程语言。如果您不知道这是什么，这里有一个简单的解释：Lisp 语言的语法主要由括号组成。“S-expression”是括号内的元素列表，`(像 这个 列表)`。列表中的第一项是_operator_，其余项是_arguments_。s表达式是可嵌套的;例如，像`(1 + 2) * (3 + 4)`这样的算术表达式在 Lisp 中写成:`(* (+ 1 2)(+ 3 4))`

Alda 2 包含一个简单的内置 Lisp 语言 (“Alda - Lisp”) ，它提供了足够的支持 Alda 的属性操作。但是它缺少 Clojure 的许多语法。Clojure 有多种您可能在 Alda scores 中看到的附加语法，包括 `:keywords`， `[vectors]` 和 `{hash maps}`。Alda-lisp 没有这些功能，所以如果使用 Clojure 的这些功能，一些 Alda scores 将无法在 Alda 2 中播放

以下属性受 Alda 2 中语法更改的影响：

<table>
  <thead>
    <tr>
      <th>Example</th>
      <th>Alda 1</th>
      <th>Alda 2</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Key signature</td>
      <td>
        <pre><code>(key-sig! "f+ c+ g+")
(key-sig! [:a :major])
(key-sig! {:f [:sharp] :c [:sharp] :g [:sharp]})</code></pre>
      </td>
      <td>
        <pre><code>(key-sig! "f+ c+ g+")
(key-sig! '(a major))
(key-sig! '(f (sharp) c (sharp) g (sharp)))</code></pre>
      </td>
    </tr>
    <tr>
      <td>Octave up/down</td>
      <td>
        <pre><code>>
<
(octave :up)
(octave :down)</code></pre>
      </td>
      <td>
        <pre><code>>
<
(octave 'up)
(octave 'down)</code></pre>
      </td>
    </tr>
  </tbody>
</table>

所有其他属性都应该正常工作，但如果您遇到现有 Alda 1 的乐谱的任何其他向后兼容性问题，请 [让我们知道][open-an-issue]！

## 乐谱默认音量

Alda 1的乐谱开头的默认音量是100 这对应MIDI力度127 也就是最大值 现在Alda 2中 可以用动态标记(如`mp`或`ff`这样的)来指定音量

在Alda 2中 所有乐谱开头默认的音量是`(mf)` 相当于`(vol 54)` 如果您之前用Alda 1时写的乐谱依赖于默认音量100 那么迁移到Alda2后需要在乐谱的开头显式地指定此属性

## Programmatic composition

Alda取消了内联Clojure代码的功能

但是，如果您有兴趣使用 Clojure 编写算法音乐，那么您很幸运！2018 年，Dave 创建了 [alda-clj]，这是一个 Clojure 库，用于使用 Alda 对音乐进行实时编码。该库提供了用于编写 Alda scores 的 Clojure DSL，该 DSL 等同于 Alda 1 中提供的 DSL

这是一个 [示例 scores][entropy]，它展示了 Clojure 程序员如何使用 alda-clj 来创作算法音乐

## `alda parse` 输出

`alda parse` 命令解析 Alda scores 并生成表示 scores 数据的 JSON 输出。这对于调试目的或在 Alda 上构建工具非常有用

Alda 2 中 `alda parse` 的输出与 Alda 1 的输出在许多方面不同。例如，以下是 Alda 1 中运行 `alda parse -c "guitar: e" -o events` 的输出：

```json
[
  {
    "event-type": "part",
    "instrument-call": {
      "names": [
        "guitar"
      ]
    },
    "events": null
  },
  {
    "event-type": "note",
    "letter": "e",
    "accidentals": [],
    "midi-note": null,
    "beats": null,
    "ms": null,
    "slur?": null
  }
]
```

在 Alda 2 中：

```json
[
  {
    "type": "part-declaration",
    "value": {
      "names": [
        "guitar"
      ]
    }
  },
  {
    "type": "note",
    "value": {
      "pitch": {
        "accidentals": [],
        "letter": "E"
      }
    }
  }
]
```

如您所见，Alda 1 和 Alda 2 以不同的方式呈现相同的信息！如果您碰巧构建了任何依赖于 Alda 1 `alda parse` 输出的工具或工作流，则可能需要在升级到 Alda 2 后进行调整

## 就是这样！

我们希望您喜欢 Alda 2！请随时加入我们的 [Slack 群组][alda-slack] 并让我们知道您的想法。如果你遇到 bug 或任何其他类型的奇怪行为，你也可以 [打开一个 issue][open-an-issue]，我们很乐意帮助您解决问题！

[why-the-rewrite]: https://blog.djy.io/why-im-rewriting-alda-in-go-and-kotlin/
[lisp]: https://en.wikipedia.org/wiki/Lisp_(programming_language)
[alda-clj]: https://github.com/daveyarwood/alda-clj
[entropy]: https://github.com/daveyarwood/alda-clj/blob/master/examples/entropy
[alda-nrepl]: https://blog.djy.io/alda-and-the-nrepl-protocol/
[open-an-issue]: https://github.com/alda-lang/alda/issues/new/choose
[alda-slack]: http://slack.alda.io


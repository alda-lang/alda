# 故障排除

*此页面翻译自[troubleshooting.md](../troubleshooting.md)*

## 通用

### `alda doctor`

如果您的Alda有任何问题 可以运行`alda docter`来进行快速的错误排查

```
$ alda doctor
OK  Parse source code
OK  Generate score model
OK  Ensure that there are no stale player processes
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
nREPL server started on port 39879 on host 127.0.0.1 - nrepl://127.0.0.1:39879
OK  Find the REPL server
OK  Interact with the REPL server
OK  Shut down the REPL server
```

这可以帮助您定位到具体问题的范围

### 播放器进程

Normally, the `alda` client operates by launching `alda-player` processes automatically in the background. These processes are designed to be short-lived, handling audio playback and eventually shutting themselves down after a period of inactivity.
通常 `alda`客户端会自动在后台启动`alda-player`进程 这些进程是临时的 它们会处理音频播放 并在一段时间没有活动后自动关闭

您可以使用`alda ps`列出当前正在运行的所有播放器进程的信息

```
$ alda ps
id      port    state   expiry  type    pid
aar     42175   ready   6 minutes from now      player  387371
hgf     40093   ready   7 minutes from now      player  387370
oau     44559   ready   8 minutes from now      player  387372
```

每个播放器进程都会将日志消息保存到文件中 如果您在播放时遇到问题 可以查看这些日志文件来排查故障 要查找这些log文件的位置 可以运行`alda-player info`:

```
$ alda-player info
alda-player 2.3.0
log path: /home/dave/.cache/alda/logs
```

如果您遇到问题 还可以尝试在前台运行播放器进程 并使用`-vv`参数进行"非常详细"的日志记录

```
$ alda-player -vv run -p 27705
ykt INFO  2024-07-12 13:10:43 StateManager.cleanUpStaleStateFiles:88 - Cleaning up stale files in /home/dave/.cache/alda/state/players...
ykt INFO  2024-07-12 13:10:43 StateManager.cleanUpStaleStateFiles:88 - Cleaning up stale files in /home/dave/.cache/alda/state/repl-servers...
ykt INFO  2024-07-12 13:10:43 Main.run:77 - Starting receiver, listening on port 27705...
ykt INFO  2024-07-12 13:10:43 MidiEngine.info:242 - [0] Initializing MIDI sequencer...
ykt INFO  2024-07-12 13:10:43 MidiEngine.info:242 - [0] Initializing MIDI synthesizer...
ykt INFO  2024-07-12 13:10:47 MidiEngine.info:242 - [0] Player ready
```

然后再开一个终端 使用`alda play`命令指定端口号来使用您刚刚运行的播放器进程(在上文中 我们将播放器运行在 27705 端口上)

```
$ alda play -p 27705 -c 'piano: c d e f g'
Playing...
```

您应该能在运行播放器的进程的终端中看到日志

## IcedTea相关问题

[IcedTea]是OpenJDK的一个变体 随一些Linux发行版(如Fedora Gentoo Debian等)一起发行 如果您使用的发行版中有IcedTea 就可能会遇到Alda无法播放的问题

一位用户在播放器的日志中看到了这个错误

```
Exception in thread "main" java.lang.UnsatisfiedLinkError: no icedtea-sound in java.library.path
```

如果您也遇到同样的问题 可以运行`apt-cache search icedtea`来查找IcedTea相关的软件包 然后运行`sudo apt install icedtea-netx libpulse-java libpulse-jni`来安装缺失的软件包 这应该能解决您的问题

## Windows

### RegCreatKeyEx 错误

当在Windows的终端中运行`alda`时 您可能会遇到这样的错误:

```
WARNING: Could not open/create prefs root node Software\JavaSoft\Prefs at root 0x80000002.
Windows RegCreateKeyEx(...) returned error code 5.
```

这个错误是因为Windows缺少`JavaSoft\Prefs`注册表项 添加上就可以了:

1. 按win+r 输入`regedit`启动Windows注册表编辑器
2. 展开`HKEY_LOCAL_MACHINE` 并找到`Software` 键
3. 在`Software`中 找到`JavaSoft`键

右击"JavaSoft" 选择"新建" -> "键" 将此键命名为`Prefs` 按下回车即可

[IcedTea]: https://openjdk.org/projects/icedtea/


# Alda REPL(Alda交互模式)

*此文档翻译自[Alda REPL](../alda-repl.md)*

Alda带有一个交互模式 -- REPL(**R**ead-**E**val-**P**lay **L**oop) 您可以用它尝试语法 在交互模式中输入每行代码后 您都能听到结果

要启动Alda REPL 请运行:

```bash
alda repl
```

在交互模式中 您可以输入Alda代码 也可以使用内置命令 这些命令以冒号`:`开头

要获得可用的命令列表 请输入`:help`

## REPL客户端和服务器

当您运行`alda repl`时 它实际上会启用REPL服务器和REPL客户端会话 客户端将您的输入发送到服务器(在本例中 服务器刚好在同一进程中运行) 并根据服务器的响应打印输出

您还可以独立启动Alda REPL客户端和服务器

```bash
# 仅启动服务器 在随机可用端口上运行
alda repl --server

# 仅启动服务器 在12345端口上运行
alda repl --server --port 12345

# 仅启动客户端 与12345端口上运行的服务器通信
alda repl --client --port 12345
```

您甚至还可以在另一台设备上运行REPL服务器 并通过指定主机或IP连接到它

```bash
# 在另一台计算机上:
alda repl --server --port 12345

# 启动客户端 与在另一台计算机上运行的Alda REPL服务器通信
# (为便于举例 设那台主机的IP地址为11.22.33.44)
alda repl --client --host 11.22.33.44 --port 12345
```


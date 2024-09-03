# 安装优质的soundfont(音色库)

*此文档翻译自[installing-a-good-soundfont.md](../installing-a-good-soundfont.md)

默认的JVM音色库通常音质较差 我们建议安装免费的音色库 如FluidR3 使您的MIDI音色更悦耳

各种流行的免费音色库 包括FluidR3 均可从[这里](https://musescore.org/en/handbook/soundfonts#list)下载

## Mac / Linux

For your convenience, there is a script in this repo that will install the
FluidR3 soundfont for Mac and Linux users.
为了便于操作 此仓库中有一个shell脚本 用于为MacOS和Linux用户安装FluidR3音色库

获取脚本的副本

```bash
curl \
  https://raw.githubusercontent.com/alda-lang/alda/master/scripts/install-fluidr3 \
  -o /tmp/install-fluidr3

chmod +x /tmp/install-fluidr3
```

如果您愿意 可以在运行脚本之前检查一下:

```bash
/tmp/install-fluidr3
```

这会下载FluidR3 并用它替换`~/.gervill/soundbank-emg.sf2`(您的JVM的默认音色库)

## Windows

<img src="windows_jre_soundfont.png"
     alt="在Windows上替换默认的音色库">

要在Windows系统上替换默认的音色库:

1. 找到Java运行时环境(JRE)的目录 并导航到`lib`文件夹
    * 如果您安装了JDK8或更早的版本 请找到JDK的目录并导航到`jre\lib`文件夹
2. 创建一个名为`audio`的新文件夹
3. 将任何".sf2"格式的文件复制到此文件夹中

# 乐器列表

*此页面翻译自[List of Instruments](../list-of-instruments.md)*

目前Alda只支持GM(通用MIDI)的乐器 未来我们计划添加[波形合成](https://github.com/alda-lang/alda/issues/100) 那时就能使用正弦波/方波/三角波/锯齿波以及由波形构建的复杂合成器

下列所有乐器的名称或别名都可以在Alda乐谱中使用 例如:

```alda
midi-harpsichord: c8 d e f g a b > c
```

## MIDI乐器

这些乐器直接对应于[通用MIDI音色集](http://www.midi.org/techspecs/gm1sound.php)中的乐器 它们根据MIDI规范按音色预设分组

乐器名称后面的括号中存放了其别名

> 请注意 某些别名可能在将来被非MIDI乐器(如采样或波形乐器)取代 为了确保您的乐谱始终使用特定的MIDI乐器 您可以使用以`midi-`为前缀的乐器名称

### 钢琴

* midi-acoustic-grand-piano (midi-piano, piano)
* midi-bright-acoustic-piano
* midi-electric-grand-piano
* midi-honky-tonk-piano
* midi-electric-piano-1
* midi-electric-piano-2
* midi-harpsichord (harpsichord)
* midi-clavi (midi-clavinet, clavinet)

### 有音调的打击乐

* midi-celesta (celesta, celeste, midi-celeste)
* midi-glockenspiel (glockenspiel)
* midi-music-box (music-box)
* midi-vibraphone (vibraphone, vibes, midi-vibes)
* midi-marimba (marimba)
* midi-xylophone (xylophone)
* midi-tubular-bells (tubular-bells)
* midi-dulcimer (dulcimer)

### 风琴

* midi-drawbar-organ
* midi-percussive-organ
* midi-rock-organ
* midi-church-organ (organ)
* midi-reed-organ
* midi-accordion (accordion)
* midi-harmonica (harmonica)
* midi-tango-accordion

### 吉他

* midi-acoustic-guitar-nylon (midi-acoustic-guitar, acoustic-guitar, guitar)
* midi-acoustic-guitar-steel
* midi-electric-guitar-jazz
* midi-electric-guitar-clean (electric-guitar-clean)
* midi-electric-guitar-palm-muted
* midi-electric-guitar-overdrive (electric-guitar-overdrive)
* midi-electric-guitar-distorted (electric-guitar-distorted)
* midi-electric-guitar-harmonics (electric-guitar-harmonics)

### 贝斯

* midi-acoustic-bass (acoustic-bass, upright-bass)
* midi-electric-bass-finger (electric-bass-finger, electric-bass)
* midi-electric-bass-pick (electric-bass-pick)
* midi-fretless-bass (fretless-bass)
* midi-bass-slap
* midi-bass-pop
* midi-synth-bass-1
* midi-synth-bass-2

### 弦乐(及定音鼓)

* midi-violin (violin)
* midi-viola (viola)
* midi-cello (cello)
* midi-contrabass (string-bass, arco-bass, double-bass, contrabass, midi-string-bass, midi-arco-bass, midi-double-bass)
* midi-tremolo-strings
* midi-pizzicato-strings
* midi-orchestral-harp (harp, orchestral-harp, midi-harp)
* midi-timpani (timpani)

### 合奏

* midi-string-ensemble-1
* midi-string-ensemble-2
* midi-synth-strings-1
* midi-synth-strings-2
* midi-choir-aahs
* midi-voice-oohs
* midi-synth-voice
* midi-orchestra-hit

### 铜管乐器

* midi-trumpet (trumpet)
* midi-trombone (trombone
* midi-tuba (tuba)
* midi-muted-trumpet
* midi-french-horn (french-horn)
* midi-brass-section
* midi-synth-brass-1
* midi-synth-brass-2

### 簧片
* midi-soprano-saxophone (midi-soprano-sax, soprano-saxophone, soprano-sax)
* midi-alto-saxophone (midi-alto-sax, alto-saxophone, alto-sax)
* midi-tenor-saxophone (midi-tenor-sax, tenor-saxophone, tenor-sax)
* midi-baritone-saxophone (midi-baritone-sax, midi-bari-sax, baritone-saxophone, baritone-sax, bari-sax)
* midi-oboe (oboe)
* midi-english-horn (english-horn)
* midi-bassoon (bassoon)
* midi-clarinet (clarinet)

### 吹管乐器

* midi-piccolo (piccolo)
* midi-flute (flute)
* midi-recorder (recorder)
* midi-pan-flute (pan-flute)
* midi-bottle (bottle)
* midi-shakuhachi (shakuhachi)
* midi-whistle (whistle)
* midi-ocarina (ocarina)

### 合成器主音

* midi-square-lead (square, square-wave, square-lead, midi-square, midi-square-wave)
* midi-saw-wave (sawtooth, saw-wave, saw-lead, midi-sawtooth, midi-saw-lead)
* midi-calliope-lead (calliope-lead, calliope, midi-calliope)
* midi-chiffer-lead (chiffer-lead, chiffer, chiff, midi-chiffer, midi-chiff)
* midi-charang (charang)
* midi-solo-vox
* midi-fifths (midi-sawtooth-fifths)
* midi-bass-and-lead (midi-bass+lead)

### 合成器音色垫

* midi-synth-pad-new-age (midi-pad-new-age, midi-new-age-pad)
* midi-synth-pad-warm (midi-pad-warm, midi-warm-pad)
* midi-synth-pad-polysynth (midi-pad-polysynth, midi-polysynth-pad)
* midi-synth-pad-choir (midi-pad-choir, midi-choir-pad)
* midi-synth-pad-bowed (midi-pad-bowed, midi-bowed-pad, midi-pad-bowed-glass, midi-bowed-glass-pad)
* midi-synth-pad-metallic (midi-pad-metallic, midi-metallic-pad, midi-pad-metal, midi-metal-pad)
* midi-synth-pad-halo (midi-pad-halo, midi-halo-pad)
* midi-synth-pad-sweep (midi-pad-sweep, midi-sweep-pad)

### 合成器效果

* midi-fx-rain (midi-fx-ice-rain, midi-rain, midi-ice-rain)
* midi-fx-soundtrack (midi-soundtrack)
* midi-fx-crystal (midi-crystal)
* midi-fx-atmosphere (midi-atmosphere)
* midi-fx-brightness (midi-brightness)
* midi-fx-goblins (midi-fx-goblin, midi-goblins, midi-goblin)
* midi-fx-echoes (midi-fx-echo-drops, midi-echoes, midi-echo-drops)
* midi-fx-sci-fi (midi-sci-fi)

### 民族风乐器

* midi-sitar (sitar)
* midi-banjo (banjo)
* midi-shamisen (shamisen)
* midi-koto (koto)
* midi-kalimba (kalimba)
* midi-bagpipes (bagpipes)
* midi-fiddle
* midi-shehnai (shehnai, shahnai, shenai, shanai, midi-shahnai, midi-shenai, midi-shanai)

### 打击乐

* midi-tinkle-bell (midi-tinker-bell)
* midi-agogo
* midi-steel-drums (midi-steel-drum, steel-drums, steel-drum)
* midi-woodblock
* midi-taiko-drum
* midi-melodic-tom
* midi-synth-drum
* midi-reverse-cymbal

### 音效

* midi-guitar-fret-noise
* midi-breath-noise
* midi-seashore
* midi-bird-tweet
* midi-telephone-ring
* midi-helicopter
* midi-applause
* midi-gunshot (midi-gun-shot)

### 打击乐(架子鼓音色)

有一种特殊的`midi-percussion`乐器(别名`percussion`) 它提供各种打击乐的音色 每种音色映射到不同的音符 每个音符对应一个独立的打击乐器 但声音的音调与音符无关

有关打击乐的更多信息 请参阅[此处](https://en.wikipedia.org/wiki/General_MIDI#Percussion)(*译者注: 此维基百科条目的中文版本没有打击乐与钢琴键盘的对照图 故此链接是英文的条目*)

鼓谱示例:

```alda
(tempo! 150)

midi-percussion:
  V1: # 底鼓和军鼓
    o2 c4 e8 c r c e c
  V2: # 镲片(踩镲 吊镲 丁丁镲 另一个吊镲)
    o2 f+8 f+ r o3 c+8~8 f16 f r8 a
```



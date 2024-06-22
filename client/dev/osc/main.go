package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/daveyarwood/go-osc/osc"
)

func bundle(msgs ...*osc.Message) *osc.Bundle {
	bundle := osc.NewBundle(time.Now())

	for _, msg := range msgs {
		bundle.Append(msg)
	}

	return bundle
}

func systemShutdownMsg(offset int32) *osc.Message {
	return osc.NewMessage("/system/shutdown", offset)
}

func systemPlayMsg() *osc.Message {
	return osc.NewMessage("/system/play")
}

func systemStopMsg() *osc.Message {
	return osc.NewMessage("/system/stop")
}

func systemClearMsg() *osc.Message {
	return osc.NewMessage("/system/clear")
}

func clearMsg(track int) *osc.Message {
	return osc.NewMessage(fmt.Sprintf("/track/%d/clear", track))
}

func exportMsg(filepath string) *osc.Message {
	return osc.NewMessage("/system/midi/export", filepath)
}

func midiPatchMsg(track int, channel int, offset int, patch int) *osc.Message {
	return osc.NewMessage(
		fmt.Sprintf("/track/%d/midi/patch", track),
		int32(channel),
		int32(offset),
		int32(patch),
	)
}

func finishLoopMsg(track int) *osc.Message {
	return osc.NewMessage(
		fmt.Sprintf("/track/%d/finish-loop", track),
		int32(0),
	)
}

func midiNoteMsg(
	track int, channel int, offset int, note int, duration int,
	audibleDuration int, velocity int,
) *osc.Message {
	return osc.NewMessage(
		fmt.Sprintf("/track/%d/midi/note", track),
		int32(channel),
		int32(offset),
		int32(note),
		int32(duration),
		int32(audibleDuration),
		int32(velocity),
	)
}

func patternMsg(track int, channel int, offset int, pattern string, times int) *osc.Message {
	return osc.NewMessage(
		fmt.Sprintf("/track/%d/pattern", track),
		int32(channel),
		int32(offset),
		pattern,
		int32(times),
	)
}

func patternMidiNoteMsg(
	pattern string, offset int, note int, duration int, audibleDuration int,
	velocity int) *osc.Message {
	return osc.NewMessage(
		fmt.Sprintf("/pattern/%s/midi/note", pattern),
		int32(offset),
		int32(note),
		int32(duration),
		int32(audibleDuration),
		int32(velocity),
	)
}

func patternClearMsg(pattern string) *osc.Message {
	return osc.NewMessage(fmt.Sprintf("/pattern/%s/clear", pattern))
}

func patternLoopMsg(track int, channel int, offset int, pattern string) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/pattern-loop", track))
	msg.Append(int32(channel))
	msg.Append(int32(offset))
	msg.Append(pattern)
	return msg
}

func oneNote() *osc.Bundle {
	return bundle(
		midiPatchMsg(1, 0, 0, 30),
		midiNoteMsg(1, 0, 0, 45, 1000, 1000, 127),
		systemPlayMsg(),
	)
}

func sixteenFastNotes() *osc.Bundle {
	interval := 100
	audibleDuration := 80
	noteNumber := 30 + rand.Intn(60)

	msgs := []*osc.Message{midiPatchMsg(1, 0, 0, 70)}

	for offset := 0; offset <= interval*16; offset += interval {
		msgs = append(
			msgs,
			midiNoteMsg(1, 0, offset, noteNumber, interval, audibleDuration, 127),
		)
	}

	msgs = append(msgs, systemPlayMsg())

	return bundle(msgs...)
}

func drumsDemo() *osc.Bundle {
	ff := 127
	mf := 100

	return bundle(
		// hi hat
		midiNoteMsg(1, 9, 0, 42, 250, 250, ff),
		midiNoteMsg(1, 9, 250, 42, 125, 125, mf),
		midiNoteMsg(1, 9, 375, 42, 250, 250, ff),
		midiNoteMsg(1, 9, 625, 42, 125, 125, mf),

		// kick
		midiNoteMsg(1, 9, 0, 36, 500, 500, ff),
		midiNoteMsg(1, 9, 625, 36, 125, 125, mf),
		// ... fill ...
		midiNoteMsg(1, 9, 1500, 36, 1500, 1500, ff),

		// snare
		midiNoteMsg(1, 9, 750, 38, 125, 125, ff),
		midiNoteMsg(1, 9, 875, 38, 125, 125, mf),

		// toms
		midiNoteMsg(1, 9, 1000, 41, 125, 125, mf),
		midiNoteMsg(1, 9, 1125, 41, 125, 125, ff),
		midiNoteMsg(1, 9, 1250, 43, 125, 125, mf),
		midiNoteMsg(1, 9, 1375, 43, 125, 125, mf),

		// crash
		midiNoteMsg(1, 9, 1500, 49, 1500, 1500, ff),

		systemPlayMsg(),
	)
}

func playPattern(pattern string, times int) *osc.Bundle {
	return bundle(
		patternClearMsg(pattern),
		patternMidiNoteMsg(pattern, 0, 57, 500, 500, 127),
		patternMidiNoteMsg(pattern, 500, 60, 500, 500, 127),
		patternMidiNoteMsg(pattern, 1000, 62, 500, 500, 127),
		patternMidiNoteMsg(pattern, 1500, 64, 500, 500, 127),
		patternMidiNoteMsg(pattern, 2000, 67, 500, 500, 127),
		midiPatchMsg(1, 0, 0, 60),
		patternMsg(1, 0, 0, pattern, times),
		systemPlayMsg(),
	)
}

func playPatternOnce(pattern string) *osc.Bundle {
	return playPattern(pattern, 1)
}

func playPatternTwice(pattern string) *osc.Bundle {
	return playPattern(pattern, 2)
}

func playPatternThrice(pattern string) *osc.Bundle {
	return playPattern(pattern, 3)
}

func randomPatternNotes(
	pattern string, quantity int, baseNote int,
) []*osc.Message {
	msgs := []*osc.Message{}

	interval := 500
	audibleDuration := 250

	for offset := 0; offset < interval*quantity; offset += interval {
		noteNumber := baseNote + rand.Intn(24)

		msgs = append(msgs, patternMidiNoteMsg(
			pattern, offset, noteNumber, interval, audibleDuration, 127,
		))
	}

	return msgs
}

func changePattern(pattern string) *osc.Bundle {
	msgs := []*osc.Message{patternClearMsg(pattern)}
	msgs = append(msgs, randomPatternNotes(pattern, 4, 30)...)
	return bundle(msgs...)
}

// Plays two loops concurrently, four times each, on the same track.
func twoFiniteLoops() *osc.Bundle {
	pattern1 := "pattern1"
	pattern2 := "pattern2"

	msgs := []*osc.Message{patternClearMsg(pattern1)}
	msgs = append(msgs, randomPatternNotes(pattern1, 4, 30)...)
	msgs = append(msgs, patternClearMsg(pattern2))
	msgs = append(msgs, randomPatternNotes(pattern2, 4, 60)...)
	msgs = append(msgs, []*osc.Message{
		midiPatchMsg(1, 0, 0, 0),
		patternMsg(1, 0, 0, pattern1, 4),
		patternMsg(1, 0, 250, pattern2, 4),
		systemPlayMsg(),
	}...)

	return bundle(msgs...)
}

// Plays two indefinite loops concurrently on the same track.
func twoInfiniteLoops() *osc.Bundle {
	pattern1 := "pattern1"
	pattern2 := "pattern2"

	msgs := []*osc.Message{patternClearMsg(pattern1)}
	msgs = append(msgs, randomPatternNotes(pattern1, 4, 30)...)
	msgs = append(msgs, patternClearMsg(pattern2))
	msgs = append(msgs, randomPatternNotes(pattern2, 4, 60)...)
	msgs = append(msgs, []*osc.Message{
		midiPatchMsg(1, 0, 0, 5),
		patternLoopMsg(1, 0, 0, pattern1),
		patternLoopMsg(1, 0, 250, pattern2),
		systemPlayMsg(),
	}...)

	return bundle(msgs...)
}

func loopPattern(pattern string) *osc.Bundle {
	return bundle(
		patternClearMsg(pattern),
		patternMidiNoteMsg(pattern, 0, 40, 400, 400, 127),
		patternMidiNoteMsg(pattern, 400, 41, 400, 400, 127),
		patternMidiNoteMsg(pattern, 800, 42, 400, 400, 127),
		patternMidiNoteMsg(pattern, 1200, 43, 400, 400, 127),
		midiPatchMsg(1, 0, 0, 10),
		patternLoopMsg(1, 0, 0, pattern),
		systemPlayMsg(),
	)
}

func tempoMsg(offset int, bpm float32) *osc.Message {
	return osc.NewMessage("/system/tempo", int32(offset), bpm)
}

// Set a bunch of arbitrary tempos. The point of this is that the tempo event
// should go into the sequence, and then we should be able to continue to
// schedule events expressed in terms of milliseconds and the player should
// convert them into ticks correctly according to whatever the tempo is at that
// point in time.
//
// Tempo isn't really important for playback, but it's important that we be able
// to set it so that we can export MIDI files that work with other software.
func variousTempos() *osc.Bundle {
	return bundle(
		tempoMsg(0, 130),
		tempoMsg(500, 62),
		tempoMsg(5000, 200),
		tempoMsg(10000, 400),
	)
}

func printUsage() {
	fmt.Printf("Usage: %s PORT EXAMPLE\n", os.Args[0])
}

func main() {
	rand.Seed(time.Now().Unix())

	numArgs := len(os.Args[1:])

	if numArgs < 1 || numArgs > 2 {
		printUsage()
		os.Exit(1)
	}

	port, err := strconv.ParseInt(os.Args[1], 10, 32)
	if err != nil {
		fmt.Println(err)
		printUsage()
		os.Exit(1)
	}

	var example string
	if numArgs < 2 {
		example = "1"
	} else {
		example = os.Args[2]
	}

	client := osc.NewClient("127.0.0.1", int(port), osc.ClientProtocol(osc.TCP))

	switch example {
	case "shutdown":
		client.Send(systemShutdownMsg(0))
	case "play":
		client.Send(systemPlayMsg())
	case "stop":
		client.Send(systemStopMsg())
	case "clear":
		client.Send(systemClearMsg())
	case "clear1":
		client.Send(clearMsg(1))
	case "export":
		client.Send(exportMsg("/tmp/alda-test.mid"))
	case "drums":
		client.Send(drumsDemo())
	case "1":
		client.Send(oneNote())
	case "16fast":
		client.Send(sixteenFastNotes())
	case "pat1":
		// NOTE: A quirk of the way that pattern scheduling works is that the MIDI
		// engine needs to be playing, first, so that the "schedule pattern"
		// meta-message will be encountered and processed.
		//
		// If the engine is stopped, then `awaitActiveTasks()` will block until the
		// events are scheduled, so the engine never starts playing. It's sort of a
		// chicken-and-egg problem. To work around that, we'll just send a "play"
		// message here to ensure that the engine is playing.
		client.Send(bundle(systemPlayMsg()))
		client.Send(playPatternOnce("simple"))
	case "pat2":
		client.Send(bundle(systemPlayMsg()))
		client.Send(playPatternTwice("simple"))
	case "pat3":
		client.Send(bundle(systemPlayMsg()))
		client.Send(playPatternThrice("simple"))
	case "patchange":
		client.Send(changePattern("simple"))
	case "patx":
		client.Send(patternClearMsg("simple"))
	case "patloop":
		client.Send(bundle(systemPlayMsg()))
		client.Send(loopPattern("simple"))
	case "patfin":
		client.Send(finishLoopMsg(1))
	case "2loops":
		client.Send(bundle(systemPlayMsg()))
		client.Send(twoFiniteLoops())
	case "2infinity":
		client.Send(bundle(systemPlayMsg()))
		client.Send(twoInfiniteLoops())
	case "tempos":
		client.Send(variousTempos())
	default:
		fmt.Printf("No such example: %s\n", example)
		os.Exit(1)
	}
}

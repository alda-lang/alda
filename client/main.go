package main

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var port int

func appendAll(bundle *osc.Bundle, msgs []*osc.Message) {
	for _, msg := range msgs {
		bundle.Append(msg)
	}
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

func muteMsg(track int) *osc.Message {
	return osc.NewMessage(fmt.Sprintf("/track/%d/mute", track))
}

func unmuteMsg(track int) *osc.Message {
	return osc.NewMessage(fmt.Sprintf("/track/%d/unmute", track))
}

func exportMsg(filepath string) *osc.Message {
	msg := osc.NewMessage("/system/midi/export")
	msg.Append(filepath)
	return msg
}

func midiPatchMsg(track int, offset int, patch int) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/patch", track))
	msg.Append(int32(offset))
	msg.Append(int32(patch))
	return msg
}

func midiPercussionMsg(track int) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/percussion", track))
	msg.Append(int32(0))
	return msg
}

func finishLoopMsg(track int) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/finish-loop", track))
	msg.Append(int32(0))
	return msg
}

func midiNoteMsg(
	track int, offset int, note int, duration int, audibleDuration int,
	velocity int) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/midi/note", track))
	msg.Append(int32(offset))
	msg.Append(int32(note))
	msg.Append(int32(duration))
	msg.Append(int32(audibleDuration))
	msg.Append(int32(velocity))
	return msg
}

func patternMsg(track int, offset int, pattern string, times int) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/pattern", track))
	msg.Append(int32(offset))
	msg.Append(pattern)
	msg.Append(int32(times))
	return msg
}

func patternMidiNoteMsg(
	pattern string, offset int, note int, duration int, audibleDuration int,
	velocity int) *osc.Message {
	msg := osc.NewMessage(fmt.Sprintf("/pattern/%s/midi/note", pattern))
	msg.Append(int32(offset))
	msg.Append(int32(note))
	msg.Append(int32(duration))
	msg.Append(int32(audibleDuration))
	msg.Append(int32(velocity))
	return msg
}

func patternClearMsg(pattern string) *osc.Message {
	return osc.NewMessage(fmt.Sprintf("/pattern/%s/clear", pattern))
}

func patternLoopMsg(pattern string, offset int) *osc.Message {
	track := 1
	msg := osc.NewMessage(fmt.Sprintf("/track/%d/pattern-loop", track))
	msg.Append(int32(offset))
	msg.Append(pattern)
	return msg
}

func oneNote() *osc.Bundle {
	bundle := osc.NewBundle(time.Now())
	bundle.Append(midiPatchMsg(1, 0, 30))
	bundle.Append(midiNoteMsg(1, 0, 45, 1000, 1000, 127))
	bundle.Append(systemPlayMsg())
	return bundle
}

func sixteenFastNotes() *osc.Bundle {
	bundle := osc.NewBundle(time.Now())

	bundle.Append(midiPatchMsg(1, 0, 70))

	interval := 100
	audibleDuration := 80

	noteNumber := 30 + rand.Intn(60)

	for offset := 0; offset <= interval*16; offset += interval {
		bundle.Append(
			midiNoteMsg(1, offset, noteNumber, interval, audibleDuration, 127))
	}

	bundle.Append(systemPlayMsg())

	return bundle
}

func playPattern(pattern string, times int) *osc.Bundle {
	bundle := osc.NewBundle(time.Now())
	bundle.Append(patternClearMsg(pattern))
	bundle.Append(patternMidiNoteMsg(pattern, 0, 57, 500, 500, 127))
	bundle.Append(patternMidiNoteMsg(pattern, 500, 60, 500, 500, 127))
	bundle.Append(patternMidiNoteMsg(pattern, 1000, 62, 500, 500, 127))
	bundle.Append(patternMidiNoteMsg(pattern, 1500, 64, 500, 500, 127))
	bundle.Append(patternMidiNoteMsg(pattern, 2000, 67, 500, 500, 127))
	bundle.Append(midiPatchMsg(1, 0, 60))
	bundle.Append(patternMsg(1, 0, pattern, times))
	bundle.Append(systemPlayMsg())
	return bundle
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

func randomPatternNotes(pattern string, quantity int) []*osc.Message {
	msgs := []*osc.Message{}

	interval := 500
	audibleDuration := 250

	for offset := 0; offset < interval*quantity; offset += interval {
		noteNumber := 30 + rand.Intn(60)

		msgs = append(msgs, patternMidiNoteMsg(
			pattern, offset, noteNumber, interval, audibleDuration, 127,
		))
	}

	return msgs
}

func changePattern(pattern string) *osc.Bundle {
	bundle := osc.NewBundle(time.Now())
	bundle.Append(patternClearMsg(pattern))
	appendAll(bundle, randomPatternNotes(pattern, 4))
	return bundle
}

// Plays two loops concurrently, four times each, on the same track.
func twoFiniteLoops() *osc.Bundle {
	bundle := osc.NewBundle(time.Now())

	pattern1 := "pattern1"
	bundle.Append(patternClearMsg(pattern1))
	appendAll(bundle, randomPatternNotes(pattern1, 4))

	pattern2 := "pattern2"
	bundle.Append(patternClearMsg(pattern2))
	appendAll(bundle, randomPatternNotes(pattern2, 4))

	bundle.Append(midiPatchMsg(1, 0, 0))
	bundle.Append(patternMsg(1, 0, pattern1, 4))
	bundle.Append(patternMsg(1, 100, pattern2, 4))

	bundle.Append(systemPlayMsg())
	return bundle
}

// Plays two indefinite loops concurrently on the same track.
func twoInfiniteLoops() *osc.Bundle {
	bundle := osc.NewBundle(time.Now())

	pattern1 := "pattern1"
	bundle.Append(patternClearMsg(pattern1))
	appendAll(bundle, randomPatternNotes(pattern1, 4))

	pattern2 := "pattern2"
	bundle.Append(patternClearMsg(pattern2))
	appendAll(bundle, randomPatternNotes(pattern2, 4))

	bundle.Append(midiPatchMsg(1, 0, 5))
	bundle.Append(patternLoopMsg(pattern1, 0))
	bundle.Append(patternLoopMsg(pattern2, 100))

	bundle.Append(systemPlayMsg())
	return bundle
}

func loopPattern(pattern string) *osc.Bundle {
	bundle := osc.NewBundle(time.Now())
	bundle.Append(patternClearMsg(pattern))
	bundle.Append(patternMidiNoteMsg(pattern, 0, 40, 400, 400, 127))
	bundle.Append(patternMidiNoteMsg(pattern, 400, 41, 400, 400, 127))
	bundle.Append(patternMidiNoteMsg(pattern, 800, 42, 400, 400, 127))
	bundle.Append(patternMidiNoteMsg(pattern, 1200, 43, 400, 400, 127))
	bundle.Append(midiPatchMsg(1, 0, 10))
	bundle.Append(patternLoopMsg(pattern, 0))
	bundle.Append(systemPlayMsg())
	return bundle
}

func tempoMsg(offset int, bpm float32) *osc.Message {
	msg := osc.NewMessage("/system/tempo")
	msg.Append(int32(offset))
	msg.Append(float32(bpm))
	return msg
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
	bundle := osc.NewBundle(time.Now())
	bundle.Append(tempoMsg(0, 130))
	bundle.Append(tempoMsg(500, 62))
	bundle.Append(tempoMsg(5000, 200))
	bundle.Append(tempoMsg(10000, 400))
	return bundle
}

func printUsage() {
	fmt.Printf("Usage: %s PORT EXAMPLE\n", os.Args[0])
}

// TODO:
// * Command/arg parsing via cobra or similar
// * Top-level -v / --verbose flag that sets the log level via
//   log.SetGlobalLevel

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

	client := osc.NewClient("localhost", int(port))

	switch example {
	case "play":
		client.Send(systemPlayMsg())
	case "stop":
		client.Send(systemStopMsg())
	case "clear":
		client.Send(systemClearMsg())
	case "clear1":
		client.Send(clearMsg(1))
	case "mute":
		client.Send(muteMsg(1))
	case "unmute":
		client.Send(unmuteMsg(1))
	case "export":
		client.Send(exportMsg("/tmp/alda-test.mid"))
	case "perc":
		client.Send(midiPercussionMsg(1))
	case "1":
		client.Send(oneNote())
	case "16fast":
		client.Send(sixteenFastNotes())
	case "pat1":
		client.Send(playPatternOnce("simple"))
	case "pat2":
		client.Send(playPatternTwice("simple"))
	case "pat3":
		client.Send(playPatternThrice("simple"))
	case "patchange":
		client.Send(changePattern("simple"))
	case "patx":
		client.Send(patternClearMsg("simple"))
	case "patloop":
		client.Send(loopPattern("simple"))
	case "patfin":
		client.Send(finishLoopMsg(1))
	case "2loops":
		client.Send(twoFiniteLoops())
	case "2infinity":
		client.Send(twoInfiniteLoops())
	case "tempos":
		client.Send(variousTempos())
	default:
		fmt.Printf("No such example: %s\n", example)
		os.Exit(1)
	}
}

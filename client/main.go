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

func systemPlayMsg() *osc.Message {
	return osc.NewMessage("/system/play")
}

func systemStopMsg() *osc.Message {
	return osc.NewMessage("/system/stop")
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

func playPattern(times int) *osc.Bundle {
	pattern := "simple"
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

func playPatternOnce() *osc.Bundle {
	return playPattern(1)
}

func playPatternTwice() *osc.Bundle {
	return playPattern(2)
}

func changePattern() *osc.Bundle {
	pattern := "simple"
	bundle := osc.NewBundle(time.Now())
	bundle.Append(patternClearMsg(pattern))

	interval := 500
	audibleDuration := 250

	for offset := 0; offset <= interval*3; offset += interval {
		noteNumber := 30 + rand.Intn(60)

		bundle.Append(
			patternMidiNoteMsg(
				pattern, offset, noteNumber, interval, audibleDuration, 127))
	}

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
	case "perc":
		client.Send(midiPercussionMsg(1))
	case "1":
		client.Send(oneNote())
	case "16fast":
		client.Send(sixteenFastNotes())
	case "pat1":
		client.Send(playPatternOnce())
	case "pat2":
		client.Send(playPatternTwice())
	case "patchange":
		client.Send(changePattern())
	case "patx":
		client.Send(patternClearMsg("simple"))
	default:
		fmt.Printf("No such example: %s\n", example)
		os.Exit(1)
	}
}

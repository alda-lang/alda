package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"alda.io/client/emitter"
	"alda.io/client/model"
	"alda.io/client/parser"
	"github.com/daveyarwood/go-osc/osc"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/midireader"
)

var _action string

func step(action string) {
	_action = action
}

func success() {
	fmt.Printf("%s %s\n", aurora.Green("OK "), _action)
}

func failure(err error) {
	fmt.Printf("%s %s\n", aurora.Red("ERR"), _action)
	fmt.Println(err)
	os.Exit(1)
}

// OSCPacketForwarder is a simple Dispatcher that puts each OSC packet that it
// receives onto a channel.
type OSCPacketForwarder struct {
	channel chan osc.Packet
}

// Dispatch implements osc.Dispatcher.Dispatch by putting the packet onto a
// channel.
func (oscpf OSCPacketForwarder) Dispatch(packet osc.Packet) {
	oscpf.channel <- packet
}

var noAudio bool

func init() {
	doctorCmd.Flags().BoolVarP(
		&noAudio,
		"no-audio",
		"",
		false,
		"disable checks that require an audio device",
	)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run health checks to determine if Alda can run correctly",
	Run: func(_ *cobra.Command, args []string) {
		testInput := "glockenspiel: o6 {c e g}8 > c4."
		expectedNotes := []uint8{84, 88, 91, 96}

		step("Parse source code")

		scoreUpdates, err := parser.ParseString(testInput)
		if err != nil {
			failure(err)
		}

		success()

		//////////////////////////////////////////////////

		step("Generate score model")

		score := model.NewScore()
		if err := score.Update(scoreUpdates...); err != nil {
			failure(err)
		}

		success()

		//////////////////////////////////////////////////

		step("Find an open port")

		port, err := findOpenPort()
		if err != nil {
			failure(err)
		}

		success()

		//////////////////////////////////////////////////

		// NB: Receiving messages isn't strictly necessary for the client to
		// function, but if it fails, it indicates that there might be networking
		// issues, which could mean that the player isn't able to listen for
		// messages.
		step("Send and receive OSC messages")

		packetsReceived := make(chan osc.Packet, 1000)
		errors := make(chan error)
		timeout := time.After(5 * time.Second)

		server := osc.NewServer(
			fmt.Sprintf("127.0.0.1:%d", port),
			OSCPacketForwarder{channel: packetsReceived},
			0,
		)

		server.SetNetworkProtocol(osc.TCP)

		go func() {
			if err := server.ListenAndServe(); err != nil {
				errors <- err
			}
		}()

		for {
			err := emitter.OSCEmitter{Port: port}.EmitScore(score)

			if err == nil {
				break
			}

			if !strings.Contains(err.Error(), "connection refused") {
				errors <- err
				break
			}

			select {
			case <-timeout:
				errors <- err
				break
			default:
				time.Sleep(500 * time.Millisecond)
			}
		}

		select {
		case <-packetsReceived:
			// success!
		case err := <-errors:
			failure(err)
		case <-timeout:
			failure(fmt.Errorf("timed out waiting for packet"))
		}

		server.CloseConnection()

		success()

		//////////////////////////////////////////////////

		step("Locate alda-player executable on PATH")

		aldaPlayer, err := exec.LookPath("alda-player")
		if err != nil {
			failure(err)
		}

		success()

		//////////////////////////////////////////////////

		step("Spawn a player process")

		port, err = findOpenPort()
		if err != nil {
			failure(err)
		}

		playerArgs := []string{"-v", "run", "-p", strconv.Itoa(port)}
		if noAudio {
			playerArgs = append(playerArgs, "--lazy-audio")
		}

		cmd := exec.Command(aldaPlayer, playerArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			failure(err)
		}

		success()

		//////////////////////////////////////////////////

		// This step ensures that the player is listening so that the following
		// steps can pass.
		step("Ping player process")

		client := osc.NewClient("localhost", port)
		client.SetNetworkProtocol(osc.TCP)

		deadline := time.Now().Add(5 * time.Second)

		for {
			err := client.Send(osc.NewMessage("/ping"))

			if err == nil {
				break
			}

			if !strings.Contains(err.Error(), "connection refused") ||
				time.Now().After(deadline) {
				failure(err)
			}

			time.Sleep(500 * time.Millisecond)
		}

		success()

		//////////////////////////////////////////////////

		if !noAudio {
			step("Play score")

			if err := (emitter.OSCEmitter{Port: port}).EmitScore(score); err != nil {
				failure(err)
			}

			success()
		}

		//////////////////////////////////////////////////

		if !noAudio {
			step("Export score as MIDI")

			tmpdir, err := ioutil.TempDir("", "alda-doctor")
			if err != nil {
				failure(err)
			}

			midiFilename := filepath.Join(
				tmpdir, fmt.Sprintf("%d.mid", time.Now().Unix()),
			)

			msg := osc.NewMessage("/system/midi/export")
			msg.Append(midiFilename)

			if err := client.Send(msg); err != nil {
				failure(err)
			}

			var midiFile *os.File
			deadline = time.Now().Add(5 * time.Second)

			for {
				midiFile, err = os.Open(midiFilename)

				if err == nil {
					break
				}

				if time.Now().After(deadline) {
					failure(err)
				}

				time.Sleep(500 * time.Millisecond)
			}

			rdr := midireader.New(bufio.NewReader(midiFile), nil)

			// NB: For some reason, the contents of the MIDI sequence (at least,
			// according to the library we're using to parse the MIDI file) appear to
			// contain a bunch of other messages that don't make sense in the context
			// of the score, in addition to containing all the notes that we expect.
			//
			// I suspect that there are some edge cases that the MIDI parsing library
			// isn't handling correctly. Perhaps there is some header information that
			// the MIDI parsing library is interpreting as NoteOn/NoteOff messages?
			//
			// When I play the MIDI file, it sounds correct, and when I load it into
			// MuseScore, the sheet music looks correct, so I'm satisfied that the
			// player is exporting usable MIDI files.
			//
			// Since this is just a smoke test, we only really need to test that the
			// MIDI sequence contains the notes we expect, and we can ignore all of
			// the other messages. If this test passes, then we can be confident that
			// the client can talk to the player, the player is up and running, and
			// the player successfully handled the "play" and "export" instructions.
		ExpectedNotesLoop:
			for _, expectedNote := range expectedNotes {
				for {
					msg, err := rdr.Read()

					// Error scenarios include reaching EOF (err == io.EOF), which we
					// consider a failure case because we reached the end before we saw
					// that all of the expected notes were included in the sequence.
					if err != nil {
						failure(err)
					}

					switch msg.(type) {
					case channel.NoteOn:
						if msg.(channel.NoteOn).Key() == expectedNote {
							continue ExpectedNotesLoop
						}
					}
				}
			}

			success()
		}

		//////////////////////////////////////////////////

		step("Shut down player process")

		if err := client.Send(osc.NewMessage("/system/shutdown")); err != nil {
			failure(err)
		}

		success()
	},
}

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
	"github.com/OpenPeeDeeP/xdg"
	"github.com/daveyarwood/go-osc/osc"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/midireader"
)

func step(action string, test func() error) {
	if err := test(); err != nil {
		fmt.Printf("%s %s\n", aurora.Red("ERR"), action)
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("%s %s\n", aurora.Green("OK "), action)
}

func await(test func() error, timeoutDuration time.Duration) error {
	timeout := time.After(timeoutDuration)

	for {
		err := test()

		if err == nil {
			return nil
		}

		select {
		case <-timeout:
			return err
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
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

		var scoreUpdates []model.ScoreUpdate

		step(
			"Parse source code",
			func() error {
				su, err := parser.ParseString(testInput)
				if err != nil {
					return err
				}

				scoreUpdates = su
				return nil
			},
		)

		//////////////////////////////////////////////////

		score := model.NewScore()

		step(
			"Generate score model",
			func() error {
				return score.Update(scoreUpdates...)
			},
		)

		//////////////////////////////////////////////////

		var port int

		step(
			"Find an open port",
			func() error {
				p, err := findOpenPort()
				if err != nil {
					return err
				}

				// Use this port in subsequent steps.
				port = p
				return nil
			},
		)

		//////////////////////////////////////////////////

		// NB: Receiving messages isn't strictly necessary for the client to
		// function, but if it fails, it indicates that there might be networking
		// issues, which could mean that the player isn't able to listen for
		// messages.
		step(
			"Send and receive OSC messages",
			func() error {
				packetsReceived := make(chan osc.Packet, 1000)
				errors := make(chan error)

				server := osc.NewServer(
					fmt.Sprintf("127.0.0.1:%d", port),
					OSCPacketForwarder{channel: packetsReceived},
					0,
				)

				defer server.CloseConnection()

				server.SetNetworkProtocol(osc.TCP)

				go func() {
					if err := server.ListenAndServe(); err != nil {
						errors <- err
					}
				}()

				if err := await(
					func() error {
						return emitter.OSCEmitter{Port: port}.EmitScore(score)
					},
					5*time.Second,
				); err != nil {
					errors <- err
				}

				select {
				case <-packetsReceived:
					return nil
				case err := <-errors:
					return err
				}
			},
		)

		//////////////////////////////////////////////////

		var aldaPlayer string

		step(
			"Locate alda-player executable on PATH",
			func() error {
				ap, err := exec.LookPath("alda-player")
				if err != nil {
					return err
				}

				aldaPlayer = ap
				return nil
			},
		)

		//////////////////////////////////////////////////

		step(
			"Spawn a player process",
			func() error {
				p, err := findOpenPort()
				if err != nil {
					return err
				}

				// Use this port is subsequent steps.
				port = p

				playerArgs := []string{"-v", "run", "-p", strconv.Itoa(port)}
				if noAudio {
					playerArgs = append(playerArgs, "--lazy-audio")
				}

				cmd := exec.Command(aldaPlayer, playerArgs...)

				if verbose {
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
				}

				return cmd.Start()
			},
		)

		//////////////////////////////////////////////////

		var client *osc.Client

		// This step ensures that the player is listening so that the following
		// steps can pass.
		step(
			"Ping player process",
			func() error {
				client = osc.NewClient("localhost", port)
				client.SetNetworkProtocol(osc.TCP)

				return await(
					func() error {
						return client.Send(osc.NewMessage("/ping"))
					},
					5*time.Second,
				)
			},
		)

		//////////////////////////////////////////////////

		if !noAudio {
			step(
				"Play score",
				func() error {
					return (emitter.OSCEmitter{Port: port}).EmitScore(score)
				},
			)
		}

		//////////////////////////////////////////////////

		if !noAudio {
			step(
				"Export score as MIDI",
				func() error {
					tmpdir, err := ioutil.TempDir("", "alda-doctor")
					if err != nil {
						return err
					}

					midiFilename := filepath.Join(
						tmpdir, fmt.Sprintf("%d.mid", time.Now().Unix()),
					)

					msg := osc.NewMessage("/system/midi/export")
					msg.Append(midiFilename)

					if err := client.Send(msg); err != nil {
						return err
					}

					var midiFile *os.File

					if err := await(
						func() error {
							mf, err := os.Open(midiFilename)
							if err != nil {
								return err
							}

							midiFile = mf
							return nil
						},
						5*time.Second,
					); err != nil {
						return err
					}

					rdr := midireader.New(bufio.NewReader(midiFile), nil)

					// NB: For some reason, the contents of the MIDI sequence (at least,
					// according to the library we're using to parse the MIDI file) appear
					// to contain a bunch of other messages that don't make sense in the
					// context of the score, in addition to containing all the notes that
					// we expect.
					//
					// I suspect that there are some edge cases that the MIDI parsing
					// library isn't handling correctly. Perhaps there is some header
					// information that the MIDI parsing library is interpreting as
					// NoteOn/NoteOff messages?
					//
					// When I play the MIDI file, it sounds correct, and when I load it
					// into MuseScore, the sheet music looks correct, so I'm satisfied
					// that the player is exporting usable MIDI files.
					//
					// Since this is just a smoke test, we only really need to test that
					// the MIDI sequence contains the notes we expect, and we can ignore
					// all of the other messages. If this test passes, then we can be
					// confident that the client can talk to the player, the player is up
					// and running, and the player successfully handled the "play" and
					// "export" instructions.
				ExpectedNotesLoop:
					for _, expectedNote := range expectedNotes {
						for {
							msg, err := rdr.Read()

							// Error scenarios include reaching EOF (err == io.EOF), which we
							// consider a failure case because we reached the end before we
							// saw that all of the expected notes were included in the
							// sequence.
							if err != nil {
								return err
							}

							switch msg.(type) {
							case channel.NoteOn:
								if msg.(channel.NoteOn).Key() == expectedNote {
									continue ExpectedNotesLoop
								}
							}
						}
					}

					return nil
				},
			)
		}

		//////////////////////////////////////////////////

		xdg := xdg.New("", "alda")
		var logFile string

		step(
			"Locate player logs",
			func() error {
				return await(
					func() error {
						logFilename := "alda-player.log"

						lf := xdg.QueryCache(logFilename)

						if lf == "" {
							return fmt.Errorf(
								"unable to locate %s in %s",
								logFilename,
								xdg.CacheHome(),
							)
						}

						logFile = lf
						return nil
					},
					5*time.Second,
				)
			},
		)

		//////////////////////////////////////////////////

		step(
			"Player logs show the ping was received",
			func() error {
				indication := "received ping"
				return await(
					func() error {
						contents, err := ioutil.ReadFile(logFile)
						if err != nil {
							return err
						}

						if !strings.Contains(string(contents), indication) {
							return fmt.Errorf("%s does not contain %s", logFile, indication)
						}

						return nil
					},
					5*time.Second,
				)
			},
		)

		//////////////////////////////////////////////////

		step(
			"Shut down player process",
			func() error {
				return client.Send(osc.NewMessage("/system/shutdown"))
			},
		)
	},
}

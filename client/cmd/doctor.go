package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"alda.io/client/color"
	"alda.io/client/help"
	"alda.io/client/model"
	"alda.io/client/parser"
	"alda.io/client/repl"
	"alda.io/client/system"
	"alda.io/client/text"
	"alda.io/client/transmitter"
	"alda.io/client/util"
	"github.com/daveyarwood/go-osc/osc"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/midireader"
)

const reasonableTimeout = 20 * time.Second

func step(action string, test func() error) error {
	if err := test(); err != nil {
		fmt.Printf("%s %s\n\n---\n\n", color.Aurora.Red("ERR"), action)
		return err
	}

	fmt.Printf("%s %s\n", color.Aurora.Green("OK "), action)
	return nil
}

func sendShutdownMessage(client *osc.Client) error {
	msg := osc.NewMessage("/system/shutdown")
	msg.Append(int32(0))
	return client.Send(msg)
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
		"Disable checks that require an audio device",
	)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run health checks to determine if Alda can run correctly",
	RunE: func(_ *cobra.Command, args []string) error {
		testInput := "glockenspiel: o6 {c e g}8 > c4."
		expectedNotes := []uint8{84, 88, 91, 96}

		var scoreUpdates []model.ScoreUpdate

		if err := step(
			"Parse source code",
			func() error {
				su, err := parser.ParseString(testInput)
				if err != nil {
					return err
				}

				scoreUpdates = su
				return nil
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		score := model.NewScore()

		if err := step(
			"Generate score model",
			func() error {
				return score.Update(scoreUpdates...)
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		var playerPort int

		if err := step(
			"Find an open port",
			func() error {
				p, err := system.FindOpenPort()
				if err != nil {
					return err
				}

				// Use this port in subsequent steps.
				playerPort = p
				return nil
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		// NB: Receiving messages isn't strictly necessary for the client to
		// function, but if it fails, it indicates that there might be networking
		// issues, which could mean that the player isn't able to listen for
		// messages.
		if err := step(
			"Send and receive OSC messages",
			func() error {
				packetsReceived := make(chan osc.Packet, 1000)
				errors := make(chan error)

				server := osc.NewServer(
					fmt.Sprintf("localhost:%d", playerPort),
					OSCPacketForwarder{channel: packetsReceived},
					0,
					osc.ServerProtocol(osc.TCP),
				)

				defer server.CloseConnection()

				go func() {
					if err := server.ListenAndServe(); err != nil {
						errors <- err
					}
				}()

				if err := util.Await(
					func() error {
						transmitter := transmitter.OSCTransmitter{Port: playerPort}
						return transmitter.TransmitScore(score)
					},
					reasonableTimeout,
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
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		var aldaPlayer string
		var sameVersion bool

		if err := step(
			"Locate alda-player executable on PATH",
			func() error {
				ap, sv, err := system.AldaPlayerPath()
				if err == system.ErrAldaPlayerNotFoundOnPath {
					if !text.PromptForConfirmation(
						fmt.Sprintf(
							"\n%s does not appear to be installed.\nInstall %s now?",
							color.Aurora.Bold("alda-player"),
							color.Aurora.Bold("alda-player"),
						),
						true,
					) {
						return system.ErrAldaPlayerNotFoundOnPath
					}

					if err := installCorrectAldaPlayerVersion(); err != nil {
						return err
					}

					// Now that we've installed alda-player, let's make another attempt to
					// find `alda-player` on the PATH. It should succeed this time, but if
					// it doesn't, we'll return the error.
					ap, sv, err = system.AldaPlayerPath()
				}

				if err != nil {
					return err
				}

				aldaPlayer = ap
				sameVersion = sv
				return nil
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		if err := step(
			"Check alda-player version",
			func() error {
				if sameVersion {
					return nil
				}

				if !text.PromptForConfirmation(
					fmt.Sprintf(
						"\nThe versions of %s and %s that you have installed are "+
							"different.\nInstall the correct version of %s?",
						color.Aurora.Bold("alda"),
						color.Aurora.Bold("alda-player"),
						color.Aurora.Bold("alda-player"),
					),
					true,
				) {
					return help.UserFacingErrorf(
						`The versions of %s and %s that you have installed are different.
This might cause unexpected problems.

For best results, run %s and follow the prompt to install the correct
version of %s.`,
						color.Aurora.Bold("alda"),
						color.Aurora.Bold("alda-player"),
						color.Aurora.BrightYellow("alda doctor"),
						color.Aurora.Bold("alda-player"),
					)
				}

				fmt.Println()
				return installCorrectAldaPlayerVersion()
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		if err := step(
			"Spawn a player process",
			func() error {
				p, err := system.FindOpenPort()
				if err != nil {
					return err
				}

				// Use this port in subsequent steps.
				playerPort = p

				playerArgs := []string{"-v", "run", "-p", strconv.Itoa(playerPort)}
				if noAudio {
					playerArgs = append(playerArgs, "--lazy-audio")
				}

				cmd := exec.Command(aldaPlayer, playerArgs...)

				if verbosity > 1 {
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
				}

				return cmd.Start()
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		var client *osc.Client

		// This step ensures that the player is listening so that the following
		// steps can pass.
		if err := step(
			"Ping player process",
			func() error {
				pingClient, err := system.PingPlayer(playerPort)
				if err != nil {
					return err
				}

				// Use this client in subsequent steps.
				client = pingClient
				return nil
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		if !noAudio {
			if err := step(
				"Play score",
				func() error {
					transmitter := transmitter.OSCTransmitter{Port: playerPort}
					return transmitter.TransmitScore(score)
				},
			); err != nil {
				return err
			}
		}

		//////////////////////////////////////////////////

		if !noAudio {
			if err := step(
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

					if err := util.Await(
						func() error {
							mf, err := os.Open(midiFilename)
							if err != nil {
								return err
							}

							midiFile = mf
							return nil
						},
						reasonableTimeout,
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
			); err != nil {
				return err
			}
		}

		//////////////////////////////////////////////////

		var logFile string

		if err := step(
			"Locate player logs",
			func() error {
				return util.Await(
					func() error {
						logFilename := filepath.Join("logs", "alda-player.log")

						lf := system.QueryCache(logFilename)

						if lf == "" {
							return fmt.Errorf(
								"unable to locate %s in %s",
								logFilename,
								system.CacheDir,
							)
						}

						logFile = lf
						return nil
					},
					reasonableTimeout,
				)
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		if err := step(
			"Player logs show the ping was received",
			func() error {
				indication := "received ping"
				return util.Await(
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
					reasonableTimeout,
				)
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		if err := step(
			"Shut down player process",
			func() error {
				return sendShutdownMessage(client)
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		if err := step(
			"Spawn a player on an unknown port",
			func() error {
				playerArgs := []string{"-v", "run"}
				if noAudio {
					playerArgs = append(playerArgs, "--lazy-audio")
				}

				return exec.Command(aldaPlayer, playerArgs...).Start()
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		var player system.PlayerState

		if err := step(
			"Discover the player",
			func() error {
				return util.Await(
					func() error {
						players, err := system.ReadPlayerStates()
						if err != nil {
							return err
						}

						foundPlayer := false
						for _, p := range players {
							// A player process only enters the "ready" state after the MIDI
							// system is initialized, and that never happens if the --no-audio
							// flag is provided.
							if p.State == "ready" || (noAudio && p.State == "starting") {
								foundPlayer = true
								player = p
								break
							}
						}

						if !foundPlayer {
							return system.ErrNoPlayersAvailable
						}

						// We're doing all of this so fast that the player we find might
						// inadvertently be the one that we spawned and shut down earlier.
						// (This can only happen if the --no-audio flag is provided, because
						// otherwise, the player we just used will have been used to play a
						// score already and thus will not be considered "available.")
						//
						// To avoid this, we check the port of the player we just found
						// and consider it a failure condition. This will cue `util.Await`
						// to keep checking until it finds a player that isn't that one.
						if player.Port == playerPort {
							return fmt.Errorf(
								"only found the player from before (%s/%d)",
								player.ID, player.Port,
							)
						}

						return nil
					},
					reasonableTimeout,
				)
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		if err := step(
			"Ping the player",
			func() error {
				pingClient, err := system.PingPlayer(player.Port)
				if err != nil {
					return err
				}

				// Use this client in subsequent steps.
				client = pingClient
				return nil
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		if err := step(
			"Shut the player down",
			func() error {
				if err := sendShutdownMessage(client); err != nil {
					return err
				}

				return util.Await(
					func() error {
						players, err := system.ReadPlayerStates()
						if err != nil {
							return err
						}

						for _, p := range players {
							if p.Port == player.Port {
								return fmt.Errorf("player state file still exists")
							}
						}

						return nil
					},
					reasonableTimeout,
				)
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		var replServer *repl.Server

		if err := step(
			"Start a REPL server",
			func() error {
				port, err := system.FindOpenPort()
				if err != nil {
					return err
				}

				server, err := repl.RunServer(port)
				if err != nil {
					return err
				}

				replServer = server
				return nil
			},
		); err != nil {
			return err
		}

		// Ensure that the server is closed on normal exit.
		defer replServer.Close()

		// Ensure that the server is closed if the process is interrupted or
		// terminated.
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-signals
			replServer.Close()
		}()

		//////////////////////////////////////////////////

		if err := step(
			"Find the REPL server",
			func() error {
				return util.Await(
					func() error {
						servers, err := system.ReadREPLServerStates()
						if err != nil {
							return err
						}

						for _, s := range servers {
							if s.Port == replServer.Port {
								return nil
							}
						}

						return fmt.Errorf("REPL server state file not found")
					},
					reasonableTimeout,
				)
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		if err := step(
			"Interact with the REPL server",
			func() error {
				client, err := repl.NewClient("localhost", replServer.Port)
				if err != nil {
					return err
				}
				defer client.Disconnect()

				// This sends "clone" and "describe" messages and relies on specific
				// values in the responses. This in itself is a good test of the
				// communication between an Alda REPL server and client.
				_, err = client.StartSession()
				return err
			},
		); err != nil {
			return err
		}

		//////////////////////////////////////////////////

		if err := step(
			"Shut down the REPL server",
			func() error {
				replServer.Close()

				return util.Await(
					func() error {
						servers, err := system.ReadREPLServerStates()
						if err != nil {
							return err
						}

						for _, s := range servers {
							if s.Port == replServer.Port {
								return fmt.Errorf("REPL server state file still exists")
							}
						}

						return nil
					},
					reasonableTimeout,
				)
			},
		); err != nil {
			return err
		}

		return nil
	},
}

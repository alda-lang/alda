package repl

import (
	encjson "encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	bencode "github.com/jackpal/bencode-go"

	"alda.io/client/generated"
	"alda.io/client/json"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"alda.io/client/system"
	"alda.io/client/transmitter"
	"alda.io/client/util"
)

const midiExportTimeout = 20 * time.Second

type nREPLRequest struct {
	conn net.Conn
	msg  map[string]interface{}
}

// Server is a stateful Alda REPL server object.
type Server struct {
	// A short, generated ID that appears in `alda ps` output.
	id string
	// The Port on which the server listens for nREPL messages from clients.
	Port int
	// The string of input that is built up over time as clients submit code, line
	// by line, to be evaluated and added to the score.
	input string
	// The stateful score object that should correspond to the input received so
	// far.
	score *model.Score
	// The current index into `score.Events`, representing where to start playing
	// any new events that are added to the score when input is added.
	eventIndex int
	// The server's most recent information about the player process it is using.
	player system.PlayerState
	// A queue onto which bdecoded messages from clients are placed in one
	// routine. In another routine, the messages are handled synchronously, one at
	// a time. Therefore, messages can be received asynchronously, but results are
	// processed synchronously to avoid concurrency issues due to global state.
	requestQueue chan nREPLRequest
}

func (server *Server) stateFile() string {
	return system.CachePath("state", "repl-servers", server.id+".json")
}

func (server *Server) respond(
	req nREPLRequest, status []string, data map[string]interface{},
) {
	if data == nil {
		data = make(map[string]interface{})
	}

	data["status"] = status

	if session, present := req.msg["session"]; present {
		data["session"] = session
	}

	if id, present := req.msg["id"]; present {
		data["id"] = id
	}

	log.Info().Interface("data", data).Msg("Sending response.")

	if err := bencode.Marshal(req.conn, data); err != nil {
		log.Warn().Interface("data", data).Msg("Failed to send response.")
	}
}

func (server *Server) respondDone(
	req nREPLRequest, data map[string]interface{},
) {
	server.respond(req, []string{"done"}, data)
}

func (server *Server) respondErrors(
	req nREPLRequest, problems []string, data map[string]interface{},
) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["problems"] = problems

	server.respond(req, []string{"done", "error"}, data)
}

func (server *Server) respondError(
	req nREPLRequest, problem string, data map[string]interface{},
) {
	server.respondErrors(req, []string{problem}, data)
}

func (server *Server) resetState() error {
	if server.hasPlayer() {
		if err := server.shutdownPlayer(); err != nil {
			return err
		}
	}

	server.input = ""
	server.score = model.NewScore()
	server.eventIndex = 0

	return nil
}

// Adapted from: https://www.calhoun.io/creating-random-strings-in-go/
func generateId() string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, 3)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}

// NewServer returns an initialized instance of an Alda REPL server.
func NewServer(port int) *Server {
	server := &Server{
		id:           generateId(),
		Port:         port,
		requestQueue: make(chan nREPLRequest),
	}
	server.resetState()
	return server
}

const nREPLPortFile = ".alda-nrepl-port"

// The nREPL server writes a file called ".alda-nrepl-port" into the current
// directory. This makes it easy for a client started in the same directory to
// discover what port the server is running on.
func (server *Server) writePortFile() {
	os.WriteFile(nREPLPortFile, []byte(strconv.Itoa(server.Port)), 0644)
}

func (server *Server) writeStateFile() {
	state := system.REPLServerState{ID: server.id, Port: server.Port}

	stateJSON, err := encjson.Marshal(state)
	if err != nil {
		log.Warn().
			Err(err).
			Interface("state", state).
			Msg("Failed to serialize REPL state JSON.")

		return
	}

	stateFile := server.stateFile()

	if err := os.MkdirAll(filepath.Dir(stateFile), os.ModePerm); err != nil {
		log.Warn().
			Err(err).
			Msg("Failed to create parent directories for REPL server state file.")
	}

	if err := os.WriteFile(stateFile, stateJSON, 0644); err != nil {
		log.Warn().
			Err(err).
			Msg("Failed to write REPL server state file.")
	}
}

func (server *Server) touchStateFile() {
	now := time.Now()

	if err := os.Chtimes(server.stateFile(), now, now); err != nil {
		log.Warn().
			Err(err).
			Msg("Failed to touch REPL server state file.")
	}
}

func (server *Server) manageStateFile() {
	// NOTE: We don't yet have a use case for exposing information about the
	// server that updates regularly. Therefore, to avoid doing unnecessary work,
	// we will just write the state file once and then we'll just continuously
	// update the last modified time without re-writing the file.
	server.writeStateFile()

	for {
		server.touchStateFile()
		time.Sleep(10 * time.Second)
	}
}

func (server *Server) removePortFile() {
	os.Remove(nREPLPortFile)
}

func (server *Server) removeStateFile() {
	os.Remove(server.stateFile())
}

// Close cleans up after a server is done serving.
//
// This includes actions like removing the nREPL port file.
func (server *Server) Close() {
	server.removePortFile()
	server.removeStateFile()
}

// RunServer creates a running Alda REPL server instance and returns it.
//
// Returns an error if something goes wrong.
//
// NOTE: The caller is responsible for calling `Close()` on the server instance
// when it is no longer needed. Otherwise, resources like the .alda-nrepl-port
// file will not be cleaned up.
func RunServer(port int) (*Server, error) {
	server := NewServer(port)

	l, err := net.Listen("tcp", "localhost:"+strconv.Itoa(server.Port))
	if err != nil {
		return nil, err
	}

	// This writes an .alda-nrepl-port file, which gets cleaned up when `Close()`
	// is invoked.
	server.writePortFile()

	// Continuously writes a state file so that this REPL server can be included
	// in the output of `alda ps`. This file also gets cleaned up by `Close()`.
	go server.manageStateFile()

	// See repl/player_management.go
	go server.managePlayers()

	go server.listen(l)
	go server.handleRequests()

	return server, nil
}

// Runs in a loop, listening for bencoded messages from clients, "bdecoding"
// them, and putting them on a channel to be handled by another routine.
//
// The processing of messages must be synchronous in order to avoid concurrency
// issues because all clients share the same (global) server state. The
// receiving of messages, however, is asynchronous, so that the transmission of
// the next message isn't blocked by the handling of the previous one.
func (server *Server) listen(l net.Listener) {
	defer l.Close()

	fmt.Printf(
		"nREPL server started on port %d on host %s - nrepl://%s:%d\n",
		server.Port,
		"localhost",
		"localhost",
		server.Port,
	)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Warn().Int("port", server.Port).Msg("Failed to accept connection.")
			continue
		}

		// We do this part (receiving and bdecoding bytes from the connection) in a
		// goroutine so that we can avoid blocking and immediately start waiting for
		// the next connection. That way, the message receiving part can be
		// asynchronous, even though we proceed to process the messages
		// synchronously by putting them onto a queue.
		go func() {
			defer conn.Close()

			for {
				decoded, err := bencode.Decode(conn)

				// I think this means the client disconnected? So assuming I'm right
				// about that, we should stop reading and close the connection.
				if err == io.EOF {
					break
				}

				if err != nil {
					log.Warn().
						Int("port", server.Port).
						Err(err).
						Msg("Failed to bdecode message from connection.")

					// If we fail to bdecode a message from the connection, then we bail
					// out and close the connection. I'm not 100% sure if this is the
					// right thing to do, but it seems reasonable, I guess.
					return
				}

				switch msg := decoded.(type) {
				default:
					log.Warn().
						Msg("Unable to process request; it isn't a map[string]interface{}")

				case map[string]interface{}:
					// Strings seem to become byte arrays somewhere in the process of
					// marshaling and unmarshaling to bencode. I don't have a use-case for
					// dealing with byte arrays, I only want to deal with strings, so
					// let's go ahead and do the conversion here.
					for k, v := range msg {
						switch vv := v.(type) {
						case []byte:
							msg[k] = string(vv)
						}
					}

					log.Info().
						Interface("decodedRequest", msg).
						Msg("Request received.")

					server.requestQueue <- nREPLRequest{conn: conn, msg: msg}
				}
			}
		}()
	}
}

var describeResponse = map[string]interface{}{
	"versions": map[string]interface{}{
		"alda": map[string]interface{}{
			"version-string": generated.ClientVersion,
		},
	},
}

func init() {
	describedOps := map[string]interface{}{}

	for op := range ops {
		// It isn't clear what information we should include in the value, so we're
		// just making it an empty map. I noticed that the Clojure nREPL server does
		// the same thing.
		describedOps[op] = map[string]interface{}{}
	}

	describeResponse["ops"] = describedOps
}

var ops = map[string]func(*Server, nREPLRequest){
	// NOTE: This is mostly for general nREPL protocol adherence. Sessions don't
	// have much meaning to an Alda REPL server. For now, we just fake it by
	// generating a session ID and giving it to the client.
	"clone": func(server *Server, req nREPLRequest) {
		server.respondDone(req, map[string]interface{}{
			"new-session": uuid.New().String(),
		})
	},

	// NOTE: This is for nREPL protocol adherence.
	"describe": func(server *Server, req nREPLRequest) {
		server.respondDone(req, describeResponse)
	},

	// NOTE: This is just for nREPL protocol adherence. It isn't clear to me yet
	// if there should be a distinct "eval" operation that does something
	// meaningful. So for now, we're just responding with a shrug.
	"eval": func(server *Server, req nREPLRequest) {
		server.respondDone(req, map[string]interface{}{"value": "¯\\_(ツ)_/¯"})
	},

	"eval-and-play": func(server *Server, req nREPLRequest) {
		errors := validateRequest(
			req.msg,
			requestFieldSpec{name: "code", valueType: typeString, required: true},
		)
		if len(errors) > 0 {
			server.respondErrors(req, errors, nil)
			return
		}

		input := req.msg["code"].(string)

		if err := server.evalAndPlay(input); err != nil {
			server.respondError(req, err.Error(), nil)
			return
		}

		server.respondDone(req, nil)
	},

	"export": func(server *Server, req nREPLRequest) {
		binaryData, err := server.export()
		if err != nil {
			server.respondError(req, err.Error(), nil)
			return
		}

		server.respondDone(req, map[string]interface{}{"binary-data": binaryData})
	},

	"instruments": func(server *Server, req nREPLRequest) {
		server.respondDone(req, map[string]interface{}{
			"instruments": model.InstrumentsList(),
		})
	},

	"load": func(server *Server, req nREPLRequest) {
		errors := validateRequest(
			req.msg,
			requestFieldSpec{name: "code", valueType: typeString, required: true},
		)
		if len(errors) > 0 {
			server.respondErrors(req, errors, nil)
			return
		}

		input := req.msg["code"].(string)

		if err := server.load(input); err != nil {
			server.respondError(req, err.Error(), nil)
			return
		}

		server.respondDone(req, nil)
	},

	"new-score": func(server *Server, req nREPLRequest) {
		if err := server.resetState(); err != nil {
			server.respondError(req, err.Error(), nil)
			return
		}

		server.respondDone(req, nil)
	},

	"replay": func(server *Server, req nREPLRequest) {
		transmitOpts := []transmitter.TransmissionOption{}

		from, hit := req.msg["from"]
		if hit {
			switch f := from.(type) {
			case string:
				transmitOpts = append(transmitOpts, transmitter.TransmitFrom(f))
			}
		}

		to, hit := req.msg["to"]
		if hit {
			switch t := to.(type) {
			case string:
				transmitOpts = append(transmitOpts, transmitter.TransmitTo(t))
			}
		}

		if err := server.replay(transmitOpts...); err != nil {
			server.respondError(req, err.Error(), nil)
			return
		}

		server.respondDone(req, nil)
	},

	"score-data": func(server *Server, req nREPLRequest) {
		server.respondDone(req, map[string]interface{}{
			"data": server.score.JSON().String(),
		})
	},

	"score-events": func(server *Server, req nREPLRequest) {
		scoreUpdates, err := parser.ParseString(server.input)
		if err != nil {
			server.respondError(req, err.Error(), nil)
			return
		}

		updates := json.Array()
		for _, update := range scoreUpdates {
			updates.ArrayAppend(update.JSON())
		}

		server.respondDone(req, map[string]interface{}{"events": updates.String()})
	},

	"score-text": func(server *Server, req nREPLRequest) {
		server.respondDone(req, map[string]interface{}{"text": server.input})
	},

	"stop": func(server *Server, req nREPLRequest) {
		if err := server.withTransmitter(
			func(transmitter transmitter.OSCTransmitter) error {
				log.Info().
					Interface("player", server.player).
					Msg("Sending \"stop\" message to player process.")
				return transmitter.TransmitStopMessage()
			},
		); err != nil {
			server.respondError(req, err.Error(), nil)
			return
		}

		server.respondDone(req, nil)
	},
}

// Runs in a loop, handling requests from the queue as they come in in a
// synchronous fashion, one at a time.
func (server *Server) handleRequests() {
	for req := range server.requestQueue {
		errors := validateRequest(
			req.msg,
			requestFieldSpec{name: "op", valueType: typeString, required: true},
		)
		if len(errors) > 0 {
			server.respondErrors(req, errors, nil)
			continue
		}

		op := req.msg["op"].(string)

		handler, supported := ops[op]
		if !supported {
			server.respond(req, []string{"done", "error", "unknown-op"}, nil)
			continue
		}

		handler(server, req)
	}
}

// Parses a string of `input`, updates the server's score and related state, and
// returns a list of transmission options that would make it so that we're
// transmitting only the new events that resulted from this string of input.
func (server *Server) updateScoreWithInput(
	input string,
) ([]transmitter.TransmissionOption, error) {
	// Take note of the current offsets of all parts in the score, for the purpose
	// of synchronization. (See below where we use the transmitter.SyncOffsets
	// option when transmitting the score.)
	partOffsets := server.score.PartOffsets()

	// Take note of the current `eventIndex` value, so that we know where to start
	// playing from when we want to play the new events.
	eventIndex := server.eventIndex

	scoreUpdates, err := parser.ParseString(input)
	if err != nil {
		return nil, err
	}

	if err := server.score.Update(scoreUpdates...); err != nil {
		return nil, err
	}

	// Add the provided `input` to our total string of input representing the
	// entire score.
	server.input += strings.TrimSpace(input) + "\n"

	// Update the starting index so that the next invocation of `evalAndPlay` for
	// this same score will result in only playing newly added events.
	server.eventIndex = len(server.score.Events)

	return []transmitter.TransmissionOption{
		// Transmit only the new events, i.e. events added as a result of parsing
		// the provided `input` and applying the resulting updates to the score.
		transmitter.TransmitFromIndex(eventIndex),
		// The previous offset of each part is subtracted from any new events for
		// that part. The effect is that we "synchronize" that part with the events
		// that we already sent to the player. For example, if a client submits the
		// following code:
		//
		//   piano: c d e f
		//
		// Followed by (sometime before the 4 notes above finish playing):
		//
		//   piano: g a b > c
		//
		// Then the notes `c d e f g a b > c` will be played in time.
		transmitter.SyncOffsets(partOffsets),
	}, nil
}

func (server *Server) evalAndPlay(
	input string, additionalTransmitOpts ...transmitter.TransmissionOption,
) error {
	return server.withTransmitter(
		func(transmitter transmitter.OSCTransmitter) error {
			transmitOpts, err := server.updateScoreWithInput(input)
			if err != nil {
				return err
			}

			log.Info().
				Interface("player", server.player).
				Msg("Sending OSC messages to player.")

			return transmitter.TransmitScore(
				server.score,
				(append(transmitOpts, additionalTransmitOpts...))...,
			)
		},
	)
}

func (server *Server) load(input string) error {
	if err := server.resetState(); err != nil {
		return err
	}

	return server.withTransmitter(
		func(t transmitter.OSCTransmitter) error {
			transmitOpts, err := server.updateScoreWithInput(input)
			if err != nil {
				return err
			}

			transmitOpts = append(transmitOpts, transmitter.LoadOnly())

			log.Info().
				Interface("player", server.player).
				Msg("Transmitting score to player.")

			err = t.TransmitScore(server.score, transmitOpts...)
			if err != nil {
				return err
			}

			newOffset := int32(0)
			for _, offset := range server.score.PartOffsets() {
				offsetRounded := int32(math.Round(offset))
				if offsetRounded > newOffset {
					newOffset = offsetRounded
				}
			}

			log.Info().
				Interface("player", server.player).
				Int32("newOffset", newOffset).
				Msg("Transmitting new offset to player.")

			return t.TransmitOffsetMessage(newOffset)
		},
	)
}

func (server *Server) reload() error {
	return server.load(server.input)
}

func (server *Server) replay(
	transmitOpts ...transmitter.TransmissionOption,
) error {
	// `input` is the one thing about the server state that we DON'T want to
	// reset, so we keep track of it here. After we reset the state, we invoke
	// `server.evalAndPlay` on this input, which has the effect of both playing it
	// and re-adding it to the state of the server.
	input := server.input

	// We reset the server state here so that we can re-transmit the score "from
	// scratch" (or just re-transmit the part that we want to hear, if `from`
	// and/or `to` parameters are provided). This makes it so that what we hear
	// corresponds more directly to the input entered so far.
	//
	// An alternative would be to tell the player to rewind to offset 0 and play
	// the sequence from the beginning, but that would preserve the pauses in
	// between the user entering each line of REPL input, which we are presuming
	// is not what the user wants. (This would also be a departure from the
	// behavior of `:play` in the Alda v1 REPL.)
	if err := server.resetState(); err != nil {
		return err
	}

	// At this point, the `managePlayers` loop should find a replacement for the
	// player, and this should generally happen quickly. `server.evalAndPlay` will
	// handle the case that a player process isn't immediately available, so it's
	// OK for us to call it immediately after resetting the state.
	return server.evalAndPlay(input, transmitOpts...)
}

// Reloads the score into a fresh player, sends a "MIDI export" message to the
// player, waits for the player to write the MIDI file, reads the file, and
// returns the bytes in the file.
//
// Returns an error if something goes wrong somewhere along the way.
func (server *Server) export() ([]byte, error) {
	// Reloading the score is important because of the subtleties of the tempo
	// messages in the MIDI sequence.
	//
	// When we're evaluating Alda code interactively at the REPL, we suppress
	// tempo messages because they serve no immediate purpose.
	//
	// When it comes time to export the score, we reload the input into the MIDI
	// sequencer, which does include sending tempo messages, so that the MIDI
	// sequence includes tempo changes in the places where we want them.
	if err := server.reload(); err != nil {
		return nil, err
	}

	tmpdir, err := ioutil.TempDir("", "alda-repl-server")
	if err != nil {
		return nil, err
	}

	midiFilename := filepath.Join(
		tmpdir, fmt.Sprintf(
			"export-%d-%d.mid",
			time.Now().Unix(),
			rand.Intn(10000),
		),
	)

	if err := server.withTransmitter(
		func(transmitter transmitter.OSCTransmitter) error {
			return transmitter.TransmitMidiExportMessage(midiFilename)
		},
	); err != nil {
		return nil, err
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
		midiExportTimeout,
	); err != nil {
		return nil, err
	}

	return ioutil.ReadAll(midiFile)
}

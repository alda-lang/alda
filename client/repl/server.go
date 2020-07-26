package repl

import (
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"alda.io/client/system"
	"alda.io/client/transmitter"
)

// Server is a stateful Alda REPL server object.
type Server struct {
	input      string
	score      *model.Score
	eventIndex int
	player     system.PlayerState
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

// RunServer creates a running Alda REPL server instance and returns it.
//
// Returns an error if something goes wrong.
func RunServer() (*Server, error) {
	server := &Server{}
	server.resetState()

	// See repl/player_management.go
	go server.managePlayers()

	return server, nil
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
	server.input += input + "\n"

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

func (server *Server) evalAndPlay(input string) error {
	return server.withTransmitter(
		func(transmitter transmitter.OSCTransmitter) error {
			transmitOpts, err := server.updateScoreWithInput(input)
			if err != nil {
				return err
			}

			log.Info().
				Interface("player", server.player).
				Msg("Sending OSC messages to player.")

			return transmitter.TransmitScore(server.score, transmitOpts...)
		},
	)
}

// TODO: support `from` and `to` parameters, when provided in the message
func (server *Server) replay() error {
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
	return server.evalAndPlay(input)
}

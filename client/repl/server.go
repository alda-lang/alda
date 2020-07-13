package repl

import (
	"time"

	"alda.io/client/emitter"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"alda.io/client/system"
	"alda.io/client/util"
)

const reasonableTimeout = 20 * time.Second

// Server is a stateful Alda REPL server object.
type Server struct {
	input      string
	score      *model.Score
	eventIndex int
	player     system.PlayerState
}

func findAvailablePlayer() (system.PlayerState, error) {
	var player system.PlayerState

	if err := util.Await(
		func() error {
			availablePlayer, err := system.FindAvailablePlayer()
			if err != nil {
				return err
			}

			player = availablePlayer
			return nil
		},
		reasonableTimeout,
	); err != nil {
		return system.PlayerState{}, err
	}

	return player, nil
}

// RunServer creates a running Alda REPL server instance and returns it.
//
// Returns an error if something goes wrong.
func RunServer() (*Server, error) {
	server := &Server{input: "", score: model.NewScore(), eventIndex: 0}

	// FIXME: Player management should be more robust.
	//
	// TODO:
	// * There should be a background routine that checks about once per minute
	// that there are enough player processes available and fill the pool of
	// available players as needed.
	//
	// * When the `:new` command is run, that should totally reset the state of
	//   the REPL server instance, and that should include sending the currently
	//   active player process a /shutdown message and selecting a new one.
	//
	// * If the player process we're using ends up becoming unavailable (e.g. the
	//   process falls over / becomes unresponsive), at a minimum we should print
	//   an error and start over with a new player process. Ideally, we can also
	//   silently recover by switching to a new player process and bootstrapping
	//   it to be in a correct-ish state. I think it might actually work out OK to
	//   just re-emit the entire score to the new player process. It won't always
	//   sound the same in light of the live-coding stuff we're going to add
	//   (where timing matters), but I think it might be good enough.
	//
	//   * I was thinking that some sort of heartbeat mechanism might make sense.
	//     Maybe those could just be ping messages, and we would send them
	//     regularly, thus preventing the player process we're using from expiring
	//     before we're done with it. (Currently the /ping message handler on the
	//     player side does nothing, but we could make it so that it prolongs the
	//     expiration)
	//
	// * We should somehow make it so that the REPL server can "claim" the player
	//   process it wants to use so that it doesn't end up also being used by e.g.
	//   another instance of a REPL server, or even just another Alda client
	//   running `alda play` on the same machine.
	//
	//   * This could simply be another side effect of the /ping message.
	player, err := findAvailablePlayer()
	if err != nil {
		return nil, err
	}
	server.player = player

	return server, nil
}

func (server *Server) play(input string) error {
	scoreUpdates, err := parser.ParseString(input)
	if err != nil {
		return err
	}

	if err := server.score.Update(scoreUpdates...); err != nil {
		return err
	}

	// Add the provided `input` to our total string of input representing the
	// entire score.
	server.input += input + "\n"

	emitOpts := []emitter.EmissionOption{
		// Emit only new events, i.e. events added as a result of parsing the
		// provided `input` and applying the resulting updates to the score.
		emitter.EmitFromIndex(server.eventIndex),
	}

	// Update the starting index so that the next invocation of `play` for this
	// same score will result in only playing newly added events.
	server.eventIndex = len(server.score.Events)

	emitter := emitter.OSCEmitter{Port: server.player.Port}

	if err := emitter.EmitScore(server.score, emitOpts...); err != nil {
		return err
	}

	log.Info().
		Interface("player", server.player).
		Msg("Sent OSC messages to player.")

	return nil
}

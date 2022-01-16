//go:build js || wasm

package main

// wasm/main.go is an entrypoint for compiling the main functionality of Alda
// into a WebAssembly module so that you can run Alda in the browser.
//
// The part that we compile to WebAssembly is the parser/compiler/OSC message
// generator part of the Alda client.
//
// The other piece of the puzzle that is necessary in order to play music in the
// browser is a JavaScript library that can receive the OSC messages and use
// them as instructions to play music via the Web Audio API / some JS MIDI
// library.

import (
	"alda.io/client/generated"
	log "alda.io/client/logging"
	"alda.io/client/model"
	"alda.io/client/parser"
	"alda.io/client/transmitter"

	"syscall/js"
)

// Boilerplate for converting a Go byte array into a JS Uint8Array that can be
// used in the browser.
func uint8Array(bytes []byte) js.Value {
	arr := js.Global().Get("Uint8Array").New(len(bytes))
	js.CopyBytesToJS(arr, bytes)
	return arr
}

// Given a function that takes an unspecified number of args from the JS caller
// and returns a single, JS-compatible value...
//
// Returns a JS function that invokes the provided function on the args provided
// by the JS caller.
//
// Recovers from panics so that the Go program doesn't exit and the WebAssembly
// module can continue to be available for use in the browser.
func safeFunction(f func([]js.Value) interface{}) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// If Alda panics, recover and log the error to the console.
		defer func() {
			if err := recover(); err != nil {
				js.Global().Get("console").Call("error", err)
			}
		}()

		return f(args)
	})
}

// Given a function that takes an unspecified number of args from the JS caller,
// does some work, and returns either a success value or an error...
//
// Returns a JS function that:
// * Invokes the provided function in a goroutine
// * Returns a promise that will either be resolved with the success value, or
//   rejected with the error
//
// Recovers from panics so that the Go program doesn't exit and the WebAssembly
// module can continue to be available for use in the browser.
func safePromiseFunction(f func([]js.Value) (interface{}, error)) js.Func {
	return safeFunction(func(args []js.Value) interface{} {
		return js.Global().Get("Promise").New(js.FuncOf(
			func(this js.Value, promiseArgs []js.Value) interface{} {
				resolve := promiseArgs[0]
				reject := promiseArgs[1]

				value, err := f(args)
				if err != nil {
					reject.Invoke(js.Global().Get("Error").New(err.Error()))
				} else {
					resolve.Invoke(js.ValueOf(value))
				}

				return nil
			},
		))
	})
}

func main() {
	log.SetGlobalLevel("warn")

	js.Global().Set("Alda", map[string]interface{}{
		"VERSION": generated.ClientVersion,

		"setLogLevel": safeFunction(func(args []js.Value) interface{} {
			level := args[0].String()
			log.SetGlobalLevel(level)
			return nil
		}),

		"toOSCBytes": safePromiseFunction(
			func(args []js.Value) (interface{}, error) {
				code := args[0].String()

				ast, err := parser.ParseString(code)
				if err != nil {
					return nil, err
				}

				updates, err := ast.Updates()
				if err != nil {
					return nil, err
				}

				score := model.NewScore()
				err = score.Update(updates...)
				if err != nil {
					return nil, err
				}

				bundle, err := transmitter.OSCTransmitter{}.ScoreToOSCBundle(
					score, transmitter.LoadOnly(),
				)
				if err != nil {
					return nil, err
				}

				bytes, err := bundle.MarshalBinary()
				if err != nil {
					return nil, err
				}

				return uint8Array(bytes), nil
			},
		),
	})

	// Run this WebAssembly module forever so that the library we've defined
	// remains available in the browser. Otherwise, the functions we've defined
	// throw an error about the Go program having already exited.
	select {}
}

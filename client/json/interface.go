package json

import "github.com/Jeffail/gabs/v2"

// A Container is a JSON array or object.
//
// This is a type alias for the Container type in github.com/Jeffail/gabs/v2.
type Container = gabs.Container

// ParseJSON takes a []byte and returns either a Container (JSON array or
// object) or an error if the parse was unsuccessful.
//
// This is an alias for the ParseJSON function in github.com/Jeffail/gabs/v2.
var ParseJSON = gabs.ParseJSON

// ParseJSONBuffer takes an io.Reader and returns either a Container (JSON array
// or object) or an error if the read or parse was unsuccessful.
//
// This is an alias for the ParseJSONBuffer function in
// github.com/Jeffail/gabs/v2.
var ParseJSONBuffer = gabs.ParseJSONBuffer

// RepresentableAsJSON is an interface implemented by types that can be
// represented as JSON data.
type RepresentableAsJSON interface {
	// JSON returns a JSON data representation of the object.
	JSON() *Container
}

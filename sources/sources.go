package sources

import (
	"errors"
	"reflect"
)

var (
	// ErrKeyExists should be returned when the key has already
	// been registered with the source and it can't be re-registered
	// e.g. when using CommandLine, which uses the flag package
	ErrKeyExists = errors.New("Key already exists")
)

// Source should be implemented by Gofigure sources, e.g.
// environment, command line, file, http etc
type Source interface {
	// Init is called at the start of a new struct
	Init(args map[string]interface{}) error
	// Cleanup is called at the end of parsing
	Cleanup()
	// Register is called to register each struct field
	Register(key, defaultValue string, t reflect.Type) error
	// Get is called to retrieve a key value
	// - FIXME could use interface{} and maintain types, e.g. json?
	Get(key string, overrideDefault *string) (string, error)
}

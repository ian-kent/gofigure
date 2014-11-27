package sources

import (
	"errors"
	"log"
	"reflect"
)

// Logger is called for each log message. If nil,
// log.Printf will be called instead.
var Logger func(message string, args ...interface{})

// Debug controls sources debug output
var Debug = false

func printf(message string, args ...interface{}) {
	if !Debug {
		return
	}
	if Logger != nil {
		Logger(message, args...)
	} else {
		log.Printf(message, args...)
	}
}

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
	Init(args map[string]string) error
	// Cleanup is called at the end of parsing
	Cleanup()
	// Register is called to register each struct field
	Register(key, defaultValue string, params map[string]string, t reflect.Type) error
	// Get is called to retrieve a key value
	// - FIXME could use interface{} and maintain types, e.g. json?
	Get(key string, overrideDefault *string) (string, error)
	// GetArray is called to retrieve an array value
	GetArray(key string, overrideDefault *[]string) ([]string, error)
}

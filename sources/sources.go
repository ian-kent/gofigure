package sources

import (
	"errors"
	"flag"
	"os"
	"reflect"

	"github.com/ian-kent/envconf"
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
	Init() error
	// Cleanup is called at the end of parsing
	Cleanup()
	// Register is called to register each struct field
	Register(key, defaultValue string, t reflect.Type) error
	// Get is called to retrieve a key value
	// - FIXME could use interface{} and maintain types, e.g. json?
	Get(key string, overrideDefault *string) (string, error)
}

// CommandLine implements command line configuration using the flag package
type CommandLine struct {
	flags map[string]*string
	oldCl *flag.FlagSet
}

// Init is called at the start of a new struct
func (cl *CommandLine) Init() error {
	cl.flags = make(map[string]*string)
	cl.oldCl = flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	return nil
}

// Cleanup is called at the end of parsing
func (cl *CommandLine) Cleanup() {
	flag.CommandLine = cl.oldCl
}

// Register is called to register each struct field
func (cl *CommandLine) Register(key, defaultValue string, t reflect.Type) error {
	if _, ok := cl.flags[key]; ok {
		return ErrKeyExists
	}
	// TODO validate key?
	// TODO use typed calls instead of StringVar
	val := defaultValue
	cl.flags[key] = &val
	flag.StringVar(&val, key, defaultValue, "TODO description")
	return nil
}

// Get is called to retrieve a key value
func (cl *CommandLine) Get(key string, overrideDefault *string) (string, error) {
	if !flag.CommandLine.Parsed() {
		flag.Parse()
	}
	// TODO check if flag exists/overrideDefault
	val := *cl.flags[key]
	if len(val) > 0 {
		return val, nil
	}
	return *overrideDefault, nil
}

// Environment implements environment variable configuration using envconf
type Environment struct {
	fields map[string]string
}

// Init is called at the start of a new struct
func (env *Environment) Init() error {
	env.fields = make(map[string]string)
	return nil
}

// Register is called to register each struct field
func (env *Environment) Register(key, defaultValue string, t reflect.Type) error {
	env.fields[key] = defaultValue
	return nil
}

// Get is called to retrieve a key value
func (env *Environment) Get(key string, overrideDefault *string) (string, error) {
	def := env.fields[key]
	if overrideDefault != nil {
		def = *overrideDefault
	}
	val, err := envconf.FromEnv(key, def)
	return val.(string), err
}

// Cleanup is called at the end of parsing
func (env *Environment) Cleanup() {

}

package sources

import (
	"flag"
	"os"
	"reflect"
)

// CommandLine implements command line configuration using the flag package
type CommandLine struct {
	flags map[string]*string
	oldCl *flag.FlagSet
}

// Init is called at the start of a new struct
func (cl *CommandLine) Init(args map[string]interface{}) error {
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
	if overrideDefault != nil {
		return *overrideDefault, nil
	}
	return "", nil
}

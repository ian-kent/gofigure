package sources

import (
	"flag"
	"os"
	"reflect"
	"regexp"
	"strings"
)

var flagRe1 = regexp.MustCompile("(.)([A-Z][a-z]+)")
var flagRe2 = regexp.MustCompile("([a-z0-9])([A-Z])")

func camelToFlag(camel string) (flag string) {
	flag = flagRe1.ReplaceAllString(camel, "${1}-${2}")
	flag = flagRe2.ReplaceAllString(flag, "${1}-${2}")
	return strings.ToLower(flag)
}

// CommandLine implements command line configuration using the flag package
type CommandLine struct {
	flags      map[string]*string
	arrayFlags map[string]*arrayValue
	oldCl      *flag.FlagSet
}

type arrayValue []string

func (aV *arrayValue) Set(value string) error {
	printf("Set called for arrayValue: %s", value)
	if *aV == nil {
		*aV = make([]string, 0, 1)
	}
	*aV = append(*aV, value)
	return nil
}

func (aV *arrayValue) String() string {
	return strings.Join(*aV, ", ")
}

// Init is called at the start of a new struct
func (cl *CommandLine) Init(args map[string]string) error {
	cl.flags = make(map[string]*string)
	cl.arrayFlags = make(map[string]*arrayValue)
	cl.oldCl = flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	return nil
}

// Cleanup is called at the end of parsing
func (cl *CommandLine) Cleanup() {
	flag.CommandLine = cl.oldCl
}

// Register is called to register each struct field
func (cl *CommandLine) Register(key, defaultValue string, params map[string]string, t reflect.Type) error {
	if _, ok := cl.flags[key]; ok {
		return ErrKeyExists
	}
	key = camelToFlag(key)

	// TODO validate key?
	// TODO use typed calls instead of StringVar
	printf("Got type %s", t.Kind())
	switch t.Kind() {
	case reflect.Slice:
		printf("Registering slice type for %s", key)
		var val arrayValue
		if len(defaultValue) > 0 {
			val = append(val, defaultValue)
		}
		cl.arrayFlags[key] = &val

		// TODO validate description in some way?
		desc := params["flagDesc"]

		flag.Var(&val, key, desc)
	default:
		printf("Registering default type for %s", key)
		val := defaultValue
		cl.flags[key] = &val

		// TODO validate description in some way?
		desc := params["flagDesc"]

		flag.StringVar(&val, key, defaultValue, desc)
	}

	return nil
}

// Get is called to retrieve a key value
func (cl *CommandLine) Get(key string, overrideDefault *string) (string, error) {
	key = camelToFlag(key)
	printf("Looking up key '%s'", key)

	if !flag.CommandLine.Parsed() {
		flag.Parse()
	}
	// TODO check if flag exists/overrideDefault
	val := ""
	if v, ok := cl.flags[key]; ok {
		printf("Found flag value '%s'", *v)
		val = *v
	}
	if len(val) > 0 {
		printf("Returning val '%s'", val)
		return val, nil
	}
	if overrideDefault != nil {
		printf("Returning overrideDefault '%s'", *overrideDefault)
		return *overrideDefault, nil
	}
	return "", nil
}

func (cl *CommandLine) GetArray(key string, overrideDefault *[]string) ([]string, error) {
	key = camelToFlag(key)
	printf("Looking up array key '%s'", key)

	if !flag.CommandLine.Parsed() {
		flag.Parse()
	}
	// TODO check if flag exists/overrideDefault
	val := []string{}
	if v, ok := cl.arrayFlags[key]; ok {
		printf("Found flag value '%s'", *v)
		val = *v
	}
	if len(val) > 0 {
		printf("Returning val '%s'", val)
		return val, nil
	}
	if overrideDefault != nil {
		printf("Returning overrideDefault '%s'", *overrideDefault)
		return *overrideDefault, nil
	}
	return val, nil
}

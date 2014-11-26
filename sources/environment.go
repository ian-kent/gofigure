package sources

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/ian-kent/envconf"
)

// Environment implements environment variable configuration using envconf
type Environment struct {
	prefix string
	infix  string
	fields map[string]string
}

var camelRe1 = regexp.MustCompile("(.)([A-Z][a-z]+)")
var camelRe2 = regexp.MustCompile("([a-z0-9])([A-Z])")

func camelToSnake(camel string) (snake string) {
	snake = camelRe1.ReplaceAllString(camel, "${1}_${2}")
	snake = camelRe2.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToUpper(snake)
}

// Init is called at the start of a new struct
func (env *Environment) Init(args map[string]string) error {
	env.infix = "_"

	if envPrefix, ok := args["prefix"]; ok {
		env.prefix = envPrefix
	}
	if envInfix, ok := args["infix"]; ok {
		env.infix = envInfix
	}

	env.fields = make(map[string]string)
	return nil
}

// Register is called to register each struct field
func (env *Environment) Register(key, defaultValue string, params map[string]string, t reflect.Type) error {
	env.fields[camelToSnake(key)] = defaultValue
	return nil
}

// Get is called to retrieve a key value
func (env *Environment) Get(key string, overrideDefault *string) (string, error) {
	def := env.fields[key]
	if overrideDefault != nil {
		def = *overrideDefault
	}
	eK := key
	if len(env.prefix) > 0 {
		eK = env.prefix + env.infix + key
	}
	val, err := envconf.FromEnv(eK, def)
	return val.(string), err
}

// Cleanup is called at the end of parsing
func (env *Environment) Cleanup() {

}

package sources

import (
	"reflect"

	"github.com/ian-kent/envconf"
)

// Environment implements environment variable configuration using envconf
type Environment struct {
	prefix string
	infix  string
	fields map[string]string
}

// Init is called at the start of a new struct
func (env *Environment) Init(args map[string]interface{}) error {
	env.infix = "_"

	if envPrefix, ok := args["envPrefix"]; ok {
		env.prefix = envPrefix.(string)
	}
	if envInfix, ok := args["envInfix"]; ok {
		env.infix = envInfix.(string)
	}

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

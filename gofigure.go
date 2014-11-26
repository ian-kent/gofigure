package gofigure

import (
	"errors"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ian-kent/gofigure/sources"
)

// Debug controls log output
var Debug = false

/* TODO
 * - Add file/http sources
 *   - Add "decoders", e.g. json/env/xml
 * - Default value (if gofigure is func()*StructType)
 * - Ignore lowercased "unexported" fields?
 */

// Gofiguration represents a parsed struct
type Gofiguration struct {
	order   []string
	params  map[string]map[string]string
	fields  map[string]*Gofiguritem
	flagged bool
	s       interface{}
}

func (gfg *Gofiguration) printf(message string, args ...interface{}) {
	if !Debug {
		return
	}
	log.Printf(message, args...)
}

// Gofiguritem represents a single struct field
type Gofiguritem struct {
	keys    map[string]string
	field   string
	goField reflect.StructField
	goValue reflect.Value
}

// Sources contains a map of struct field tag names to source implementation
var Sources = map[string]sources.Source{
	"env":  &sources.Environment{},
	"flag": &sources.CommandLine{},
}

// DefaultOrder sets the default order used
var DefaultOrder = []string{"env", "flag"}

var (
	// ReEnvPrefix is used to restrict envPrefix config values
	ReEnvPrefix = regexp.MustCompile("^([A-Z][A-Z0-9_]+|)$")
)

var (
	// ErrInvalidOrder is returned if the "order" struct tag is invalid
	ErrInvalidOrder = errors.New("Invalid order")
	// ErrUnsupportedFieldType is returned for unsupported field types,
	// e.g. chan or func
	ErrUnsupportedFieldType = errors.New("Unsupported field type")
	// ErrInvalidEnvPrefix is returned if the value of envPrefix doesn't
	// match ReEnvPrefix
	ErrInvalidEnvPrefix = errors.New("Invalid environment variable name prefix")
)

// ParseStruct creates a Gofiguration from a struct
func ParseStruct(s interface{}) (*Gofiguration, error) {
	t := reflect.TypeOf(s).Elem()
	v := reflect.ValueOf(s).Elem()

	gfg := &Gofiguration{
		params: make(map[string]map[string]string),
		order:  DefaultOrder,
		fields: make(map[string]*Gofiguritem),
		s:      s,
	}

	err := gfg.parseGofigureField(t)
	if err != nil {
		return nil, err
	}

	gfg.parseFields(v, t)

	return gfg, nil
}

func getStructTags(tag string) map[string]string {
	// http://golang.org/src/pkg/reflect/type.go?s=20885:20906#L747
	m := make(map[string]string)
	for tag != "" {
		// skip leading space
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// scan to colon.
		// a space or a quote is a syntax error
		i = 0
		for i < len(tag) && tag[i] != ' ' && tag[i] != ':' && tag[i] != '"' {
			i++
		}
		if i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// scan quoted string to find value
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		value, _ := strconv.Unquote(qvalue)
		m[name] = value
	}
	return m
}

var argRe = regexp.MustCompile("([a-z]+)([A-Z][a-z]+)")

func (gfg *Gofiguration) parseGofigureField(t reflect.Type) error {
	gf, ok := t.FieldByName("gofigure")
	if ok {
		tags := getStructTags(string(gf.Tag))
		for name, value := range tags {
			if name == "order" {
				oParts := strings.Split(value, ",")
				for _, p := range oParts {
					if _, ok := Sources[p]; !ok {
						return ErrInvalidOrder
					}
				}
				gfg.order = oParts
				continue
			}
			// Parse orderKey:"value" tags, e.g.
			// envPrefix, which gets split into
			//   gfg.params["env"]["prefix"] = "value"
			// gfg.params["env"] is then passed to
			// source registered with that key
			match := argRe.FindStringSubmatch(name)
			if len(match) > 1 {
				if _, ok := gfg.params[match[1]]; !ok {
					gfg.params[match[1]] = make(map[string]string)
				}
				gfg.params[match[1]][strings.ToLower(match[2])] = value
			}
		}
	}
	return nil
}

func (gfg *Gofiguration) parseFields(v reflect.Value, t reflect.Type) {
	gfg.printf("Found %d fields", t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i).Name
		if f == "gofigure" {
			gfg.printf("Skipped field '%s'", f)
			continue
		}

		gfg.printf("Parsed field '%s'", f)

		gfi := &Gofiguritem{
			field:   f,
			goField: t.Field(i),
			goValue: v.Field(i),
			keys:    make(map[string]string),
		}
		tag := t.Field(i).Tag
		if len(tag) > 0 {
			gfi.keys = getStructTags(string(tag))
		}
		gfg.fields[f] = gfi
	}
}

func (gfg *Gofiguration) cleanupSources() {
	for _, o := range gfg.order {
		Sources[o].Cleanup()
	}
}

func (gfg *Gofiguration) initSources() error {
	for _, o := range gfg.order {
		err := Sources[o].Init(gfg.params[o])
		if err != nil {
			return err
		}
	}
	return nil
}

func (gfg *Gofiguration) registerFields() error {
	for _, gfi := range gfg.fields {
		for _, o := range gfg.order {
			gfg.printf("Registering '%s' for source '%s'", gfi.field, o)
			kn := gfi.field
			if k, ok := gfi.keys[o]; ok {
				kn = k
			}
			err := Sources[o].Register(kn, "", gfi.keys, gfi.goField.Type)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (gfg *Gofiguration) populateStruct() error {
	for _, gfi := range gfg.fields {
		var prevVal *string
		for _, o := range gfg.order {
			if prevVal == nil {
				var s = ""
				prevVal = &s
			}
			val, err := Sources[o].Get(gfi.keys[o], prevVal)
			prevVal = &val
			gfg.printf("Got value '%s' from source '%s' for key '%s'", val, gfi.field, gfi.keys[o])
			if err != nil {
				return err
			}

			// FIXME returns ErrUnsupportedFieldType
			// but should it...
			// Two choices:
			// - ignore errors so other fields are parsed
			// - keep errors as "fatal", but support "unexported" fields

			switch gfi.goField.Type.Kind() {
			case reflect.Invalid:
				return ErrUnsupportedFieldType
			case reflect.Bool:
				b, err := strconv.ParseBool(val)
				if err != nil {
					return err
				}
				gfi.goValue.SetBool(b)
			case reflect.Int:
				i, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return err
				}
				gfi.goValue.SetInt(i)
			case reflect.Int8:
				i, err := strconv.ParseInt(val, 10, 8)
				if err != nil {
					return err
				}
				gfi.goValue.SetInt(i)
			case reflect.Int16:
				i, err := strconv.ParseInt(val, 10, 16)
				if err != nil {
					return err
				}
				gfi.goValue.SetInt(i)
			case reflect.Int32:
				i, err := strconv.ParseInt(val, 10, 32)
				if err != nil {
					return err
				}
				gfi.goValue.SetInt(i)
			case reflect.Int64:
				i, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					return err
				}
				gfi.goValue.SetInt(i)
			case reflect.Uint:
				i, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					return err
				}
				gfi.goValue.SetUint(i)
			case reflect.Uint8:
				i, err := strconv.ParseUint(val, 10, 8)
				if err != nil {
					return err
				}
				gfi.goValue.SetUint(i)
			case reflect.Uint16:
				i, err := strconv.ParseUint(val, 10, 16)
				if err != nil {
					return err
				}
				gfi.goValue.SetUint(i)
			case reflect.Uint32:
				i, err := strconv.ParseUint(val, 10, 32)
				if err != nil {
					return err
				}
				gfi.goValue.SetUint(i)
			case reflect.Uint64:
				i, err := strconv.ParseUint(val, 10, 64)
				if err != nil {
					return err
				}
				gfi.goValue.SetUint(i)
			case reflect.Uintptr:
				return ErrUnsupportedFieldType
			case reflect.Float32:
				f, err := strconv.ParseFloat(val, 32)
				if err != nil {
					return err
				}
				gfi.goValue.SetFloat(f)
			case reflect.Float64:
				f, err := strconv.ParseFloat(val, 64)
				if err != nil {
					return err
				}
				gfi.goValue.SetFloat(f)
			case reflect.Complex64:
				return ErrUnsupportedFieldType
			case reflect.Complex128:
				return ErrUnsupportedFieldType
			case reflect.Array:
				// TODO
			case reflect.Chan:
				return ErrUnsupportedFieldType
			case reflect.Func:
				return ErrUnsupportedFieldType
			case reflect.Interface:
			case reflect.Map:
				// TODO
			case reflect.Ptr:
				return ErrUnsupportedFieldType
			case reflect.Slice:
				// TODO
			case reflect.String:
				gfi.goValue.SetString(val)
			case reflect.Struct:
				// TODO
			case reflect.UnsafePointer:
				return ErrUnsupportedFieldType
			default:
				return ErrUnsupportedFieldType
			}
		}
	}
	return nil
}

// Apply applies the Gofiguration to the struct
func (gfg *Gofiguration) Apply(s interface{}) error {
	defer gfg.cleanupSources()

	err := gfg.initSources()
	if err != nil {
		return err
	}

	err = gfg.registerFields()
	if err != nil {
		return err
	}

	return gfg.populateStruct()
}

// Gofigure parses and applies the configuration defined by the struct
func Gofigure(s interface{}) error {
	gfg, err := ParseStruct(s)
	if err != nil {
		return err
	}
	return gfg.Apply(s)
}

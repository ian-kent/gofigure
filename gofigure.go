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

/* TODO
 * - Convert env/cmd to "sources"
 * - Add file/http sources
 *   - Add "decoders", e.g. json/env/xml
 * - Default value (if gofigure is func()*StructType)
 * - Ignore lowercased "unexported" fields?
 */

// Gofiguration represents a parsed struct
type Gofiguration struct {
	envPrefix string
	order     []string
	fields    map[string]*Gofiguritem
	flagged   bool
	s         interface{}
}

func (gfg *Gofiguration) printf(message string, args ...interface{}) {
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
	"env": &sources.Environment{},
	"cmd": &sources.CommandLine{},
}

// DefaultOrder sets the default order used
var DefaultOrder = []string{"env", "cmd"}

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
		envPrefix: "",
		order:     DefaultOrder,
		fields:    make(map[string]*Gofiguritem),
		s:         s,
	}

	err := gfg.parseGofigureField(t)
	if err != nil {
		return nil, err
	}

	gfg.parseFields(v, t)

	return gfg, nil
}

func (gfg *Gofiguration) parseGofigureField(t reflect.Type) error {
	gf, ok := t.FieldByName("gofigure")
	if ok {
		gfg.envPrefix = gf.Tag.Get("envPrefix")
		if ReEnvPrefix.FindAllStringSubmatch(gfg.envPrefix, -1) == nil {
			return ErrInvalidEnvPrefix
		}
		order := gf.Tag.Get("order")
		if len(order) > 0 {
			oParts := strings.Split(order, ",")
			for _, p := range oParts {
				if _, ok := Sources[p]; !ok {
					return ErrInvalidOrder
				}
			}
			gfg.order = oParts
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
			for k := range Sources {
				gfi.keys[k] = tag.Get(k)
			}
		} else {
			// TODO parse CamelCase into CAMEL_CASE and camel-case

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
		err := Sources[o].Init()
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
			err := Sources[o].Register(gfi.keys[o], "", gfi.goField.Type)
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
			case reflect.Int16:
			case reflect.Int32:
			case reflect.Int64:
			case reflect.Uint:
			case reflect.Uint8:
			case reflect.Uint16:
			case reflect.Uint32:
			case reflect.Uint64:
			case reflect.Uintptr:
			case reflect.Float32:
			case reflect.Float64:
			case reflect.Complex64:
			case reflect.Complex128:
			case reflect.Array:
			case reflect.Chan:
				return ErrUnsupportedFieldType
			case reflect.Func:
				return ErrUnsupportedFieldType
			case reflect.Interface:
			case reflect.Map:
			case reflect.Ptr:
			case reflect.Slice:
			case reflect.String:
				gfi.goValue.SetString(val)
			case reflect.Struct:
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

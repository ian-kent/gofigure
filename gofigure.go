// Package gofigure simplifies configuration of Go applications.
//
// Define a struct and call Gofigure()
package gofigure

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ian-kent/gofigure/sources"
)

// Debug controls log output
var Debug = false

func init() {
	env := os.Getenv("GOFIGURE_DEBUG")
	if len(env) > 0 {
		Debug, _ = strconv.ParseBool(env)
	}

	sources.Logger = printf
	sources.Debug = Debug
	return
}

func printf(message string, args ...interface{}) {
	if !Debug {
		return
	}
	log.Printf(message, args...)
}

/* TODO
 * - Add file/http sources
 *   - Add "decoders", e.g. json/env/xml
 * - Default value (if gofigure is func()*StructType)
 * - Ignore lowercased "unexported" fields?
 */

// gofiguration represents a parsed struct
type gofiguration struct {
	order    []string
	params   map[string]map[string]string
	fields   map[string]*gofiguritem
	flagged  bool
	parent   *gofiguration
	children []*gofiguration
	s        interface{}
}

func (gfg *gofiguration) printf(message string, args ...interface{}) {
	printf(message, args...)
}

// gofiguritem represents a single struct field
type gofiguritem struct {
	keys    map[string]string
	field   string
	goField reflect.StructField
	goValue reflect.Value
	inner   *gofiguration
}

// Sources contains a map of struct field tag names to source implementation
var Sources = map[string]sources.Source{
	"env":  &sources.Environment{},
	"flag": &sources.CommandLine{},
}

// DefaultOrder sets the default order used
var DefaultOrder = []string{"env", "flag"}

// ErrUnsupportedType is returned if the interface isn't a
// pointer to a struct
var ErrUnsupportedType = errors.New("Unsupported interface type")

// ErrInvalidOrder is returned if the "order" struct tag is invalid
var ErrInvalidOrder = errors.New("Invalid order")

// ErrUnsupportedFieldType is returned for unsupported field types,
// e.g. chan or func
var ErrUnsupportedFieldType = errors.New("Unsupported field type")

// ParseStruct creates a gofiguration from a struct.
//
// It returns ErrUnsupportedType if s is not a struct or a
// pointer to a struct.
func parseStruct(s interface{}) (*gofiguration, error) {
	var v reflect.Value
	if reflect.TypeOf(s) != reflect.TypeOf(v) {
		v = reflect.ValueOf(s)

		if v.Kind() == reflect.PtrTo(reflect.TypeOf(s)).Kind() {
			v = reflect.Indirect(v)
		}
	} else {
		v = s.(reflect.Value)
	}

	if v.Kind() != reflect.Struct {
		return nil, ErrUnsupportedType
	}

	t := v.Type()

	gfg := &gofiguration{
		params: make(map[string]map[string]string),
		order:  DefaultOrder,
		fields: make(map[string]*gofiguritem),
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

func (gfg *gofiguration) parseGofigureField(t reflect.Type) error {
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

func (gfg *gofiguration) parseFields(v reflect.Value, t reflect.Type) {
	gfg.printf("Found %d fields", t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i).Name
		if f == "gofigure" {
			gfg.printf("Skipped field '%s'", f)
			continue
		}

		gfg.printf("Parsed field '%s'", f)

		gfi := &gofiguritem{
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

func (gfg *gofiguration) cleanupSources() {
	for _, o := range gfg.order {
		Sources[o].Cleanup()
	}
}

func (gfg *gofiguration) initSources() error {
	for _, o := range gfg.order {
		err := Sources[o].Init(gfg.params[o])
		if err != nil {
			return err
		}
	}
	return nil
}

func (gfg *gofiguration) registerFields() error {
	for _, gfi := range gfg.fields {
		kn := gfi.field

		var err error
		switch gfi.goField.Type.Kind() {
		case reflect.Struct:
			gfg.printf("Registering as struct type")
			// TODO do shit
			sGfg, err := parseStruct(gfi.goValue)
			if err != nil {
				return err
			}
			sGfg.apply(gfg)
			gfi.inner = sGfg
		default:
			gfg.printf("Registering as default type")
			for _, o := range gfg.order {
				if k, ok := gfi.keys[o]; ok {
					kn = k
				}
				gfg.printf("Registering '%s' for source '%s' with key '%s'", gfi.field, o, kn)
				err = Sources[o].Register(kn, "", gfi.keys, gfi.goField.Type)
			}
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func numVal(i string) string {
	if len(i) == 0 {
		return "0"
	}
	return i
}

func (gfi *gofiguritem) populateDefaultType(order []string) error {
	// FIXME could just preserve types
	var v string
	switch gfi.goField.Type.Kind() {
	case reflect.Bool:
		v = fmt.Sprintf("%t", gfi.goValue.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = fmt.Sprintf("%d", gfi.goValue.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = fmt.Sprintf("%d", gfi.goValue.Uint())
	case reflect.Float32, reflect.Float64:
		v = fmt.Sprintf("%f", gfi.goValue.Float())
	case reflect.String:
		v = gfi.goValue.String()
	}

	var prevVal = &v

	for _, source := range order {
		kn := gfi.field
		if k, ok := gfi.keys[source]; ok {
			kn = k
		}

		val, err := Sources[source].Get(kn, prevVal)
		if err != nil {
			return err
		}

		prevVal = &val

		printf("Got value '%s' from source '%s' for key '%s'", val, source, gfi.field)

		switch gfi.goField.Type.Kind() {
		case reflect.Bool:
			if len(val) == 0 {
				printf("Setting bool value to false")
				val = "false"
			}
			b, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			gfi.goValue.SetBool(b)
		case reflect.Int:
			i, err := strconv.ParseInt(numVal(val), 10, 64)
			if err != nil {
				return err
			}
			gfi.goValue.SetInt(i)
		case reflect.Int8:
			i, err := strconv.ParseInt(numVal(val), 10, 8)
			if err != nil {
				return err
			}
			gfi.goValue.SetInt(i)
		case reflect.Int16:
			i, err := strconv.ParseInt(numVal(val), 10, 16)
			if err != nil {
				return err
			}
			gfi.goValue.SetInt(i)
		case reflect.Int32:
			i, err := strconv.ParseInt(numVal(val), 10, 32)
			if err != nil {
				return err
			}
			gfi.goValue.SetInt(i)
		case reflect.Int64:
			i, err := strconv.ParseInt(numVal(val), 10, 64)
			if err != nil {
				return err
			}
			gfi.goValue.SetInt(i)
		case reflect.Uint:
			i, err := strconv.ParseUint(numVal(val), 10, 64)
			if err != nil {
				return err
			}
			gfi.goValue.SetUint(i)
		case reflect.Uint8:
			i, err := strconv.ParseUint(numVal(val), 10, 8)
			if err != nil {
				return err
			}
			gfi.goValue.SetUint(i)
		case reflect.Uint16:
			i, err := strconv.ParseUint(numVal(val), 10, 16)
			if err != nil {
				return err
			}
			gfi.goValue.SetUint(i)
		case reflect.Uint32:
			i, err := strconv.ParseUint(numVal(val), 10, 32)
			if err != nil {
				return err
			}
			gfi.goValue.SetUint(i)
		case reflect.Uint64:
			i, err := strconv.ParseUint(numVal(val), 10, 64)
			if err != nil {
				return err
			}
			gfi.goValue.SetUint(i)
		case reflect.Float32:
			f, err := strconv.ParseFloat(numVal(val), 32)
			if err != nil {
				return err
			}
			gfi.goValue.SetFloat(f)
		case reflect.Float64:
			f, err := strconv.ParseFloat(numVal(val), 64)
			if err != nil {
				return err
			}
			gfi.goValue.SetFloat(f)
		case reflect.String:
			gfi.goValue.SetString(val)
		default:
			return ErrUnsupportedFieldType
		}
	}

	return nil
}

func (gfi *gofiguritem) populateSliceType(order []string) error {
	var prevVal *[]string

	for _, source := range order {
		kn := gfi.field
		if k, ok := gfi.keys[source]; ok {
			kn = k
		}

		printf("Looking for field '%s' with key '%s' in source '%s'", gfi.field, kn, source)
		val, err := Sources[source].GetArray(kn, prevVal)
		if err != nil {
			return err
		}

		// This causes duplication between array sources depending on order
		//prevVal = &val

		printf("Got value '%+v' from array source '%s' for key '%s'", val, source, gfi.field)

		switch gfi.goField.Type.Kind() {
		case reflect.Slice:
			switch gfi.goField.Type.Elem().Kind() {
			case reflect.String:
				for _, s := range val {
					printf("Appending string value '%s' to slice", s)
					gfi.goValue.Set(reflect.Append(gfi.goValue, reflect.ValueOf(s)))
				}
			case reflect.Int:
				for _, s := range val {
					printf("Appending int value '%s' to slice", s)
					i, err := strconv.ParseInt(numVal(s), 10, 64)
					if err != nil {
						return err
					}
					gfi.goValue.Set(reflect.Append(gfi.goValue, reflect.ValueOf(int(i))))
				}
			case reflect.Int8:
				for _, s := range val {
					printf("Appending int8 value '%s' to slice", s)
					i, err := strconv.ParseInt(numVal(s), 10, 8)
					if err != nil {
						return err
					}
					gfi.goValue.Set(reflect.Append(gfi.goValue, reflect.ValueOf(int8(i))))
				}
			case reflect.Int16:
				for _, s := range val {
					printf("Appending int16 value '%s' to slice", s)
					i, err := strconv.ParseInt(numVal(s), 10, 16)
					if err != nil {
						return err
					}
					gfi.goValue.Set(reflect.Append(gfi.goValue, reflect.ValueOf(int16(i))))
				}
			case reflect.Int32:
				for _, s := range val {
					printf("Appending int32 value '%s' to slice", s)
					i, err := strconv.ParseInt(numVal(s), 10, 32)
					if err != nil {
						return err
					}
					gfi.goValue.Set(reflect.Append(gfi.goValue, reflect.ValueOf(int32(i))))
				}
			case reflect.Int64:
				for _, s := range val {
					printf("Appending int64 value '%s' to slice", s)
					i, err := strconv.ParseInt(numVal(s), 10, 64)
					if err != nil {
						return err
					}
					gfi.goValue.Set(reflect.Append(gfi.goValue, reflect.ValueOf(int64(i))))
				}
			case reflect.Uint:
				for _, s := range val {
					printf("Appending uint value '%s' to slice", s)
					i, err := strconv.ParseUint(numVal(s), 10, 64)
					if err != nil {
						return err
					}
					gfi.goValue.Set(reflect.Append(gfi.goValue, reflect.ValueOf(uint(i))))
				}
			case reflect.Uint8:
				for _, s := range val {
					printf("Appending uint8 value '%s' to slice", s)
					i, err := strconv.ParseUint(numVal(s), 10, 8)
					if err != nil {
						return err
					}
					gfi.goValue.Set(reflect.Append(gfi.goValue, reflect.ValueOf(uint8(i))))
				}
			case reflect.Uint16:
				for _, s := range val {
					printf("Appending uint16 value '%s' to slice", s)
					i, err := strconv.ParseUint(numVal(s), 10, 16)
					if err != nil {
						return err
					}
					gfi.goValue.Set(reflect.Append(gfi.goValue, reflect.ValueOf(uint16(i))))
				}
			case reflect.Uint32:
				for _, s := range val {
					printf("Appending uint32 value '%s' to slice", s)
					i, err := strconv.ParseUint(numVal(s), 10, 32)
					if err != nil {
						return err
					}
					gfi.goValue.Set(reflect.Append(gfi.goValue, reflect.ValueOf(uint32(i))))
				}
			case reflect.Uint64:
				for _, s := range val {
					printf("Appending uint64 value '%s' to slice", s)
					i, err := strconv.ParseUint(numVal(s), 10, 64)
					if err != nil {
						return err
					}
					gfi.goValue.Set(reflect.Append(gfi.goValue, reflect.ValueOf(uint64(i))))
				}
			// TODO floats
			default:
				//return ErrUnsupportedFieldType
			}
		}
	}

	return nil
}

func (gfi *gofiguritem) populateStructType(order []string) error {
	return gfi.inner.populateStruct()
}

func (gfg *gofiguration) populateStruct() error {
	if gfg == nil {
		return nil
	}

	for _, gfi := range gfg.fields {
		printf("Populating field %s", gfi.field)
		switch gfi.goField.Type.Kind() {
		case reflect.Invalid, reflect.Uintptr, reflect.Complex64,
			reflect.Complex128, reflect.Chan, reflect.Func,
			reflect.Ptr, reflect.UnsafePointer:
			return ErrUnsupportedFieldType
		case reflect.Interface:
			// TODO
			return ErrUnsupportedFieldType
		case reflect.Map:
			// TODO
			return ErrUnsupportedFieldType
		case reflect.Slice:
			printf("Calling populateSliceType")
			err := gfi.populateSliceType(gfg.order)
			if err != nil {
				return err
			}
		case reflect.Struct:
			printf("Calling populateStructType")
			err := gfi.populateStructType(gfg.order)
			if err != nil {
				return err
			}
		case reflect.Array:
			// TODO
			return ErrUnsupportedFieldType
		default:
			printf("Calling populateDefaultType")
			err := gfi.populateDefaultType(gfg.order)
			if err != nil {
				return err
			}
		}
	}

	for _, c := range gfg.children {
		err := c.populateStruct()
		if err != nil {
			return err
		}
	}

	return nil
}

// Apply applies the gofiguration to the struct
func (gfg *gofiguration) apply(parent *gofiguration) error {
	gfg.parent = parent

	if parent == nil {
		defer gfg.cleanupSources()

		err := gfg.initSources()
		if err != nil {
			return err
		}
	} else {
		parent.children = append(parent.children, gfg)
	}

	err := gfg.registerFields()
	if err != nil {
		return err
	}

	if parent == nil {
		return gfg.populateStruct()
	}

	return nil
}

// Gofigure parses and applies the configuration defined by the struct.
//
// It returns ErrUnsupportedType if s is not a pointer to a struct.
func Gofigure(s interface{}) error {
	gfg, err := parseStruct(s)
	if err != nil {
		return err
	}
	return gfg.apply(nil)
}

package gofigure

import (
	"os"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/ian-kent/gofigure/sources"
)

// TODO
// - flagDesc
// - flag and env default keys

// MyConfigFoo is a basic test struct
type MyConfigFoo struct {
	gofigure interface{} `envPrefix:"FOO" order:"env,flag"`
	BindAddr string      `env:"BIND_ADDR" flag:"bind-addr"`
}

// MyConfigBar is a basic test struct with multiple fields
type MyConfigBar struct {
	gofigure   interface{} `envPrefix:"BAR" order:"flag,env"`
	RemoteAddr string      `env:"REMOTE_ADDR" flag:"remote-addr"`
	LocalAddr  string      `env:"LOCAL_ADDR" flag:"local-addr"`
}

// MyConfigBaz is used to test invalid order values
type MyConfigBaz struct {
	gofigure interface{} `order:"FOO,BAR"`
}

// MyConfigBay is used to test invalid envPrefix values
type MyConfigBay struct {
	gofigure interface{} `envPrefix:"!"`
}

// MyConfigFull is used to test Go type support
type MyConfigFull struct {
	gofigure         interface{}
	BoolField        bool
	IntField         int
	Int8Field        int8
	Int16Field       int16
	Int32Field       int32
	Int64Field       int64
	UintField        uint
	Uint8Field       uint8
	Uint16Field      uint16
	Uint32Field      uint32
	Uint64Field      uint64
	Float32Field     float32
	Float64Field     float64
	ArrayIntField    []int
	ArrayStringField []string
}

func TestParseStruct(t *testing.T) {
	Convey("ParseStruct should keep a reference to the struct", t, func() {
		ref := &MyConfigFoo{}
		info, e := ParseStruct(ref)
		So(e, ShouldBeNil)
		So(info, ShouldNotBeNil)
		So(info.s, ShouldEqual, ref)
	})

	Convey("ParseStruct should read gofigure envPrefix tag correctly", t, func() {
		info, e := ParseStruct(&MyConfigFoo{})
		So(e, ShouldBeNil)
		So(info, ShouldNotBeNil)
		So(info.params["env"]["prefix"], ShouldEqual, "FOO")

		info, e = ParseStruct(&MyConfigBar{})
		So(e, ShouldBeNil)
		So(info, ShouldNotBeNil)
		So(info.params["env"]["prefix"], ShouldEqual, "BAR")
	})

	Convey("ParseStruct should read gofigure order tag correctly", t, func() {
		info, e := ParseStruct(&MyConfigFoo{})
		So(e, ShouldBeNil)
		So(info, ShouldNotBeNil)
		So(info.order, ShouldResemble, []string{"env", "flag"})

		info, e = ParseStruct(&MyConfigBar{})
		So(e, ShouldBeNil)
		So(info, ShouldNotBeNil)
		So(info.order, ShouldResemble, []string{"flag", "env"})
	})

	Convey("Invalid order should return error", t, func() {
		info, e := ParseStruct(&MyConfigBaz{})
		So(e, ShouldNotBeNil)
		So(e, ShouldEqual, ErrInvalidOrder)
		So(info, ShouldBeNil)
	})

	Convey("ParseStruct should read fields correctly", t, func() {
		info, e := ParseStruct(&MyConfigFoo{})
		So(e, ShouldBeNil)
		So(info, ShouldNotBeNil)
		So(len(info.fields), ShouldEqual, 1)
		_, ok := info.fields["BindAddr"]
		So(ok, ShouldEqual, true)
		So(info.fields["BindAddr"].field, ShouldEqual, "BindAddr")
		So(info.fields["BindAddr"].keys["env"], ShouldEqual, "BIND_ADDR")
		So(info.fields["BindAddr"].keys["flag"], ShouldEqual, "bind-addr")
		So(info.fields["BindAddr"].goField, ShouldNotBeNil)
		So(info.fields["BindAddr"].goField.Type.Kind(), ShouldEqual, reflect.String)
		So(info.fields["BindAddr"].goValue, ShouldNotBeNil)

		info, e = ParseStruct(&MyConfigBar{})
		So(e, ShouldBeNil)
		So(info, ShouldNotBeNil)
		So(len(info.fields), ShouldEqual, 2)
		_, ok = info.fields["RemoteAddr"]
		So(ok, ShouldEqual, true)
		So(info.fields["RemoteAddr"].field, ShouldEqual, "RemoteAddr")
		So(info.fields["RemoteAddr"].keys["env"], ShouldEqual, "REMOTE_ADDR")
		So(info.fields["RemoteAddr"].keys["flag"], ShouldEqual, "remote-addr")
		So(info.fields["RemoteAddr"].goField, ShouldNotBeNil)
		So(info.fields["RemoteAddr"].goField.Type.Kind(), ShouldEqual, reflect.String)
		So(info.fields["RemoteAddr"].goValue, ShouldNotBeNil)
		_, ok = info.fields["LocalAddr"]
		So(ok, ShouldEqual, true)
		So(info.fields["LocalAddr"].field, ShouldEqual, "LocalAddr")
		So(info.fields["LocalAddr"].keys["env"], ShouldEqual, "LOCAL_ADDR")
		So(info.fields["LocalAddr"].keys["flag"], ShouldEqual, "local-addr")
		So(info.fields["LocalAddr"].goField, ShouldNotBeNil)
		So(info.fields["LocalAddr"].goField.Type.Kind(), ShouldEqual, reflect.String)
		So(info.fields["LocalAddr"].goValue, ShouldNotBeNil)
	})
}

func TestGofigure(t *testing.T) {
	Convey("Gofigure should set field values", t, func() {
		os.Clearenv()
		os.Args = []string{"gofigure", "-bind-addr", "abcdef"}
		var cfg MyConfigFoo
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.BindAddr, ShouldEqual, "abcdef")
	})

	Convey("Gofigure should set multiple field values", t, func() {
		os.Clearenv()
		os.Args = []string{"gofigure", "-remote-addr", "foo", "-local-addr", "bar"}
		var cfg2 MyConfigBar
		err := Gofigure(&cfg2)
		So(err, ShouldBeNil)
		So(cfg2, ShouldNotBeNil)
		So(cfg2.RemoteAddr, ShouldEqual, "foo")
		So(cfg2.LocalAddr, ShouldEqual, "bar")
	})

	Convey("Gofigure should support environment variables", t, func() {
		os.Clearenv()
		os.Args = []string{"gofigure"}
		os.Setenv("FOO_BIND_ADDR", "bindaddr")
		var cfg MyConfigFoo
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.BindAddr, ShouldEqual, "bindaddr")
	})

	Convey("Gofigure should preserve order", t, func() {
		os.Clearenv()
		os.Args = []string{"gofigure", "-bind-addr", "abc"}
		os.Setenv("FOO_BIND_ADDR", "def")
		var cfg MyConfigFoo
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.BindAddr, ShouldEqual, "abc")

		os.Clearenv()
		os.Args = []string{"gofigure", "-remote-addr", "abc"}
		os.Setenv("BAR_REMOTE_ADDR", "def")
		var cfg2 MyConfigBar
		err = Gofigure(&cfg2)
		So(err, ShouldBeNil)
		So(cfg2, ShouldNotBeNil)
		So(cfg2.RemoteAddr, ShouldEqual, "def")
	})
}

func TestBoolField(t *testing.T) {
	Convey("Can set a bool field to true (flag)", t, func() {
		os.Clearenv()
		os.Args = []string{
			"gofigure",
			"-bool-field", "true",
		}
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.BoolField, ShouldEqual, true)
	})

	Convey("Can set a bool field to false (flag)", t, func() {
		os.Clearenv()
		os.Args = []string{
			"gofigure",
			"-bool-field", "false",
		}
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.BoolField, ShouldEqual, false)
	})

	Convey("Can set a bool field to true (env)", t, func() {
		os.Clearenv()
		os.Args = []string{
			"gofigure",
		}
		os.Setenv("BOOL_FIELD", "true")
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.BoolField, ShouldEqual, true)
	})

	Convey("Can set a bool field to false (env)", t, func() {
		os.Clearenv()
		os.Args = []string{
			"gofigure",
		}
		os.Setenv("BOOL_FIELD", "false")
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.BoolField, ShouldEqual, false)
	})

	Convey("Not setting a bool field gives false", t, func() {
		os.Clearenv()
		os.Args = []string{
			"gofigure",
		}
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.BoolField, ShouldEqual, false)
	})
}

func TestIntField(t *testing.T) {
	Convey("Can set int fields (flag)", t, func() {
		os.Clearenv()
		os.Args = []string{
			"gofigure",
			"-int-field", "123",
			"-int8-field", "2",
			"-int16-field", "10",
			"-int32-field", "33",
			"-int64-field", "81",
		}
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.IntField, ShouldEqual, 123)
		So(cfg.Int8Field, ShouldEqual, 2)
		So(cfg.Int16Field, ShouldEqual, 10)
		So(cfg.Int32Field, ShouldEqual, 33)
		So(cfg.Int64Field, ShouldEqual, 81)
	})

	Convey("Can set int fields to negative values (flag)", t, func() {
		os.Args = []string{
			"gofigure",
			"-int-field", "-123",
			"-int8-field", "-2",
			"-int16-field", "-10",
			"-int32-field", "-33",
			"-int64-field", "-81",
		}
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.IntField, ShouldEqual, -123)
		So(cfg.Int8Field, ShouldEqual, -2)
		So(cfg.Int16Field, ShouldEqual, -10)
		So(cfg.Int32Field, ShouldEqual, -33)
		So(cfg.Int64Field, ShouldEqual, -81)
	})

	Convey("Can set int fields (env)", t, func() {
		os.Clearenv()
		os.Setenv("INT_FIELD", "123")
		os.Setenv("INT8_FIELD", "2")
		os.Setenv("INT16_FIELD", "10")
		os.Setenv("INT32_FIELD", "33")
		os.Setenv("INT64_FIELD", "81")
		os.Args = []string{
			"gofigure",
		}
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.IntField, ShouldEqual, 123)
		So(cfg.Int8Field, ShouldEqual, 2)
		So(cfg.Int16Field, ShouldEqual, 10)
		So(cfg.Int32Field, ShouldEqual, 33)
		So(cfg.Int64Field, ShouldEqual, 81)
	})

	Convey("Can set int fields to negative values (env)", t, func() {
		os.Clearenv()
		os.Args = []string{
			"gofigure",
		}
		os.Setenv("INT_FIELD", "-123")
		os.Setenv("INT8_FIELD", "-2")
		os.Setenv("INT16_FIELD", "-10")
		os.Setenv("INT32_FIELD", "-33")
		os.Setenv("INT64_FIELD", "-81")
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.IntField, ShouldEqual, -123)
		So(cfg.Int8Field, ShouldEqual, -2)
		So(cfg.Int16Field, ShouldEqual, -10)
		So(cfg.Int32Field, ShouldEqual, -33)
		So(cfg.Int64Field, ShouldEqual, -81)
	})

	Convey("Unset int fields are 0", t, func() {
		os.Clearenv()
		os.Args = []string{
			"gofigure",
		}
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.IntField, ShouldEqual, 0)
		So(cfg.Int8Field, ShouldEqual, 0)
		So(cfg.Int16Field, ShouldEqual, 0)
		So(cfg.Int32Field, ShouldEqual, 0)
		So(cfg.Int64Field, ShouldEqual, 0)
	})
}

func TestUintField(t *testing.T) {
	Convey("Can set uint fields (flag)", t, func() {
		os.Clearenv()
		os.Args = []string{
			"gofigure",
			"-uint-field", "123",
			"-uint8-field", "2",
			"-uint16-field", "10",
			"-uint32-field", "33",
			"-uint64-field", "81",
		}
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.UintField, ShouldEqual, 123)
		So(cfg.Uint8Field, ShouldEqual, 2)
		So(cfg.Uint16Field, ShouldEqual, 10)
		So(cfg.Uint32Field, ShouldEqual, 33)
		So(cfg.Uint64Field, ShouldEqual, 81)
	})

	Convey("Can set int fields (env)", t, func() {
		os.Clearenv()
		os.Setenv("UINT_FIELD", "123")
		os.Setenv("UINT8_FIELD", "2")
		os.Setenv("UINT16_FIELD", "10")
		os.Setenv("UINT32_FIELD", "33")
		os.Setenv("UINT64_FIELD", "81")
		os.Args = []string{
			"gofigure",
		}
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.UintField, ShouldEqual, 123)
		So(cfg.Uint8Field, ShouldEqual, 2)
		So(cfg.Uint16Field, ShouldEqual, 10)
		So(cfg.Uint32Field, ShouldEqual, 33)
		So(cfg.Uint64Field, ShouldEqual, 81)
	})

	Convey("Unset uint fields are 0", t, func() {
		os.Clearenv()
		os.Args = []string{
			"gofigure",
		}
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.UintField, ShouldEqual, 0)
		So(cfg.Uint8Field, ShouldEqual, 0)
		So(cfg.Uint16Field, ShouldEqual, 0)
		So(cfg.Uint32Field, ShouldEqual, 0)
		So(cfg.Uint64Field, ShouldEqual, 0)
	})
}

func TestArrayField(t *testing.T) {
	Debug = true
	sources.Debug = true
	Convey("String array should work", t, func() {
		os.Clearenv()
		os.Args = []string{
			"gofigure",
			"-array-string-field", "foo",
			"-array-string-field", "bar",
		}
		var cfg MyConfigFull
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.ArrayStringField, ShouldNotBeNil)
		So(len(cfg.ArrayStringField), ShouldEqual, 2)
	})
}

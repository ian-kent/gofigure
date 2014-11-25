package gofigure

import (
	"os"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// MyConfigFoo is a basic test struct
type MyConfigFoo struct {
	gofigure interface{} `envPrefix:"FOO" order:"env,cmd"`
	BindAddr string      `env:"BIND_ADDR" cmd:"bind-addr"`
}

// MyConfigBar is a basic test struct with multiple fields
type MyConfigBar struct {
	gofigure   interface{} `envPrefix:"BAR" order:"cmd,env"`
	RemoteAddr string      `env:"REMOTE_ADDR" cmd:"remote-addr"`
	LocalAddr  string      `env:"LOCAL_ADDR" cmd:"local-addr"`
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
	UintptrField     uintptr
	Float32Field     float32
	Float64Field     float64
	Complex64Field   complex64
	Complex128Field  complex128
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
		So(info.envPrefix, ShouldEqual, "FOO")

		info, e = ParseStruct(&MyConfigBar{})
		So(e, ShouldBeNil)
		So(info, ShouldNotBeNil)
		So(info.envPrefix, ShouldEqual, "BAR")
	})

	Convey("ParseStruct should read gofigure order tag correctly", t, func() {
		info, e := ParseStruct(&MyConfigFoo{})
		So(e, ShouldBeNil)
		So(info, ShouldNotBeNil)
		So(info.order, ShouldResemble, []string{"env", "cmd"})

		info, e = ParseStruct(&MyConfigBar{})
		So(e, ShouldBeNil)
		So(info, ShouldNotBeNil)
		So(info.order, ShouldResemble, []string{"cmd", "env"})
	})

	Convey("Invalid order should return error", t, func() {
		info, e := ParseStruct(&MyConfigBaz{})
		So(e, ShouldNotBeNil)
		So(e, ShouldEqual, ErrInvalidOrder)
		So(info, ShouldBeNil)
	})

	Convey("Invalid envPrefix should return error", t, func() {
		info, e := ParseStruct(&MyConfigBay{})
		So(e, ShouldNotBeNil)
		So(e, ShouldEqual, ErrInvalidEnvPrefix)
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
		So(info.fields["BindAddr"].keys["cmd"], ShouldEqual, "bind-addr")
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
		So(info.fields["RemoteAddr"].keys["cmd"], ShouldEqual, "remote-addr")
		So(info.fields["RemoteAddr"].goField, ShouldNotBeNil)
		So(info.fields["RemoteAddr"].goField.Type.Kind(), ShouldEqual, reflect.String)
		So(info.fields["RemoteAddr"].goValue, ShouldNotBeNil)
		_, ok = info.fields["LocalAddr"]
		So(ok, ShouldEqual, true)
		So(info.fields["LocalAddr"].field, ShouldEqual, "LocalAddr")
		So(info.fields["LocalAddr"].keys["env"], ShouldEqual, "LOCAL_ADDR")
		So(info.fields["LocalAddr"].keys["cmd"], ShouldEqual, "local-addr")
		So(info.fields["LocalAddr"].goField, ShouldNotBeNil)
		So(info.fields["LocalAddr"].goField.Type.Kind(), ShouldEqual, reflect.String)
		So(info.fields["LocalAddr"].goValue, ShouldNotBeNil)
	})
}

func TestGofigure(t *testing.T) {
	Convey("Gofigure should set field values", t, func() {
		os.Args = []string{"gofigure", "-bind-addr", "abcdef"}
		var cfg MyConfigFoo
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.BindAddr, ShouldEqual, "abcdef")
	})

	Convey("Gofigure should set multiple field values", t, func() {
		os.Args = []string{"gofigure", "-remote-addr", "foo", "-local-addr", "bar"}
		var cfg2 MyConfigBar
		err := Gofigure(&cfg2)
		So(err, ShouldBeNil)
		So(cfg2, ShouldNotBeNil)
		So(cfg2.RemoteAddr, ShouldEqual, "foo")
		So(cfg2.LocalAddr, ShouldEqual, "bar")
	})

	Convey("Gofigure should support environment variables", t, func() {
		os.Args = []string{"gofigure"}
		os.Setenv("FOO_BIND_ADDR", "bindaddr")
		var cfg MyConfigFoo
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.BindAddr, ShouldEqual, "bindaddr")
	})

	Convey("Gofigure should preserve order", t, func() {
		os.Args = []string{"gofigure", "-bind-addr", "abc"}
		os.Setenv("FOO_BIND_ADDR", "def")
		var cfg MyConfigFoo
		err := Gofigure(&cfg)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg.BindAddr, ShouldEqual, "abc")

		os.Args = []string{"gofigure", "-remote-addr", "abc"}
		os.Setenv("BAR_REMOTE_ADDR", "def")
		var cfg2 MyConfigBar
		err = Gofigure(&cfg2)
		So(err, ShouldBeNil)
		So(cfg2, ShouldNotBeNil)
		So(cfg2.RemoteAddr, ShouldEqual, "def")
	})
}

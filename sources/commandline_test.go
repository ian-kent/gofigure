package sources

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCamelToFlag(t *testing.T) {
	Convey("camelToFlag converts CamelCase to flag-case", t, func() {
		So(camelToFlag("CamelCase"), ShouldEqual, "camel-case")
		So(camelToFlag("camelCase"), ShouldEqual, "camel-case")
		So(camelToFlag("CaMeLCase"), ShouldEqual, "ca-me-l-case")
	})
}

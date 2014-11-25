package sources

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCamelToSnake(t *testing.T) {
	Convey("camelToSnake converts CamelCase to snake_case", t, func() {
		So(camelToSnake("CamelCase"), ShouldEqual, "CAMEL_CASE")
		So(camelToSnake("camelCase"), ShouldEqual, "CAMEL_CASE")
		So(camelToSnake("CaMeLCase"), ShouldEqual, "CA_ME_L_CASE")
	})
}

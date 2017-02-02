package models

import (
	"github.com/jinzhu/gorm"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSpec(t *testing.T) {

	Convey("Models Tests", t, func() {

		Convey("should connect", func() {
			db, err := Connect()

			So(err, ShouldBeNil)
			So(db, ShouldHaveSameTypeAs, gorm.DB{})
		})
	})
}

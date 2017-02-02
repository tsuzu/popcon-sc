package vault

import (
	"github.com/derekdowling/bursa/models"
	"github.com/derekdowling/bursa/testutils"
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"testing"
)

func TestSpec(t *testing.T) {
	// TODO Doing this all over is ugly.
	db, err := models.Connect()
	if err != nil {
		log.Fatalf("Couldn't connect to database during testing", err)
	}

	Convey("Vault Tests", t, func() {
		Convey("NewMaster()", func() {
			Convey("Should generate a new HD Master Key.", func() {
				user := models.User{
					Name: testutils.SuffixedId("hd_key"),
				}
				So(db.Save(&user).Error, ShouldBeNil)

				new_master, err := NewMasterForUser(user.Id)
				So(err, ShouldBeNil)
				So(new_master, ShouldHaveSameTypeAs, "")
			})
		})

		Convey("GetEncodedAddress()", func() {
			Convey("Should convert a private key to a hash.", func() {
				// master, _ := NewMaster()
				// So(master, ShouldHaveSameTypeAs, "")

				// public_key = GetEncodedAddress(master)
				// So(public_key, ShouldNotEqual, master)
			})
		})
	})
}

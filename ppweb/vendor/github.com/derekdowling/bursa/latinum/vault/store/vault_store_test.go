package store

import (
	"github.com/derekdowling/bursa/models"
	"github.com/derekdowling/bursa/testutils"
	"log"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSpec(t *testing.T) {
	// TODO Doing this all over is ugly.
	db, err := models.Connect()
	if err != nil {
		log.Fatalf("Couldn't connect to database during testing", err)
	}

	Convey("Vault Store Tests", t, func() {
		Convey("Store()", func() {
			Convey("Should store a key", func() {
				user := models.User{
					Name: testutils.SuffixedId("vault_store"),
				}
				key := "makethisavalidbase58key"

				So(db.Save(&user).Error, ShouldBeNil)
				So(Store(user.Id, key), ShouldBeNil)
			})

			Convey("Should retrieve a stored key", func() {
				user := models.User{
					Name: testutils.SuffixedId("vault_retrieve"),
				}

				So(db.Save(&user).Error, ShouldBeNil)
				So(Store(user.Id, "makethisavalidbase58key"), ShouldBeNil)

				retrieved_key, err := Retrieve(user.Id)
				So(err, ShouldBeNil)
				So(retrieved_key, ShouldEqual, "makethisavalidbase58key")
			})
		})
	})
}

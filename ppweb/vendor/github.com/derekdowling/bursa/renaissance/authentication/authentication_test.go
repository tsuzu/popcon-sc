package authentication

import (
	"code.google.com/p/go.crypto/bcrypt"
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"strings"
	"testing"
)

func TestSpec(t *testing.T) {

	Convey("Authentication Testing", t, func() {

		Convey("generateSalt()", func() {
			salt := generateSalt()
			So(salt, ShouldNotBeBlank)
			So(len(salt), ShouldEqual, SaltLength)
		})

		Convey("combine()", func() {
			salt := generateSalt()
			password := "boomchuckalucka"
			expectedLength := len(salt) + len(password)
			combo := combine(salt, password)

			So(combo, ShouldNotBeBlank)
			So(len(combo), ShouldEqual, expectedLength)
			So(strings.HasPrefix(combo, salt), ShouldBeTrue)
		})

		Convey("hashPassword()", func() {
			combo := combine(generateSalt(), "hershmahgersh")
			hash := hashPassword(combo)
			So(hash, ShouldNotBeBlank)

			cost, err := bcrypt.Cost([]byte(hash))
			if err != nil {
				log.Print(err)
			}
			So(cost, ShouldEqual, EncryptCost)
		})

		Convey("CreatePassword()", func() {
			passString := "mmmPassword1"
			salt, hash := CreatePassword(passString)

			So(salt, ShouldNotBeBlank)
			So(hash, ShouldNotBeBlank)
			So(len(salt), ShouldEqual, SaltLength)
		})

		Convey("PasswordMatch()", func() {
			password := "megaman49"
			salt, hash := CreatePassword(password)

			So(PasswordMatch("megaman49", salt, hash), ShouldBeTrue)
			So(PasswordMatch("lolfail", salt, hash), ShouldBeFalse)
			So(PasswordMatch("MegAman49", salt, hash), ShouldBeFalse)
		})
	})
}

package mailchimp

import (
	"github.com/derekdowling/bursa/testutils"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSpec(t *testing.T) {

	Convey("Mailchimp Tests", t, func() {

		email := testutils.TestEmail("mailchimp")

		Convey("SubscribeToChimp()", func() {
			result := SubscribeToChimp(email)
			So(result, ShouldBeTrue)
		})
	})
}

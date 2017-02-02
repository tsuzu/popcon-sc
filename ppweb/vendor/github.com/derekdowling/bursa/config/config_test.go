package config

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestSpec(t *testing.T) {

	Convey("Config Tests", t, func() {

		Convey("load()", func() {

			Convey("Environments", func() {
				original_env := os.Getenv("BURSA_ENV")

				Convey("With Production Environment", func() {
					os.Setenv("BURSA_ENV", "production")
					So(getConfigPaths(), ShouldResemble, []string{"server.yml", "server.production.yml", "database.yml"})
					config := load()
					So(config.GetString("logging.mode"), ShouldEqual, "prod")
				})

				Convey("With Development Environment", func() {
					os.Setenv("BURSA_ENV", "development")
					So(getConfigPaths(), ShouldResemble, []string{"server.yml", "server.development.yml", "database.yml"})
					config := load()
					So(config.GetString("logging.mode"), ShouldEqual, "dev")
				})

				Convey("With A Blank Environment", func() {
					os.Setenv("BURSA_ENV", "")
					config := load()
					So(config.GetString("logging.mode"), ShouldEqual, "dev")
				})

				os.Setenv("BURSA_ENV", original_env)
			})
		})
	})
}

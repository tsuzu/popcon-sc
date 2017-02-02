// A package for retrying things repeatedly until they work.
package klutz

import(
	"log"
	"time"
)

type closure func() error

// Calls a closure up to `times` times, waiting `sleep_time` in between calls.
func Flail(times int, sleep_time time.Duration, fn closure) error {
	var err error = nil
	for i := 0; i < times; i++ {
		if err = fn(); err != nil {
			log.Println("Function returned error, flailing", err)
			time.Sleep(sleep_time)
		} else {
			return nil
		}
	}
	return err
}

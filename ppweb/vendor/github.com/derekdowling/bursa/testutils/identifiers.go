package testutils

import (
	"fmt"
	"strings"
	"time"
)

var test_id string

func init() {
	TestId()
}

func TestId() string {
	if test_id == "" {
		test_id = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return test_id
}

// Returns a nicely formatted test id to help identify/keep tests separate
func SuffixedId(suffix string) string {
	return strings.Join([]string{"test", TestId(), suffix}, "_")
}

func TestEmail(suffix string) string {
	return "admin+" + SuffixedId(suffix) + "@bursa.io"
}

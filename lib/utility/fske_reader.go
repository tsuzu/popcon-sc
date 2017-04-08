package utility

import "io"

type FakeEmptyReadCloser struct{}

func (r *FakeEmptyReadCloser) Read(b []byte) (n int, err error) {
	return 0, io.EOF
}
func (r *FakeEmptyReadCloser) Close() error {
	return nil
}

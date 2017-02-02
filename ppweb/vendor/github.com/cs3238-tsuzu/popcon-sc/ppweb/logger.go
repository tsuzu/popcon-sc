package main

import (
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/derekdowling/bursa/middleware/logtext"
)

var HttpLog, DBLog *logrus.Entry

type MulLog struct {
	console *os.File
	fp      *os.File
}

func (lo MulLog) Write(p []byte) (n int, err error) {
	lo.console.Write(p)

	if lo.fp != nil {
		lo.fp.Write(p)
	}

	return len(p), nil
}

func NewLogMultipleOutput(path string) (*MulLog, error) {
	var fp *os.File
	var err error

	if len(path) != 0 {
		fp, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

		if err != nil {
			return nil, err
		}
	}

	return &MulLog{
		console: os.NewFile(os.Stdout.Fd(), "stdout"),
		fp:      fp,
	}, nil
}

type CustomizedWriter struct {
	cb func([]byte) (int, error)
}

func (cw CustomizedWriter) Write(b []byte) (int, error) {
	return cw.cb(b)
}

func NewCustomizedWriter(cb func([]byte) (int, error)) CustomizedWriter {
	return CustomizedWriter{cb}
}

func CreateLogger(writer io.Writer) {
	logrus.SetOutput(writer)
	logrus.SetFormatter(logtext.NewLogtext(new(logrus.TextFormatter), true))

	HttpLog = logrus.WithField("category", "http")
	DBLog = logrus.WithField("category", "database")
}
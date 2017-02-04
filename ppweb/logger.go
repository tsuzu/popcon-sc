package main

import (
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/bursa/middleware/logtext"
)

var HttpLog, DBLog, MailLog *logrus.Entry

type LogMultipleOutput struct {
	console *os.File
	fp      *os.File
}

func (lo LogMultipleOutput) Write(p []byte) (n int, err error) {
	lo.console.Write(p)

	if lo.fp != nil {
		lo.fp.Write(p)
	}

	return len(p), nil
}

func NewLogMultipleOutput(path string) (*LogMultipleOutput, error) {
	var fp *os.File
	var err error

	if len(path) != 0 {
		fp, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

		if err != nil {
			return nil, err
		}
	}

	return &LogMultipleOutput{
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

	lt := &logtext.Logtext{
		Formatter:  new(logrus.TextFormatter),
		TargetMask: logtext.NewLogtextTargetMask(logrus.DebugLevel, logrus.ErrorLevel),
		LogDepth:   6,
	}
	logrus.SetFormatter(lt)

	HttpLog = logrus.WithField("category", "http")
	DBLog = logrus.WithField("category", "database")
	MailLog = logrus.WithField("category", "mail")
}

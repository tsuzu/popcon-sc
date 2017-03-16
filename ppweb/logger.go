package main

import (
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/bursa/middleware/logtext"
)

var HttpLog, DBLog, MailLog, FSLog *logrus.Entry

type CustomizedWriter struct {
	cb func([]byte) (int, error)
}

func (cw CustomizedWriter) Write(b []byte) (int, error) {
	return cw.cb(b)
}

func NewCustomizedWriter(cb func([]byte) (int, error)) CustomizedWriter {
	return CustomizedWriter{cb}
}

func InitLogger(writer io.Writer) {
	logrus.SetOutput(writer)

	lt := &logtext.Logtext{
		Formatter:  new(logrus.TextFormatter),
		TargetMask: logtext.NewLogtextTargetMask(logrus.DebugLevel, logrus.ErrorLevel, logrus.FatalLevel),
		LogDepth:   6,
	}
	logrus.SetFormatter(lt)

	HttpLog = logrus.WithField("category", "http")
	DBLog = logrus.WithField("category", "database")
	FSLog = logrus.WithField("category", "mongofs")
	MailLog = logrus.WithField("category", "mail")
}

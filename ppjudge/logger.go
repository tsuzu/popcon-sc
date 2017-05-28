package main

import (
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/bursa/middleware/logtext"
)

var HttpLog, JudgeLog, DBLog, MailLog, FSLog func() *logrus.Entry

type CustomizedWriter struct {
	cb func([]byte) (int, error)
}

func (cw CustomizedWriter) Write(b []byte) (int, error) {
	return cw.cb(b)
}

func NewCustomizedWriter(cb func([]byte) (int, error)) CustomizedWriter {
	return CustomizedWriter{cb}
}

func InitLogger(writer io.Writer, isDebug bool) {
	logrus.SetOutput(writer)
	if isDebug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	lt := &logtext.Logtext{
		Formatter:  new(logrus.TextFormatter),
		TargetMask: logtext.NewLogtextTargetMask(logrus.DebugLevel, logrus.ErrorLevel, logrus.FatalLevel),
		LogDepth:   6,
	}
	logrus.SetFormatter(lt)

	HttpLog = func() *logrus.Entry { return logrus.WithField("category", "http") }
	JudgeLog = func() *logrus.Entry { return logrus.WithField("category", "judge") }
	DBLog = func() *logrus.Entry { return logrus.WithField("category", "database") }
	FSLog = func() *logrus.Entry { return logrus.WithField("category", "mongofs") }
	MailLog = func() *logrus.Entry { return logrus.WithField("category", "mail") }
}

package main

import (
	"io"
	"net/http"
	"os"
	"syscall"

	"sync"

	"os/signal"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/bursa/middleware/logtext"
	"github.com/cs3238-tsuzu/popcon-sc/types"
)

func InitLogger(writer io.Writer) {
	logrus.SetOutput(writer)

	lt := &logtext.Logtext{
		Formatter:  new(logrus.TextFormatter),
		TargetMask: logtext.NewLogtextTargetMask(logrus.DebugLevel, logrus.ErrorLevel, logrus.FatalLevel),
		LogDepth:   6,
	}
	logrus.SetFormatter(lt)

}

func main() {
	InitLogger(os.NewFile(os.Stdout.Fd(), "stdout"))
	token := os.Getenv("PP_TOKEN")
	addr := os.Getenv("PP_LISTEN")
	// env: PP_MONGO

	mux := http.NewServeMux()

	wg := sync.WaitGroup{}
	finCh := make(chan bool)

	// remove_file.go
	InitRemoveFile(mux, finCh, &wg)

	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-signal_chan

		close(finCh)
	}()

	if len(addr) == 0 {
		addr = ":7502"
	}

	listener := func(rw http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get(sctypes.InternalHTTPToken)

		if auth != token {
			sctypes.ResponseTemplateWrite(http.StatusForbidden, rw)

			return
		}
	}

	if err := http.ListenAndServe(addr, http.HandlerFunc(listener)); err != nil {
		logrus.WithError(err).Error("ListenAndServe error")
	}
}

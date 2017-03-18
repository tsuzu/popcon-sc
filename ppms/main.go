package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"

	"sync"

	"os/signal"

	"context"

	"github.com/Sirupsen/logrus"
	"github.com/cs3238-tsuzu/bursa/middleware/logtext"
	"github.com/cs3238-tsuzu/popcon-sc/types"
)

func InitLogger(writer io.Writer) {
	logrus.SetOutput(writer)
	logrus.SetLevel(logrus.DebugLevel)

	lt := &logtext.Logtext{
		Formatter:  new(logrus.TextFormatter),
		TargetMask: logtext.NewLogtextTargetMask(logrus.DebugLevel, logrus.ErrorLevel, logrus.FatalLevel),
		LogDepth:   6,
	}
	logrus.SetFormatter(lt)

}

func main() {
	fmt.Println("started")
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

	if len(addr) == 0 {
		addr = ":7502"
	}

	listener := func(rw http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get(sctypes.InternalHTTPToken)

		if auth != token {
			sctypes.ResponseTemplateWrite(http.StatusForbidden, rw)

			return
		}
		mux.ServeHTTP(rw, req)
	}
	server := http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(listener),
	}

	go func() {
		logrus.Info("Starting to listen and serve.")
		if err := server.ListenAndServe(); err != nil {
			logrus.WithError(err).Error("ListenAndServe error")
		}
	}()

	<-signal_chan

	close(finCh)
	if err := server.Shutdown(context.Background()); err != nil {
		logrus.WithError(err).Error("Shutdown(graceful shutdown) error")
	}

	os.Exit(0)
}

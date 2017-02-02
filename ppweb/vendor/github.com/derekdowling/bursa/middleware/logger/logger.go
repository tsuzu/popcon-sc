package logger

// A request/response logger for our Middleware. Records the start of a request,
// and some basic information after it has been finished

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"net/http"
	"time"
)

// Logger is a middleware handler that logs the request as it goes in and the response as it goes out.
type Logger struct {
	// Logger inherits from log.Logger used to log messages with the Logger middleware
	*log.Logger
}

// NewLogger returns a new Logger instance
func NewLogger() *Logger {
	return &Logger{log.New()}
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	l.WithFields(log.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	}).Info("Starting request")

	next(rw, r)

	res := rw.(negroni.ResponseWriter)
	l.WithFields(log.Fields{
		"status":      res.Status(),
		"status_text": http.StatusText(res.Status()),
		"time":        time.Since(start),
	}).Info("Request completed")
}

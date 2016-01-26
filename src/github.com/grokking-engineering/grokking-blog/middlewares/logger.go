package middlewares

import (
	"net/http"
	"time"

	"github.com/grokking-engineering/grokking-blog/utils/logs"
)

var l = logs.New("recovery")

type Logger struct {
}

func NewLogger() func(http.Handler) http.Handler {
	logger := Logger{}
	return logger.factory
}

type responseWriter struct {
	http.ResponseWriter

	httpCode int
}

func (w *responseWriter) WriteHeader(status int) {
	w.httpCode = status
	w.ResponseWriter.WriteHeader(status)
}

func (this Logger) factory(next http.Handler) http.Handler {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			wrapper := &responseWriter{
				ResponseWriter: w,
			}

			start := time.Now()
			l.WithFields(logs.M{
				"method": r.Method,
				"url":    r.URL.Path,
			}).Info("Started request")

			next.ServeHTTP(wrapper, r)

			logEntry := l.WithFields(logs.M{
				"method": r.Method,
				"url":    r.URL.Path,
				"time":   time.Since(start).String(),
				"status": wrapper.httpCode,
			})
			if wrapper.httpCode < 400 {
				logEntry.Info("Completed request")
			} else {
				logEntry.Error("Completed request with error")
			}
		})
}

package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/grokking-engineering/grokking-blog/utils/logs"
)

type Recovery struct {
}

func NewRecovery() func(http.Handler) http.Handler {
	r := Recovery{}
	return r.factory
}

func (r Recovery) factory(next http.Handler) http.Handler {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					l.WithFields(logs.M{
						"error": err,
					}).Error("500 Internal Server Error")
					debug.PrintStack()
					w.Write([]byte("500 Server Error"))
				}
			}()

			next.ServeHTTP(w, r)
		})
}

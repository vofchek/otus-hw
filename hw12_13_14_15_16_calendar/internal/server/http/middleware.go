package internalhttp

import (
	"fmt"
	"net/http"
	"time"
)

// инжектим Logger, поменял сигнатуру, чтобы не тащить конфиг через контекст.
func loggingMiddleware(logg Logger, next http.Handler) http.Handler { //nolint:unused
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		decoratedWrapper := &ResponseWriter{
			w: w,
		}

		next.ServeHTTP(decoratedWrapper, r)

		logg.Info(
			fmt.Sprintf(
				"%s [%s] %s %s %s %d %d \"%s\"",
				r.RemoteAddr,                    // IP клиента
				time.Now().Format(time.RFC3339), // дата и время запроса
				r.Method,                        // метод
				r.RequestURI,                    // path
				r.Proto,                         // версия HTTP
				decoratedWrapper.StatusCode,     // код ответа
				time.Since(start),               // время обработки запроса
				r.UserAgent(),                   // user agent
			),
		)
	})
}

type ResponseWriter struct {
	w          http.ResponseWriter
	StatusCode int
}

func (w *ResponseWriter) Header() http.Header {
	return w.w.Header()
}
func (w *ResponseWriter) Write(data []byte) (int, error) {
	return w.w.Write(data)
}
func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.w.WriteHeader(statusCode)
}

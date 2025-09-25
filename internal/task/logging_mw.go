package task

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

func LoggingMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bodyCopy []byte
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			if r.Body != nil {
				data, err := io.ReadAll(r.Body)
				if err == nil {
					bodyCopy = data
					r.Body = io.NopCloser(bytes.NewBuffer(data)) // восстанавливаем тело для хэндлера
				}
			}
		}

		if len(bodyCopy) > 0 {
			log.Printf("%s %s %q", r.Method, r.URL.Path, string(bodyCopy))
		} else {
			log.Printf("%s %s", r.Method, r.URL.Path)
		}

		next.ServeHTTP(w, r)
	})
}

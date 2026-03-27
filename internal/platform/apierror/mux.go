package apierror

import (
	"bytes"
	"net/http"
)

func NormalizeServeMux(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler, pattern := mux.Handler(r)
		if pattern != "" {
			handler.ServeHTTP(w, r)
			return
		}

		recorder := &responseRecorder{}
		handler.ServeHTTP(recorder, r)

		switch recorder.status {
		case http.StatusNotFound:
			Write(w, r, NotFound("Route not found"))
		case http.StatusMethodNotAllowed:
			if allow := recorder.header.Get("Allow"); allow != "" {
				w.Header().Set("Allow", allow)
			}
			Write(w, r, MethodNotAllowed("Method not allowed"))
		default:
			recorder.FlushTo(w)
		}
	})
}

type responseRecorder struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func (r *responseRecorder) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}

	return r.header
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}

	r.status = status
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	return r.body.Write(data)
}

func (r *responseRecorder) FlushTo(w http.ResponseWriter) {
	for key, values := range r.Header() {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	if r.status != 0 {
		w.WriteHeader(r.status)
	}

	if r.body.Len() == 0 {
		return
	}

	_, _ = w.Write(r.body.Bytes())
}

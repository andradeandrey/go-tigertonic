package tigertonic

import (
	"fmt"
	"net/http"
)

// FirstHandler is an http.Handler that, for each handler in its slice of
// handlers, calls ServeHTTP until the first one that calls w.WriteHeader.
type FirstHandler []http.Handler

// First returns an http.Handler that, for each handler in its slice of
// handlers, calls ServeHTTP until the first one that calls w.WriteHeader.
func First(handlers ...http.Handler) FirstHandler {
	return handlers
}

// ServeHTTP calls each handler in its slice of handlers until the first one
// that calls w.WriteHeader.
func (fh FirstHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w0 := &firstResponseWriter{w, false}
	for _, h := range fh {
		h.ServeHTTP(w0, r)
		if w0.written {
			break
		}
	}
}

// If returns an http.Handler that conditionally calls another http.Handler
// unless the given function returns an error.  In that case, the error
// is used to create a plaintext or JSON response as dictated by the Accept
// header.
func If(f func(*http.Request) (http.Header, error), h http.Handler) FirstHandler {
	return First(ifHandler(f), h)
}

type firstResponseWriter struct {
	http.ResponseWriter
	written bool
}

func (w *firstResponseWriter) WriteHeader(status int) {
	w.written = true
	w.ResponseWriter.WriteHeader(status)
}

type ifHandler func(*http.Request) (http.Header, error)

func (ih ifHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	header, err := ih(r)
	for name, values := range header {
		for _, value := range values {
			w.Header().Set(name, value)
		}
	}
	if nil == err {
		return
	}
	description := err.Error()
	status := http.StatusInternalServerError
	if httpEquivError, ok := err.(HTTPEquivError); ok {
		status = httpEquivError.Status()
	}
	if acceptJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		writeJSONError(w, err)
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(status)
		fmt.Fprint(w, description)
	}
}

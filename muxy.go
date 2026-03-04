package muxy

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

type Mux struct {
	mux             *http.ServeMux
	tree            *node
	notFoundHandler http.HandlerFunc
	handler         http.Handler
}

func New() *Mux {
	s := &Mux{mux: http.NewServeMux()}
	notFound := HandlerFunc(func(w *Response, r *Request) {
		w.Error(404, errors.New("Route not found"))
	})
	s.notFoundHandler = notFound
	return s
}

func (s *Mux) MultiHandleFunc(patterns []string, handler func(w Response, r *Request)) {
	for _, pattern := range patterns {
		s.HandleFunc(pattern, handler)
	}
}

func (s *Mux) HandleFunc(pattern string, handler func(w Response, r *Request)) {
	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		handler(Response{ResponseWriter: w}, &Request{Request: r})
	})
}

func HandlerFunc(handler func(w *Response, r *Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(&Response{ResponseWriter: w}, &Request{Request: r})
	})
}

func (s *Mux) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Mux) ServeHTTP(w *Response, r *Request) {
	w.Start = time.Now()

	path := r.URL.Path

	segments := split(path)

	params := make(map[string]string)

	handler := s.tree.match(segments, params)

	if handler == nil {
		if s.notFoundHandler == nil {
			http.NotFound(w, r.Request)
			return
		}
		s.notFoundHandler.ServeHTTP(w, r.Request)
		return
	}

	handler.ServeHTTP(w, r.Request)
}

func split(path string) []string {
	if path == "/" {
		return nil
	}

	return strings.Split(strings.Trim(path, "/"), "/")
}

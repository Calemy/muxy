package muxy

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Mux struct {
	mux     *chi.Mux
	handler http.Handler
}

func New() *Mux {
	s := &Mux{mux: chi.NewMux()}
	notFound := HandlerFunc(func(w *Response, r *Request) {
		w.Error(404, errors.New("Route not found"))
	})
	s.mux.NotFound(notFound)
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
	s.mux.ServeHTTP(w, r.Request)
}

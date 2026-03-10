package muxy

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

type Mux struct {
	tree            *node
	prefix          string
	notFoundHandler HandlerFunc
}

func New() *Mux {
	s := &Mux{tree: &node{}}
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
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(Response{ResponseWriter: w}, &Request{Request: r})
	})

	s.Handle(pattern, h)
}

func (s *Mux) HandleMethod(method string, pattern string, handler func(w Response, r *Request)) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(Response{ResponseWriter: w}, &Request{Request: r})
	})

	s.handle(method, pattern, h)
}

func (s *Mux) Handle(pattern string, handler http.Handler) {
	s.handle("*", pattern, handler)
}

func (s *Mux) Get(pattern string, handler func(w Response, r *Request)) {
	s.HandleMethod("GET", pattern, handler)
}

func (s *Mux) Post(pattern string, handler func(w Response, r *Request)) {
	s.HandleMethod("POST", pattern, handler)
}

func (s *Mux) Put(pattern string, handler func(w Response, r *Request)) {
	s.HandleMethod("PUT", pattern, handler)
}

func (s *Mux) Patch(pattern string, handler func(w Response, r *Request)) {
	s.HandleMethod("PATCH", pattern, handler)
}

func (s *Mux) Delete(pattern string, handler func(w Response, r *Request)) {
	s.HandleMethod("DELETE", pattern, handler)
}

func (s *Mux) handle(method string, pattern string, handler http.Handler) {
	path := s.prefix + pattern

	if s.tree == nil {
		s.tree = &node{}
	}

	s.tree.insert(method, path, handler)
}

func (s *Mux) Group(prefix string) *Mux {
	return &Mux{
		tree:            s.tree,
		prefix:          s.prefix + clean(prefix),
		notFoundHandler: s.notFoundHandler,
	}
}

func (s *Mux) ServeHTTP(w *Response, r *Request) {
	w.Start = time.Now()

	path := r.URL.Path

	segments := split(path)

	params := make(map[string]string)

	handler := s.tree.match(r.Method, segments, params)

	if handler == nil {
		if s.notFoundHandler == nil {
			http.NotFound(w, r.Request)
			return
		}
		s.notFoundHandler.ServeHTTP(w, r.Request)
		return
	}

	for k, v := range params {
		r.Request.SetPathValue(k, v)
	}

	handler.ServeHTTP(w, r.Request)
}

type HandlerFunc func(w *Response, r *Request)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(&Response{ResponseWriter: w}, &Request{Request: r})
}

func split(path string) []string {
	if path == "/" {
		return nil
	}

	return strings.Split(strings.Trim(path, "/"), "/")
}

func clean(p string) string {
	if p == "" {
		return ""
	}

	if p[0] != '/' {
		p = "/" + p
	}

	return strings.TrimRight(p, "/")
}

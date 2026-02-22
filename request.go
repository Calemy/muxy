package muxy

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type Request struct {
	*http.Request
}

func (r *Request) check(sFunc func(string) string, keys ...string) string {
	if len(keys) == 0 {
		panic("muxy: need to provide at least one key to check")
	}

	for _, key := range keys {
		if s := sFunc(key); s != "" {
			return s
		}
	}

	if len(keys) < 2 { //* If there is no default value, return empty string
		return ""
	}

	return keys[len(keys)-1] //* Returns last key as default value
}

func (r *Request) checkArray(keys ...string) []string {
	if len(keys) == 0 {
		panic("muxy: need to provide at least one key to check")
	}

	queries := make(map[string][]string)
	query := r.URL.RawQuery

	for _, pair := range strings.Split(query, "&") {
		kv := strings.SplitN(pair, "=", 2) //* Splits the query pair into key and value
		if len(kv) != 2 {
			continue
		}

		if kv[1] == "" { //* If the value is empty, we ignore it and move on
			continue
		}

		if _, ok := queries[kv[0]]; !ok {
			queries[kv[0]] = make([]string, 0)
		}

		queries[kv[0]] = append(queries[kv[0]], kv[1])
	}

	for _, key := range keys {
		if values, ok := queries[key]; ok {
			return values
		}
	}
	return make([]string, 0)
}

func (r *Request) Queries(keys ...string) []string {
	value := r.checkArray(keys...)
	if len(value) == 0 { //* If the key is not found, return empty array
		return value
	} else if len(value) == 1 { //* If the key is found and has only one value, we check if it was tried to be seperated by comma.
		return strings.Split(value[0], ",")
	}
	return value
}

func (r *Request) Query(keys ...string) string {
	return r.check(r.Request.URL.Query().Get, keys...)
}

func (r *Request) Param(keys ...string) string {
	return r.check(r.PathValue, keys...)
}

func (r *Request) Auth(fallback ...string) string {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		return auth
	}

	if len(fallback) == 0 {
		return ""
	}

	return fallback[0]
}

func (r *Request) IP() string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return ip
}

func (r *Request) Context() context.Context {
	return r.Request.Context()
}

func (r *Request) WithContext(ctx context.Context) *Request {
	r.Request = r.Request.WithContext(ctx)
	return r
}

func (r *Request) Clone(ctx context.Context) *Request {
	r.Request = r.Request.Clone(ctx)
	return r
}

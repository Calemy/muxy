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

/*
Queries returns the values of the given keys from the request query.
If no key is not found, an empty array is returned.

The order of seperated values is ?query=value1&query=value2 > ?query=value1,value2
*/
func (r *Request) Queries(keys ...string) []string {
	value := r.checkArray(keys...)
	if len(value) == 0 { //* If the key is not found, return empty array
		return value
	} else if len(value) == 1 { //* If the key is found and has only one value, we check if it was tried to be seperated by comma.
		return strings.Split(value[0], ",")
	}
	return value
}

/*
Query returns the first value of the given keys from the request query.

If more than one key is given, the last value is returned as a fallback, otherwise an empty string is returned.
*/
func (r *Request) Query(keys ...string) string {
	return r.check(r.Request.URL.Query().Get, keys...)
}

// ? Should this really be an array? In case we could make a fallback argument | Only on minor version bump

/*
Param returns the first value of the given keys from the request path.

If more than one key is given, the last value is returned as a fallback, otherwise an empty string is returned.
*/
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

/*
IP returns the first value of the given keys from the request headers.

If the ip cannot be found in the headers, the remote address is used.

Following headers are checked by default: CF-Connecting-IP, X-Real-IP, X-Forwarded-For
*/
func (r *Request) IP(header ...string) string {
	header = append(header, "CF-Connecting-IP", "X-Real-IP", "X-Forwarded-For")

	ip := r.check(r.Header.Get, header...)
	if ip == header[len(header)-1] {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return ip
}

// Aquires the context from the request
func (r *Request) Context() context.Context {
	return r.Request.Context()
}

// Sets the context to the request
func (r *Request) WithContext(ctx context.Context) *Request {
	r.Request = r.Request.WithContext(ctx)
	return r
}

// Clones the request with the given context
func (r *Request) Clone(ctx context.Context) *Request {
	r.Request = r.Request.Clone(ctx)
	return r
}

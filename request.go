package muxy

import (
	"context"
	"maps"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
)

type Request struct {
	*http.Request
	ctx         context.Context
	matches     []string          // values for the matching wildcards in pat
	otherValues map[string]string // for calls to SetPathValue that don't match a wildcard
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

	if len(keys) < 2 { // * If there is no default value, return empty string
		return ""
	}

	return keys[len(keys)-1] // * Returns last key as default value
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
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("muxy: nil context")
	}
	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx
	return r2
}

func (r *Request) Clone(ctx context.Context) *Request {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx
	r2.URL = cloneURL(r.URL)
	r2.Header = r.Header.Clone()
	r2.Trailer = r.Trailer.Clone()
	if s := r.TransferEncoding; s != nil {
		s2 := make([]string, len(s))
		copy(s2, s)
		r2.TransferEncoding = s2
	}
	r2.Form = cloneURLValues(r.Form)
	r2.PostForm = cloneURLValues(r.PostForm)
	r2.MultipartForm = cloneMultipartForm(r.MultipartForm)

	// Copy matches and otherValues. See issue 61410.
	if s := r.matches; s != nil {
		s2 := make([]string, len(s))
		copy(s2, s)
		r2.matches = s2
	}
	r2.otherValues = maps.Clone(r.otherValues)
	return r2
}

func cloneURLValues(v url.Values) url.Values {
	if v == nil {
		return nil
	}
	return url.Values(http.Header(v).Clone())
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	u2 := new(url.URL)
	*u2 = *u
	if u.User != nil {
		u2.User = new(url.Userinfo)
		*u2.User = *u.User
	}
	return u2
}

func cloneMultipartForm(f *multipart.Form) *multipart.Form {
	if f == nil {
		return nil
	}
	f2 := &multipart.Form{
		Value: (map[string][]string)(http.Header(f.Value).Clone()),
	}
	if f.File != nil {
		m := make(map[string][]*multipart.FileHeader, len(f.File))
		for k, vv := range f.File {
			vv2 := make([]*multipart.FileHeader, len(vv))
			for i, v := range vv {
				vv2[i] = cloneMultipartFileHeader(v)
			}
			m[k] = vv2
		}
		f2.File = m
	}
	return f2
}

func cloneMultipartFileHeader(fh *multipart.FileHeader) *multipart.FileHeader {
	if fh == nil {
		return nil
	}
	fh2 := new(multipart.FileHeader)
	*fh2 = *fh
	fh2.Header = textproto.MIMEHeader(http.Header(fh.Header).Clone())
	return fh2
}

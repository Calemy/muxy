# muxy
a mux router using chi as base so i don't have to figure out patterns :)

*A simple mux wrapper, without the common annoyances*

Changes are expected.

This is based on [chi](github.com/go-chi/chi) because wildcard patterns are hard :(

---

##  How to use it

### Download the module

```bash
go get -u github.com/calemy/muxy
```

### Serve a new mux

You can just call a new mux router with muxy, it implements the http.Handler interface properly.

```go
import (
	"fmt"
	"log"
	"net/http"

	"github.com/calemy/muxy"
)

func main() {
	mux := muxy.New()
	handler := muxy.HandlerFunc(func(w *muxy.Response, r *muxy.Request) {
		mux.ServeHTTP(w, r)
	})

	http.ListenAndServe(":2445", handler)
}
```
---

## Timings

muxy has important timings built-in that can be called via the response. The output can be chosen via the muxy.Time format

```go
handler := muxy.HandlerFunc(func(w *muxy.Response, r *muxy.Request) {
    mux.ServeHTTP(w, r)
    log.Printf("%.2fms, %.2fms", w.Latency(muxy.Millisecond), w.Duration(muxy.Millisecond))
})
```

---

## Query & Param

You can get a query by calling Query() inside request, while receiving a param via Param().
For queries, you can attach multiple keys and use a default value when no key matches.
Keys should be set by priority, the first key matched returns.

```go
mux.HandleFunc("/hello/{user}", func(w muxy.Response, r *muxy.Request) {
    q := r.Query("query", "q", "Newest") // searches for query first, q if query is empty, and returns "Newest" as default value.
    limit := r.Query("limit") // searches for limit only, returns "" as default value 
    user := r.Param("user") // searches for user only, returns "" as default value 
})
```

You can also receive an array of queries by calling Queries().
Yet again you can attach multiple keys, but no default value. The router accepts ?a=1&a=2 and falls back to ?a=1,2 in case the first one is not present.

```go
mux.HandleFunc("/hello", func(w muxy.Response, r *muxy.Request) {
    status := r.Queries("status", "ranked") // checks for status first, is status is empty, it checks ranked.
})
```

---

## Errors

Return an error easily using Error(). We currently format errors as json.

```go
mux.HandleFunc("/error", func(w muxy.Response, r *muxy.Request) {
    w.Error(500, errors.New("Something went wrong with this request"))
})
```

---

## Shortcuts

We added shortcuts to certain request headers to make your life a little easier. Those include but are not limited to:
- Auth (Authentication)
- IP (X-Real-IP, RemoteAddress)

Note: We are aware of cloudflare, and forwarded headers. Those will be added at a later time.

---

## Example

Here's a very simple http server with a logger.

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/calemy/muxy"
)

func main() {
	mux := muxy.New()
	handler := muxy.HandlerFunc(func(w *muxy.Response, r *muxy.Request) {
		mux.ServeHTTP(w, r)
		log.Printf("%s -> %s | Latency: %.2fms, Duration: %.2fms", r.IP(), r.Path, w.Latency(muxy.Millisecond), w.Duration(muxy.Millisecond))
	})

	mux.HandleFunc("/hello", func(w muxy.Response, r *muxy.Request) {
		w.WriteHeader(200)
		w.Write([]byte("Hello, World!"))
	})

	mux.HandleFunc("/error", func(w muxy.Response, r *muxy.Request) {
      w.Error(500, errors.New("Something went wrong with this request"))
	})

	http.ListenAndServe(":2445", handler)
}

```

---

## Roadmap

We have a list of things we want to shortcut or improve including:
- response.Stream
- response.Reply (string)
- improve request.IP Headers
- grouping

If you find this package useful, feel free to leave a star on the repository!


Contributions are always welcomed via Issues and PRs.

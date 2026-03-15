package muxy

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	http.ResponseWriter
	Status    int
	Start     time.Time
	firstByte time.Time
	end       time.Time
}

type Time float64

type APIError struct {
	Error string `json:"error"`
}

const (
	_           Time = 0 // TODO: Auto
	Nanosecond       = 1
	Microsecond      = 1000 * Nanosecond
	Millisecond      = 1000 * Microsecond
	Second           = 1000 * Millisecond
)

/*
WriteHeader writes the header to the response writer and sets the status code. It also marks the time of the first byte.
*/
func (w *Response) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.firstByte = time.Now()
	w.Status = code
}

/*
Write writes the given bytes to the response writer. It also marks the time of the last byte written to the client.
*/
func (w *Response) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.end = time.Now()
	return n, err
}

/*
Receives the time between the start of the request and the first byte was written to the client.
*/
func (w *Response) Latency(unit Time) float64 {
	return timeSince(w.firstByte, w.Start, unit)
}

/*
Receives the duration between the first byte and last byte that was written to the client.
*/
func (w *Response) Duration(unit Time) float64 {
	return timeSince(w.end, w.firstByte, unit)
}

func timeSince(t time.Time, start time.Time, unit Time) float64 {
	return float64(t.Sub(start).Nanoseconds()) / float64(unit)
}

/*
Error sends an error to the client with the given code and message.

The error message is sent as a json object.

This was inspired by JavaScript Frameworks (fastify, elysia, express)
*/
func (w *Response) Error(code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	bytes, _ := json.Marshal(APIError{err.Error()})
	w.Write(bytes)
}

/*
Sets the deprecation header and the sunset header if the sunset argument is given.
*/
func (w *Response) Deprecate(deprecation time.Time, sunset ...time.Time) {
	w.Header().Set("Deprecated", deprecation.Format(http.TimeFormat))
	if len(sunset) == 0 {
		return
	}
	w.Header().Set("Sunset", sunset[0].Format(http.TimeFormat))
}

/*
Streams the given reader to the client to avoid buffering the whole response with the given content type.

If no content length is given, the transfer encoding is set to chunked to avoid malformed responses.
*/
func (w *Response) Stream(r io.Reader, contentType string, size ...int) {
	w.Header().Set("Content-Type", contentType)
	if len(size) > 0 {
		w.Header().Set("Content-Length", strconv.Itoa(size[0]))
	} else {
		w.Header().Set("Transfer-Encoding", "chunked")
	}
	io.Copy(w, r)
}

/*
Sends a string as plain text response to the client.
*/
func (w *Response) Reply(code int, message string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(message))
}

/*
Sends a json object as response to the client.
*/
func (w *Response) JSON(code int, v interface{}) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	bytes, _ := json.Marshal(v)
	w.Write(bytes)
}

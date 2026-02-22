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

func (w *Response) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.firstByte = time.Now()
	w.Status = code
}

func (w *Response) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.end = time.Now()
	return n, err
}

func (w *Response) Latency(unit Time) float64 {
	return timeSince(w.firstByte, w.Start, unit)
}

func (w *Response) Duration(unit Time) float64 {
	return timeSince(w.end, w.firstByte, unit)
}

func timeSince(t time.Time, start time.Time, unit Time) float64 {
	return float64(t.Sub(start).Nanoseconds()) / float64(unit)
}

func (w *Response) Error(code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	bytes, _ := json.Marshal(APIError{err.Error()})
	w.Write(bytes)
}

func (w *Response) Deprecate(deprecation time.Time, sunset ...time.Time) {
	w.Header().Set("Deprecated", deprecation.Format(http.TimeFormat))
	if len(sunset) == 0 {
		return
	}
	w.Header().Set("Sunset", sunset[0].Format(http.TimeFormat))
}

func (w *Response) Stream(r io.Reader, contentType string, size ...int) {
	w.Header().Set("Content-Type", contentType)
	if len(size) > 0 {
		w.Header().Set("Content-Length", strconv.Itoa(size[0]))
	}
	io.Copy(w, r)
}

func (w *Response) Reply(message string) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(message))
}

func (w *Response) JSON(v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	bytes, _ := json.Marshal(v)
	w.Write(bytes)
}

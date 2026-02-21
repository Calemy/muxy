package muxy

import (
	"encoding/json"
	"net/http"
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

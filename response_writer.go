package gem

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

// The responseWriter is about to be used as a buffer for writes.
// Writes the value stored in it to the real Writer at the end of the response.
type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher

	// Status returns the HTTP response status code of the current request.
	Status() int

	// Size returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// WriteString writes the string into the response body.
	WriteString(string) (int, error)

	// Written returns true if the response body was already written.
	Written() bool

	// WriteHeaderNow forces to write the http header (status code + headers).
	WriteHeaderNow()

	// Pusher get the http.Pusher for server push
	Pusher() http.Pusher
}

func (r *responseWriter) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func (r *responseWriter) reset(writer http.ResponseWriter) {
	r.ResponseWriter = writer
	r.size = noWritten
	r.status = defaultStatus
}

func (r *responseWriter) Write(data []byte) (int, error) {
	r.WriteHeaderNow()
	n, err := r.ResponseWriter.Write(data)
	r.size += n

	return n, err
}

func (r *responseWriter) WriteHeader(code int) {
	if code > 0 && r.status != code {
		if r.Written() {
			return
		}
		r.status = code
	}
}

func (r *responseWriter) WriteString(s string) (int, error) {
	r.WriteHeaderNow()
	n, err := io.WriteString(r.ResponseWriter, s)
	r.size += n

	return n, err
}

func (r *responseWriter) Written() bool {
	return r.size != noWritten
}

func (r *responseWriter) WriteHeaderNow() {
	if !r.Written() {
		r.size = 0
		r.ResponseWriter.WriteHeader(r.status)
	}
}

func (r *responseWriter) Pusher() http.Pusher {
	if pusher, ok := r.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}

func (r *responseWriter) Status() int {
	return r.status
}

func (r *responseWriter) Size() int {
	return r.size
}

// Hijack implements the http.Hijacker interface.
// Supports scenarios such as WebSocket or long connections
// Custom network communication outside the HTTP protocol,
// converting HTTP connections to WebSocket and HTTP long connections.
func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if r.size < 0 {
		r.size = 0
	}
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush implements the http.Flusher interface.
// The ability to manually refresh the response that has been written to the client.
// Typically used in scenarios with long connections or where real-time pushes are required,
// such as Server-Sent Events (SSE) or long polling.
func (r *responseWriter) Flush() {
	r.WriteHeaderNow()
	r.ResponseWriter.(http.Flusher).Flush()
}

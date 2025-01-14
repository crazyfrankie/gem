package gem

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/crazyfrankie/gem/internal/traceconv"
)

const (
	// traceKey is used to set the tracer's key in context.
	traceKey = "otel-contrib-go-tracer"

	// scopeName is the traceconv middleware scope name
	scopeName = "go.opentelemetry.io/contrib/instrumentation/github.com/crazyfrankie/gem/otelgem"
)

type TraceBuilder struct {
	TraceProvider trace.TracerProvider
	Propagations  propagation.TextMapPropagator
	httpconv      *traceconv.HttpConv
	SpanName      string
}

// HTTPRequset
// http.method             string
// http.scheme             string
// net.host.name           string
// net.host.port           int
// net.sock.peer.addr      string
// net.sock.peer.port      int
// user_agent.original     string
// http.client_ip          string
// net.protocol.name       string Note: not set if the value is "http".
// net.protocol.version    string
// http.target             string Note: doesn't include the query parameter.
func (t TraceBuilder) HTTPRequset(server string, r *http.Request) []attribute.KeyValue {
	// http Method, scheme, and host name.
	count := 3

	var host string
	var port int
	if server == "" {
		host, port = t.httpconv.SplitHostPort(r.Host)
	} else {
		host = server
		_, port = t.httpconv.SplitHostPort(r.Host)
	}
	count++

	protos := t.httpconv.NetConv.NetProtocol(r.Proto)
	if protos[0] != "" && protos[1] != "" {
		count++
	}

	attrs := make([]attribute.KeyValue, 0, count)

	attrs = append(attrs, t.httpconv.HTTPMethod(r.Method))
	attrs = append(attrs, t.httpconv.HTTPScheme(r.URL.Scheme == "http"))
	attrs = append(attrs, t.httpconv.NetConv.NetHostName(host))

	if port > 0 {
		attrs = append(attrs, attribute.Int("net.host.port", port))
	}

	attrs = append(attrs, attribute.String("net.protocol.version", protos[1]))

	return attrs
}

func (t TraceBuilder) SetStatusCode(status int) (codes.Code, string) {
	if status < 100 || status >= 600 {
		return codes.Error, fmt.Sprintf("Invalid http status code:%d", status)
	}
	if status >= 500 {
		return codes.Error, ""
	}

	return codes.Unset, ""
}

func (t TraceBuilder) Trace(service string) HandlerFunc {
	if t.httpconv == nil {
		t.httpconv = traceconv.NewHttpConv()
	}
	if t.TraceProvider == nil {
		t.TraceProvider = otel.GetTracerProvider()
	}
	tracer := t.TraceProvider.Tracer(
		scopeName,
		trace.WithInstrumentationVersion("v0.58.0"),
	)
	if t.Propagations == nil {
		t.Propagations = otel.GetTextMapPropagator()
	}
	return func(c *Context) {
		c.Set(traceKey, t.TraceProvider)

		// Get current context
		oldCtx := c.Request.Context()
		// defer does not prevent the context from being restored even if subsequent code throws an error.
		// And it ensures that the context of each request is separate
		defer func() {
			c.Request = c.Request.WithContext(oldCtx)
		}()

		ctx := t.Propagations.Extract(oldCtx, propagation.HeaderCarrier(c.Request.Header))
		opts := []trace.SpanStartOption{
			trace.WithAttributes(t.HTTPRequset(service, c.Request)...),
			trace.WithAttributes(semconv.HTTPRoute(c.FullPath())),
			trace.WithSpanKind(trace.SpanKindServer),
		}

		var spanName string
		if t.SpanName != "" {
			spanName = t.SpanName
		} else {
			spanName = c.FullPath()
		}

		ctx, span := tracer.Start(c, spanName, opts...)
		defer span.End()

		// Set the updated ctx back to the request context to ensure that the traceconv information
		// is accessible for subsequent processing
		c.Request = c.Request.WithContext(ctx)

		// Continued implementation of subsequent middleware
		c.Next()

		status := c.Writer.Status()
		span.SetStatus(t.SetStatusCode(status))
		if status > 0 {
			span.SetAttributes(attribute.Int("http.status_code", status))
		}
	}
}

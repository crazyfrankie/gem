package gem

import (
	"strings"
	"time"

	"github.com/crazyfrankie/gem/config"
)

// WithKeepAliveTimeout sets keep-alive timeout.
//
// In most cases, there is no need to care about this option.
func WithKeepAliveTimeout(t time.Duration) config.Option {
	return config.Option{F: func(o *config.Options) {
		o.KeepAliveTimeout = t
	}}
}

// WithReadTimeout sets read timeout.
//
// Close the connection when read request timeout.
func WithReadTimeout(t time.Duration) config.Option {
	return config.Option{F: func(o *config.Options) {
		o.ReadTimeout = t
	}}
}

// WithWriteTimeout sets write timeout.
//
// Connection will be closed when write request timeout.
func WithWriteTimeout(t time.Duration) config.Option {
	return config.Option{F: func(o *config.Options) {
		o.WriteTimeout = t
	}}
}

// WithExitWaitTime sets timeout for graceful shutdown.
//
// The server may exit ahead after all connections closed.
// All responses after shutdown will be added 'Connection: close' header.
func WithExitWaitTime(timeout time.Duration) config.Option {
	return config.Option{F: func(o *config.Options) {
		o.ExitWaitTimeout = timeout
	}}
}

// WithUseRawPath sets useRawPath.
//
// If enabled, the url.RawPath will be used to find parameters.
func WithUseRawPath(b bool) config.Option {
	return config.Option{
		F: func(o *config.Options) {
			o.UseRawPath = b
		},
	}
}

// WithHostPort sets listening address.
func WithHostPort(addr string) config.Option {
	return config.Option{
		F: func(o *config.Options) {
			o.Addr = addr
		},
	}
}

// WithRedirectFixedPath sets redirectFixedPath.
//
// If enabled, the router tries to fix the current request path, if no
// handle is registered for it.
// First superfluous path elements like ../ or // are removed.
// Afterwards the router does a case-insensitive lookup of the cleaned path.
// If a handle can be found for this route, the router makes a redirection
// to the corrected path with status code 301 for GET requests and 308 for
// all other request methods.
// For example /FOO and /..//Foo could be redirected to /foo.
// RedirectTrailingSlash is independent of this option.
func WithRedirectFixedPath(b bool) config.Option {
	return config.Option{F: func(o *config.Options) {
		o.RedirectFixedPath = b
	}}
}

// WithRedirectTrailingSlash sets redirectTrailingSlash.
//
// Enables automatic redirection if the current route can't be matched but a
// handler for the path with (without) the trailing slash exists.
// For example if /foo/ is requested but a route only exists for /foo, the
// client is redirected to /foo with http status code 301 for GET requests
// and 307 for all other request methods.
func WithRedirectTrailingSlash(b bool) config.Option {
	return config.Option{F: func(o *config.Options) {
		o.RedirectTrailingSlash = b
	}}
}

// WithRemoveExtraSlash sets removeExtraSlash.
//
// RemoveExtraSlash a parameter can be parsed from the URL even with extra slashes.
// If UseRawPath is false (by default), the RemoveExtraSlash effectively is true,
// as url.Path gonna be used, which is already cleaned.
func WithRemoveExtraSlash(b bool) config.Option {
	return config.Option{F: func(o *config.Options) {
		o.RemoveExtraSlash = b
	}}
}

// WithUnescapePathValues sets unescapePathValues.
//
// If true, the path value will be unescaped.
// If UseRawPath is false (by default), the UnescapePathValues effectively is true,
// as url.Path gonna be used, which is already unescaped.
func WithUnescapePathValues(b bool) config.Option {
	return config.Option{F: func(o *config.Options) {
		o.UnescapePathValues = b
	}}
}

// WithBasePath sets basePath.Must be "/" prefix and suffix,If not the default concatenate "/"
func WithBasePath(basePath string) config.Option {
	return config.Option{F: func(o *config.Options) {
		// Must be "/" prefix and suffix,If not the default concatenate "/"
		if !strings.HasPrefix(basePath, "/") {
			basePath = "/" + basePath
		}
		if !strings.HasSuffix(basePath, "/") {
			basePath = basePath + "/"
		}
		o.BasePath = basePath
	}}
}

// WithNetwork sets network. Support "tcp", "udp", "unix"(unix domain socket).
func WithNetwork(nw string) config.Option {
	return config.Option{F: func(o *config.Options) {
		o.Network = nw
	}}
}

// WithH2C sets whether enable H2C.
func WithH2C(enable bool) config.Option {
	return config.Option{F: func(o *config.Options) {
		o.H2C = enable
	}}
}

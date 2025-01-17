package config

import (
	"crypto/tls"
	"time"
)

// Option is the only struct that can be used to set Options.
type Option struct {
	F func(o *Options)
}

const (
	defaultKeepAliveTimeOut = 1 * time.Minute
	defaultReadTimeOut      = 3 * time.Minute
	defaultWaitExitTimeOut  = 5 * time.Second
	defaultAddr             = ":9090"
	defaultNetwork          = "tcp"
	defaultBasePath         = "/"
)

type Options struct {
	KeepAliveTimeout      time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	ExitWaitTimeout       time.Duration
	RedirectTrailingSlash bool
	RedirectFixedPath     bool
	RemoveExtraSlash      bool
	UnescapePathValues    bool
	UseRawPath            bool
	H2C                   bool
	Network               string
	Addr                  string
	BasePath              string
	TLS                   *tls.Config
}

func (o *Options) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

func NewOptions(opts []Option) *Options {
	options := &Options{
		KeepAliveTimeout:      defaultKeepAliveTimeOut,
		ReadTimeout:           defaultReadTimeOut,
		ExitWaitTimeout:       defaultWaitExitTimeOut,
		RedirectTrailingSlash: true,
		RedirectFixedPath:     false,
		RemoveExtraSlash:      false,
		UseRawPath:            false,
		UnescapePathValues:    true,
		H2C:                   false,
		BasePath:              defaultBasePath,
		Addr:                  defaultAddr,
		Network:               defaultNetwork,
		TLS:                   nil,
	}
	options.Apply(opts)
	return options
}

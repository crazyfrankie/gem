package binding

import "net/http"

type Binding interface {
	Name() string
	Bind(*http.Request, any) error
}

type BindingBody interface {
	Binding
	BindBody([]byte, any) error
}

type BindingUri interface {
	Name() string
	BindingUri(map[string][]string, any) error
}

var (
	JSON   BindingBody = jsonBinding{}
	PLAIN  BindingBody = plainBinding{}
	YAML   BindingBody = yamlBinding{}
	XML    BindingBody = xmlBinding{}
	Query  Binding     = queryBinding{}
	Header Binding     = headerBinding{}
	Uri    BindingUri  = uriBinding{}
)

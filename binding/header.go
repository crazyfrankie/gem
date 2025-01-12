package binding

import (
	"net/http"
)

type headerBinding struct {
}

func (h headerBinding) Name() string {
	return "header"
}

func (h headerBinding) Bind(request *http.Request, a any) error {
	panic("")
}

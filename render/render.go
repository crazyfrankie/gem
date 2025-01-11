package render

import "net/http"

type Render interface {
	Render(http.ResponseWriter) error

	WriteContentType(w http.ResponseWriter)
}

var (
	_ Render = (*JSON)(nil)
	_ Render = (*Data)(nil)
	_ Render = (*String)(nil)
	_ Render = (*ProtoBuf)(nil)
	_ Render = (*Redirect)(nil)
	_ Render = (*YAML)(nil)
	_ Render = (*XML)(nil)
)

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

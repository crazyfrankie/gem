package binding

import (
	"bytes"
	"encoding/xml"
	"net/http"
)

type xmlBinding struct {
}

func (x xmlBinding) Name() string {
	return "xml"
}

func (x xmlBinding) Bind(request *http.Request, obj any) error {
	decoder := xml.NewDecoder(request.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}

	return nil
}

func (x xmlBinding) BindBody(b []byte, obj any) error {
	decoder := xml.NewDecoder(bytes.NewReader(b))
	if err := decoder.Decode(obj); err != nil {
		return err
	}

	return nil
}

package binding

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type jsonBinding struct {
}

func (j jsonBinding) Name() string {
	return "json"
}

func (j jsonBinding) Bind(request *http.Request, obj any) error {
	decoder := json.NewDecoder(request.Body)

	if err := decoder.Decode(obj); err != nil {
		return err
	}

	return nil
}

func (j jsonBinding) BindBody(body []byte, obj any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))

	if err := decoder.Decode(obj); err != nil {
		return err
	}

	return nil
}

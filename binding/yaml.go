package binding

import (
	"bytes"
	"gopkg.in/yaml.v3"
	"net/http"
)

type yamlBinding struct {
}

func (y yamlBinding) Name() string {
	return "yaml"
}

func (y yamlBinding) Bind(request *http.Request, obj any) error {
	decoder := yaml.NewDecoder(request.Body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}

	return nil
}

func (y yamlBinding) BindBody(b []byte, obj any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(b))
	if err := decoder.Decode(obj); err != nil {
		return err
	}

	return nil
}

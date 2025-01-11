package render

import (
	"gopkg.in/yaml.v3"
	"net/http"
)

// YAML contains the given interface object.
type YAML struct {
	Data any
}

var yamlContentType = []string{"application/yaml; charset=utf-8"}

// Render (YAML) marshals the given interface object and writes data with custom ContentType.
func (y YAML) Render(writer http.ResponseWriter) error {
	y.WriteContentType(writer)

	bytes, err := yaml.Marshal(y.Data)
	if err != nil {
		return err
	}

	_, err = writer.Write(bytes)
	return err
}

// WriteContentType (YAML) writes YAML ContentType for response.
func (y YAML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, yamlContentType)
}

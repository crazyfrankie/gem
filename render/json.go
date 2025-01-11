package render

import (
	"encoding/json"
	"net/http"
)

// JSON　contains the given interface object.
type JSON struct {
	Data any
}

var jsonContentType = []string{"application/json; charset=utf-8"}

// Render (JSON) writes data with custom ContentType.
func (j JSON) Render(writer http.ResponseWriter) error {
	j.WriteContentType(writer)

	bytes, err := json.Marshal(j.Data)
	if err != nil {
		return err
	}

	_, err = writer.Write(bytes)
	return err
}

// WriteContentType (JSON) writes custom ContentType.
func (j JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

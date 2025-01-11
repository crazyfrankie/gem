package render

import (
	"net/http"
)

// Data contains ContentType and bytes data.
type Data struct {
	ContentType string
	Data        []byte
}

// Render (Data) writes data with custom ContentType.
func (d Data) Render(writer http.ResponseWriter) error {
	d.WriteContentType(writer)

	_, err := writer.Write(d.Data)
	return err
}

// WriteContentType (Data) writes custom ContentType.
func (d Data) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, []string{d.ContentType})
}

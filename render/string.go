package render

import (
	"fmt"
	"net/http"

	"github.com/crazyfrankie/gem/internal/bytestrconv"
)

// String contains the given interface object slice and its format.
type String struct {
	Format string
	Data   []any
}

var plainContentType = []string{"text/plain; charset=utf-8"}

// Render (String) writes data with custom ContentType.
func (s String) Render(writer http.ResponseWriter) error {
	s.WriteContentType(writer)

	if len(s.Data) > 0 {
		_, err := fmt.Fprintf(writer, s.Format, s.Data...)
		return err
	}

	_, err := writer.Write(bytestrconv.StringToBytes(s.Format))
	return err
}

// WriteContentType (String) writes Plain ContentType.
func (s String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, plainContentType)
}

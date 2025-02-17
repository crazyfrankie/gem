package render

import (
	"fmt"
	"net/http"
)

// Redirect contains the http request reference and redirects status code and location.
type Redirect struct {
	Code      int
	Request   *http.Request
	Localtion string
}

// Render (Redirect) redirects the http request to new location and writes redirect response.
func (r Redirect) Render(writer http.ResponseWriter) error {
	if (r.Code < http.StatusMultipleChoices || r.Code > http.StatusPermanentRedirect) && r.Code != http.StatusCreated {
		panic(fmt.Sprintf("Cannot redirect with status code %d", r.Code))
	}
	http.Redirect(writer, r.Request, r.Localtion, r.Code)
	return nil
}

// WriteContentType (Redirect) don't write any ContentType.
func (r Redirect) WriteContentType(w http.ResponseWriter) {}

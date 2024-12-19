package gem

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]any

type Context struct {
	// Origin Objects
	Request *http.Request
	Writer  http.ResponseWriter

	//RequestInfo
	Path   string
	Method string
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request: r,
		Writer:  w,
		Path:    r.URL.Path,
		Method:  r.Method,
	}
}

/*************************/
/***** REQUEST INFO ******/
/*************************/

// PostForm gets the value of the key in the form
func (c *Context) PostForm(key string) string {
	return c.Request.PostFormValue(key)
}

// Query gets the dynamic routing parameter
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// func (c *Context) Params(key string) string {
//
// }

/*************************/
/***** RESPONSE INFO ******/
/*************************/

// Status sets the HTTP response code.
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// SetHeader sets key in http header
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// String render a string to client
func (c *Context) String(code int, format string, values ...any) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON render a json to client
func (c *Context) JSON(code int, obj any) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	err := encoder.Encode(obj)
	if err != nil {
		panic(err)
	}
}

// Data render a byte stream to client
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// HTML render a html to client
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

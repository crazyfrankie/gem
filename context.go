package gem

import (
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/crazyfrankie/gem/binding"
	"github.com/crazyfrankie/gem/render"
)

// ContextKey is the key that a Context returns itself for.
const ContextKey = "_gem/contextkey"

type ContextKeyType int

const ContextRequestKey ContextKeyType = 0

const abortIndex int8 = math.MaxInt8 >> 1

type Context struct {
	// Origin Objects
	writemem responseWriter
	Request  *http.Request
	Writer   ResponseWriter

	index    int8
	fullPath string
	Params   Params
	handlers HandlersChain
	// This mutex protects Keys map.
	mu sync.RWMutex

	server       *Server
	params       *Params
	skippedNodes *[]skippedNode

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]any

	// queryCache caches the query result from c.Request.URL.Query().
	queryCache url.Values
}

/***************************************/
/*********** CONTEXT CREATION **********/
/***************************************/

// reset is called before the start of
// each HTTP request to ensure that the
// currently requested context is clean
func (c *Context) reset() {
	c.Writer = &c.writemem
	c.Params = c.Params[:0]
	c.handlers = nil
	c.index = -1

	c.fullPath = ""
	c.Keys = nil
	c.queryCache = nil
	*c.params = (*c.params)[:0]
	*c.skippedNodes = (*c.skippedNodes)[:0]
}

// Copy returns a copy of the current context that can be safely used outside the request's scope.
// This has to be used when the context has to be passed to a goroutine.
func (c *Context) copy() *Context {
	ctx := &Context{
		writemem: c.writemem,
		Writer:   c.Writer,
		server:   c.server,
	}

	ctx.writemem.ResponseWriter = nil
	ctx.Writer = &c.writemem
	ctx.index = abortIndex
	ctx.handlers = nil
	ctx.fullPath = c.fullPath

	cKeys := c.Keys
	ctx.Keys = make(map[string]any, len(cKeys))
	c.mu.RLock()
	for k, v := range c.Keys {
		ctx.Keys[k] = v
	}
	c.mu.RUnlock()

	return ctx
}

// HandlerNames returns a list of all registered handlers for this context in descending order,
// following the semantics of HandlerName()
func (c *Context) HandlerNames() []string {
	hn := make([]string, 0, len(c.handlers))
	for _, val := range c.handlers {
		if val == nil {
			continue
		}
		hn = append(hn, nameOfFunction(val))
	}
	return hn
}

/***************************************/
/********** METADATA MANAGEMENT ********/
/***************************************/

// Set is used to store a new key/value pair exclusively for this context.
func (c *Context) Set(key string, val any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = val
}

// Get returns the value for the given key, if exists: return(val, true)
// If not exists: return(nil, false)
func (c *Context) Get(key string) (value any, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.Keys[key]
	return
}

// MustGet returns the value for the given key if exists, otherwise panic
func (c *Context) MustGet(key string) any {
	if value, exists := c.Keys[key]; exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// FullPath returns a matched route full path.
// For not found routes returns an empty string.
func (c *Context) FullPath() string {
	return c.fullPath
}

/************************************/
/*********** FLOW CONTROL  **********/
/************************************/

// Next should be used only inside middleware.
// It executes the pending handlers in the chain inside the calling handler.
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		if c.handlers[c.index] != nil {
			c.handlers[c.index](c)
		}
		c.index++
	}
}

// Abort prevents pending handlers from being called. Note that this will not stop the current handler.
// Let's say you have an authorization middleware that validates that the current request is authorized.
// If the authorization fails (ex: the password does not match), call Abort to ensure the remaining handlers
// for this request are not called.
func (c *Context) Abort() {
	c.index = abortIndex
}

// IsAborted returns true if the current context was aborted.
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

// AbortWithStatus calls `Abort()` and writes the headers with the specified status code.
// For example, a failed attempt to authenticate a request could use: context.AbortWithStatus(401).
func (c *Context) AbortWithStatus(status int) {
	c.Status(status)
	c.Writer.WriteHeaderNow()
	c.Abort()
}

// AbortWithJSON calls `Abort()` and then `JSON` internally.
// This method stops the chain, writes the status code and return a JSON body.
// It also sets the Content-Type as "application/json".
func (c *Context) AbortWithJSON(status int, obj any) {
	c.Abort()
	c.JSON(status, obj)
}

/************************************/
/************ INPUT DATA ************/
/************************************/

// GetParam returns the value of the URL param.
// It is a shortcut for c.Params.ByName(key)
//
//	router.GET("/user/:id", func(c *gin.Context) {
//	    // a GET request to /user/john
//	    id := c.Param("id") // id == "john"
//	    // a GET request to /user/john/
//	    id := c.Param("id") // id == "/john/"
//	})
func (c *Context) GetParam(key string) string {
	return c.Params.ByName(key)
}

// GetFormValue returns
func (c *Context) GetFormValue(key string) (string, error) {
	err := c.Request.ParseForm()
	if err != nil {
		return "", err
	}

	return c.Request.FormValue(key), nil
}

// GetQueryValue returns
func (c *Context) GetQueryValue(key string) (string, bool) {
	if c.queryCache == nil {
		if c.Request != nil && c.Request.URL != nil {
			c.queryCache = c.Request.URL.Query()
		} else {
			c.queryCache = url.Values{}
		}
	}

	values, ok := c.queryCache[key]
	if !ok {
		return "", false
	}

	return values[0], false
}

// GetHeader returns value from request headers.
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// GetRawData returns stream data.
func (c *Context) GetRawData() ([]byte, error) {
	if c.Request.Body == nil {
		return nil, errors.New("cannot read nil body")
	}
	return io.ReadAll(c.Request.Body)
}

// Bind Method
// Methods of type Bind allow data to be bound to a structure.
// Structured data such as JSON can simply be provided with a Bind method.
// While Query Header data may need to be handled differently in different scenarios: it may be taken singly,
// or it may be bound to a structure.

// MustBind binds the passed struct pointer using the specified binding engine.
// It will abort the request with HTTP 400 if any error occurs.
// See the binding package.
func (c *Context) MustBind(obj any, bind binding.Binding) error {
	if err := bind.Bind(c.Request, obj); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return err
	}

	return nil
}

// BindJSON binds json data to a structure.
func (c *Context) BindJSON(obj any) error {
	return c.MustBind(obj, binding.JSON)
}

// BindPlain binds
func (c *Context) BindPlain(obj any) error {
	return c.MustBind(obj, binding.PLAIN)
}

func (c *Context) BindYAML(obj any) error {
	return c.MustBind(obj, binding.YAML)
}

func (c *Context) BindXML(obj any) error {
	return c.MustBind(obj, binding.XML)
}

func (c *Context) BindQuery(obj any) error {
	return c.MustBind(obj, binding.Query)
}

func (c *Context) BindHeader(obj any) error {
	return c.MustBind(obj, binding.Header)
}

func (c *Context) BindUri(obj any) error {
	m := make(map[string][]string, len(c.Params))
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	if err := binding.Uri.BindingUri(m, obj); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return err
	}

	return nil
}

/*************************/
/***** RESPONSE INFO ******/
/*************************/

// bodyAllowedForStatus is a copy of http.bodyAllowedForStatus non-exported function.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}

// Status sets the HTTP response code.
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// Header sets key in http header
func (c *Context) Header(key, value string) {
	if value == "" {
		c.Writer.Header().Del(key)
		return
	}
	c.Writer.Header().Set(key, value)
}

// Render writes the response headers and calls render.Render to render data.
func (c *Context) Render(code int, r render.Render) {
	c.Status(code)

	if !bodyAllowedForStatus(code) {
		r.WriteContentType(c.Writer)
		c.Writer.WriteHeaderNow()
		return
	}

	if err := r.Render(c.Writer); err != nil {
		c.Abort()
	}
}

// String render a string to client
func (c *Context) String(code int, format string, values ...any) {
	c.Render(code, render.String{Format: format, Data: values})
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (c *Context) JSON(code int, obj any) {
	c.Render(code, render.JSON{Data: obj})
}

// Data render a byte stream to client
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, render.Data{ContentType: contentType, Data: data})
}

// HTML render a html to client
func (c *Context) HTML(code int, name string, obj any) {
	instance := c.server.HTMLRender.Instance(name, obj)
	c.Render(code, instance)
}

// ProtoBuf serializes the given struct as ProtoBuf into the response body.\
func (c *Context) ProtoBuf(code int, data any) {
	c.Render(code, render.ProtoBuf{Data: data})
}

// YAML serializes the given struct as YAML into the response body.
func (c *Context) YAML(code int, data []byte) {
	c.Render(code, render.YAML{Data: data})
}

// Redirect returns an HTTP redirect to the specific location.
func (c *Context) Redirect(code int, location string) {
	c.Render(-1, render.Redirect{
		Code:      code,
		Request:   c.Request,
		Localtion: location,
	})
}

// hasRequestContext returns whether c.Request has Context and fallback.
func (c *Context) hasRequestContext() bool {
	hasFallback := c.server != nil && c.server.ContextWithFallback
	hasRequestContext := c.Request != nil && c.Request.Context() != nil
	return hasFallback && hasRequestContext
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	if !c.hasRequestContext() {
		return
	}
	return c.Request.Context().Deadline()
}

func (c *Context) Done() <-chan struct{} {
	if !c.hasRequestContext() {
		return nil
	}
	return c.Request.Context().Done()
}

func (c *Context) Err() error {
	if !c.hasRequestContext() {
		return nil
	}
	return c.Request.Context().Err()
}

func (c *Context) Value(key any) any {
	if key == ContextRequestKey {
		return c.Request
	}
	if key == ContextKey {
		return c
	}
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}
	if !c.hasRequestContext() {
		return nil
	}
	return c.Request.Context().Value(key)
}

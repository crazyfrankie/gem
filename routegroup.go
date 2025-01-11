package gem

import (
	"net/http"
	"regexp"
)

var (
	regEnLetter = regexp.MustCompile("^[A-Z]+$")

	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

type Router interface {
	Routes
	Group(string, ...HandlerFunc) *RouterGroup
}

type Routes interface {
	Use(...HandlerFunc) Routes

	AddRoute(string, string, ...HandlerFunc) Routes

	Any(string, ...HandlerFunc) Routes
	GET(string, ...HandlerFunc) Routes
	POST(string, ...HandlerFunc) Routes
	DELETE(string, ...HandlerFunc) Routes
	PATCH(string, ...HandlerFunc) Routes
	PUT(string, ...HandlerFunc) Routes
	OPTIONS(string, ...HandlerFunc) Routes
	HEAD(string, ...HandlerFunc) Routes
}

type RouterGroup struct {
	Handlers HandlersChain
	server   *Server
	basePath string
	root     bool
}

// Use adds middleware to the group
// Unlike server's Use, it works on routes that have the same prefix and are not root paths like "/user/login" and "/user/signup"
func (group *RouterGroup) Use(middleware ...HandlerFunc) Routes {
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnObj()
}

// Group creates a new router group. You should add all the routes that have common middlewares or the same path prefix.
// For example, all the routes that use a common middleware for authorization could be grouped.
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		server:   group.server,
	}
}

func (group *RouterGroup) handleRoute(httpMethod, relativePath string, handlers HandlersChain) Routes {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.server.addRoute(httpMethod, absolutePath, handlers)

	return group.returnObj()
}

// AddRoute registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
// See the example code in GitHub.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (group *RouterGroup) AddRoute(httpMethod, relativePath string, handlers ...HandlerFunc) Routes {
	if matched := regEnLetter.MatchString(httpMethod); !matched {
		panic("http method" + httpMethod + "is not valid")
	}
	return group.handleRoute(httpMethod, relativePath, handlers)
}

func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) Routes {
	return group.handleRoute(http.MethodGet, relativePath, handlers)
}

func (group *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) Routes {
	return group.handleRoute(http.MethodPost, relativePath, handlers)
}

func (group *RouterGroup) PUT(relativePath string, handlers ...HandlerFunc) Routes {
	return group.handleRoute(http.MethodPut, relativePath, handlers)
}

func (group *RouterGroup) DELETE(relativePath string, handlers ...HandlerFunc) Routes {
	return group.handleRoute(http.MethodDelete, relativePath, handlers)
}

func (group *RouterGroup) PATCH(relativePath string, handlers ...HandlerFunc) Routes {
	return group.handleRoute(http.MethodPatch, relativePath, handlers)
}

func (group *RouterGroup) HEAD(relativePath string, handlers ...HandlerFunc) Routes {
	return group.handleRoute(http.MethodHead, relativePath, handlers)
}

func (group *RouterGroup) OPTIONS(relativePath string, handlers ...HandlerFunc) Routes {
	return group.handleRoute(http.MethodOptions, relativePath, handlers)
}

func (group *RouterGroup) Any(relativePath string, handlers ...HandlerFunc) Routes {
	for _, method := range anyMethods {
		group.handleRoute(method, relativePath, handlers)
	}

	return group.returnObj()
}

// BasePath returns the base path of router group.
// For example, if v := router.Group("/api/v1/user/"), v.BasePath() is "/api/v1/user/".
func (group *RouterGroup) BasePath() string {
	return group.basePath
}

func (group *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(group.Handlers) + len(handlers)
	assert(finalSize < int(abortIndex), "too many handlers")
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}

func (group *RouterGroup) returnObj() Routes {
	if group.root {
		return group.server
	}
	return group
}

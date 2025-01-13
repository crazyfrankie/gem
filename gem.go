package gem

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/crazyfrankie/gem/render"
)

const escapedColon = "\\:"
const colon = ":"
const backslash = "\\"

type HandlerFunc func(*Context)

type HandlersChain []HandlerFunc

type Server struct {
	// route
	RouterGroup
	trees methodTrees

	// Context pool
	ctxPool sync.Pool

	// Render
	delims     render.Delims
	HTMLRender render.HTMLRender

	// UnescapePathValues if true, the path value will be unescaped.
	// If UseRawPath is false (by default), the UnescapePathValues effectively is true,
	// as url.Path gonna be used, which is already unescaped.
	UnescapePathValues bool

	// UseRawPath if enabled, the url.RawPath will be used to find parameters.
	UseRawPath bool

	// UseH2C enable h2c support.
	UseH2C bool

	// ContextWithFallback enable fallback Context.Deadline(), Context.Done(), Context.Err() and Context.Value() when Context.Request.Context() is not nil.
	ContextWithFallback bool
	maxParams           uint16
	maxSections         uint16
}

func New() *Server {
	server := &Server{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		trees:              make(methodTrees, 0, 9),
		delims:             render.Delims{Left: "{{", Right: "}}"},
		UseRawPath:         false,
		UnescapePathValues: true,
	}
	server.RouterGroup.server = server
	server.ctxPool.New = func() any {
		return server.allocateContext(server.maxParams)
	}

	return server
}

func Default() *Server {
	server := New()
	server.Use(Recovery())
	return server
}

func (server *Server) allocateContext(maxParams uint16) *Context {
	v := make(Params, 0, maxParams)
	skippedNodes := make([]skippedNode, 0, server.maxSections)
	return &Context{server: server, params: &v, skippedNodes: &skippedNodes}
}

// Use Register the middleware in the root path like "/"
func (server *Server) Use(middleware ...HandlerFunc) Routes {
	server.RouterGroup.Use(middleware...)

	return server
}

// Delims sets template left and right delims and returns an Engine instance.
func (server *Server) Delims(left, right string) *Server {
	server.delims = render.Delims{Left: left, Right: right}
	return server
}

func (server *Server) Handler() http.Handler {
	if !server.UseH2C {
		return server
	}

	h2s := &http2.Server{}
	return h2c.NewHandler(server, h2s)
}

// updateRouteTree do update to the route tree recursively
func updateRouteTree(n *node) {
	n.path = strings.ReplaceAll(n.path, escapedColon, colon)
	n.fullPath = strings.ReplaceAll(n.fullPath, escapedColon, colon)
	n.indices = strings.ReplaceAll(n.indices, backslash, colon)
	if n.children == nil {
		return
	}
	for _, child := range n.children {
		updateRouteTree(child)
	}
}

// updateRouteTrees do update to the route trees
func (server *Server) updateRouteTrees() {
	for _, tree := range server.trees {
		updateRouteTree(tree.root)
	}
}

func (server *Server) Start(addr string) error {
	server.updateRouteTrees()
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
		return err
	}

	httpServer := &http.Server{
		Addr:    addr,
		Handler: server,
	}
	go func() {
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()
	log.Printf("Server is running address: http://localhost%s\n", addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutdown signal received")

	// 关闭服务器前的操作：比如停止定时任务等
	// 在这里可以进行一些清理工作，或者执行回调函数
	// log.Println("Performing shutdown tasks...")
	// 如果有回调函数可以在这里执行
	// shutdownCallback() // 假设有回调函数，进行一些清理任务

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown failed: %v", err)
		return err
	}

	// 服务器完全关闭后的操作：比如服务注销等
	// 在这里可以执行一些服务注销的操作
	// log.Println("Deregistering server from the registry...")

	log.Println("Server exited gracefully")
	return nil
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := server.ctxPool.Get().(*Context)
	ctx.writemem.reset(writer)
	ctx.Request = request
	// Ensure that the context of the current request is clean
	ctx.reset()

	// Find Route
	// Execute business logic
	server.handleHTTPRequest(ctx)

	server.ctxPool.Put(ctx)
}

func (server *Server) addRoute(method string, path string, handlers HandlersChain) {
	// Why panic?
	// Most users prefer to ignore errors when registering routes.
	// Or maybe it's a mechanism for early calibration
	assert(path[0] == '/', "path must start with '/'")
	assert(method != "", "HTTP method can not be empty")
	assert(len(handlers) > 0, "handlers can not be empty")

	root := server.trees.get(method)
	if root == nil {
		root = new(node)
		root.fullPath = "/"
		server.trees = append(server.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)

	if countParams := countParams(path); countParams > server.maxParams {
		server.maxParams = countParams
	}

	if countSections := countSections(path); countSections > server.maxSections {
		server.maxSections = countSections
	}
}

func (server *Server) handleHTTPRequest(ctx *Context) {
	httpMethod := ctx.Request.Method
	path := ctx.Request.URL.Path
	unescape := false

	if server.UseRawPath && len(ctx.Request.URL.RawPath) > 0 {
		path = ctx.Request.URL.RawPath
		unescape = server.UnescapePathValues
	}

	// Find Route
	tree := server.trees
	for i, tl := 0, len(tree); i < tl; i++ {
		if tree[i].method != httpMethod {
			continue
		}
		root := tree[i].root
		// Find route in tree
		value := root.getValue(path, ctx.params, ctx.skippedNodes, unescape)
		if value.params != nil {
			ctx.Params = *value.params
		}
		if value.handlers != nil {
			ctx.handlers = value.handlers
			ctx.fullPath = value.fullPath
			ctx.Next()
			ctx.writemem.WriteHeaderNow()
			return
		}
	}
}

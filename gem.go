package gem

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/crazyfrankie/gem/render"
)

type HandlerFunc func(*Context)

type HandlersChain []HandlerFunc

type Server struct {
	// route
	RouterGroup
	trees methodTrees

	// Context pool
	ctxPool sync.Pool

	maxParams   uint16
	maxSections uint16

	// ContextWithFallback enable fallback Context.Deadline(), Context.Done(), Context.Err() and Context.Value() when Context.Request.Context() is not nil.
	ContextWithFallback bool

	// Render
	delims     render.Delims
	HTMLRender render.HTMLRender
}

func New() *Server {
	server := &Server{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		trees:  make(methodTrees, 0, 9),
		delims: render.Delims{Left: "{{", Right: "}}"},
	}
	server.RouterGroup.server = server
	server.ctxPool.New = func() any {
		return server.allocateContext(server.maxParams)
	}

	return server
}

func Default() *Server {
	server := New()
	server.Use()
	return server
}

func (server *Server) allocateContext(maxParams uint16) *Context {
	v := make(Params, 0, maxParams)
	return &Context{server: server, params: &v}
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

func (server *Server) Start(addr string) error {
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

	}
}

func (server *Server) handleHTTPRequest(ctx *Context) {

}

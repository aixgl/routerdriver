package routerdriver

import (
	"net/http"
	"os"
	"path/filepath"
)

type IParams interface {
	ByName(string) (string, bool)
	By(uint) (string, bool)
}

type Router struct {
	UrlMap           *UrlMap
	NotFound         http.Handler
	MethodNotAllowed http.Handler
	PanicHandler     func(http.ResponseWriter, *http.Request, interface{})
	HandleOPTIONS    bool
}

func New() *Router {
	return &Router{
		HandleOPTIONS: true,
		UrlMap:        &UrlMap{Store: make(map[string][]*node), Num: 0},
	}
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *Router) GET(path string, handle interface{}) {
	r.Handle("GET", path, handle)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, handle interface{}) {
	r.Handle("HEAD", path, handle)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *Router) OPTIONS(path string, handle interface{}) {
	r.Handle("OPTIONS", path, handle)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *Router) POST(path string, handle interface{}) {
	r.Handle("POST", path, handle)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *Router) PUT(path string, handle interface{}) {
	r.Handle("PUT", path, handle)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *Router) PATCH(path string, handle interface{}) {
	r.Handle("PATCH", path, handle)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, handle interface{}) {
	r.Handle("DELETE", path, handle)
}

func (r *Router) Handle(method string, path string, handle interface{}) {
	if path[0] != '/' {
		panic("path must begin with '/' in path.your defined path is: " + path)
	}

	container := r.UrlMap
	container.addRouter(path, handle, method)
}

func (r *Router) Handler(method, path string, handler http.Handler) {
	r.Handle(method, path,
		func(w http.ResponseWriter, req *http.Request) {
			handler.ServeHTTP(w, req)
		},
	)
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (r *Router) ServeFiles(path string, root string) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	if len(root) > 1 && root[0] == '.' {
		pwd, _ := os.Getwd()
		root = filepath.Join(pwd, root)
	}
	fileServer := http.FileServer(http.Dir(root))

	r.Handle(METHOD_STATIC, path, func(w http.ResponseWriter, req *http.Request, ps IParams) {
		req.URL.Path, _ = ps.ByName("filepath")
		fileServer.ServeHTTP(w, req)
	})
}

func (r *Router) Lookup(method, path string) *RouterRet {
	ret := r.UrlMap.getValue(path)

	if ret.Type == "" || ret.Type == method {
		return ret
	}

	return nil
}

//painc recover
func (r *Router) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(w, req, rcv)
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.PanicHandler != nil {
		defer r.recv(w, req)
	}

	path := req.URL.Path
	if pnode := r.UrlMap.getValue(path, req.Method); pnode != nil {
		handle, ok := pnode.Handle.(func(http.ResponseWriter, *http.Request, IParams))
		if ok {
			handle(w, req, pnode)
			return
		} else if req.Method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if req.Method != "GET" {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}

			//http.Redirect(w, req, req.URL.String(), code)
			Print(code)
		}
	}

	//option
	if req.Method == "OPTIONS" {
		// Handle OPTIONS requests
		if r.HandleOPTIONS {
			w.Header().Set("Allow", "GET,POST,PUT,DELETE")
			return
		}
	}
	// Handle 404
	if r.NotFound != nil {
		r.NotFound.ServeHTTP(w, req)
	} else {
		http.NotFound(w, req)
	}
}

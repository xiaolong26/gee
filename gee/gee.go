package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

type Handlefunc func(c *Context)

type Engine struct {
	//engine拥有routergroup所有方法,&RouterGroup{engine: engine}成为所有routergroup节点的父节点
	*RouterGroup
	router *router
	groups []*RouterGroup
	htmlTemplates *template.Template // for html render
	funcMap       template.FuncMap   // for html render
}

type RouterGroup struct {
	//共享一个engine
	engine 		*Engine
	prefix 		string
	parent      *RouterGroup
	middlewares []Handlefunc
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (group *RouterGroup)Group(prefix string) *RouterGroup{
	//gee.group,使&RouterGroup{engine: engine}成为所有routergroup节点的父节点
	newGroup := &RouterGroup{
		engine:      group.engine,
		prefix:      group.prefix + prefix,
		parent:      group,
		middlewares: nil,
	}
	group.engine.groups = append(group.engine.groups,newGroup)
	return newGroup
}

func (engine *Engine)addRoute(method string,pattern string,handle Handlefunc){
	engine.router.addRoute(method,pattern,handle)
}

func (engine *Engine) GET(pattern string, handler Handlefunc) {
	engine.router.addRoute("GET", pattern, handler)
}

func (engine *Engine) POST(pattern string, handler Handlefunc) {
	engine.router.addRoute("POST", pattern, handler)
}

func (group *RouterGroup) addRoute(method string, comp string, handler Handlefunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler Handlefunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler Handlefunc) {
	group.addRoute("POST", pattern, handler)
}

func (group *RouterGroup) Use(middlewares ...Handlefunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) Handlefunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// serve static files
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	group.GET(urlPattern, handler)
}

func (engine *Engine)Run(add string)(err error)  {
	return http.ListenAndServe(add,engine)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter,r *http.Request)  {
	//封装成context,函数实现由用户写
	var middlewares []Handlefunc
	for _, group := range engine.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, r)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

func Logger() Handlefunc {
	return func(c *Context) {
		// Start timer
		t := time.Now()
		// Process request
		c.Next()
		// Calculate resolution time
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
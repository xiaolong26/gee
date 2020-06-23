package gee

import (
	"log"
	"net/http"
	"strings"
)

type Handlefunc func(c *Context)

type Engine struct {
	//engine拥有routergroup所有方法,&RouterGroup{engine: engine}成为所有routergroup节点的父节点
	*RouterGroup
	router *router
	groups []*RouterGroup
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

func (engine *Engine)Run(add string)(err error)  {
	return http.ListenAndServe(add,engine)
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
	engine.router.handle(c)
}
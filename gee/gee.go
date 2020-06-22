package gee

import (
	"net/http"
)

type Handlefunc func(c *Context)

type Engine struct {
	router *router
}

func New() *Engine {
	return &Engine{router: newRouter()}
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

func (engine *Engine)Run(add string)(err error)  {
	return http.ListenAndServe(add,engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter,r *http.Request)  {
	//封装成context,函数实现由用户写
	c := newContext(w,r)
	engine.router.handle(c)
}
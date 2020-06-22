package gee

import (
	"net/http"
	"strings"
)

type router struct {
	handlers map[string]Handlefunc
	roots	 map[string]*node
}

func newRouter()*router  {
	return &router{make(map[string]Handlefunc),make(map[string]*node)}
}

func (r *router)addRoute(method string,pattern string,handler Handlefunc)  {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_,ok := r.roots[method]
	if !ok{
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern,parts,0)
	r.handlers[key] = handler
}

func (r *router)handle(c *Context){
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		r.handlers[key](c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}

func parsePattern(pattern string)[]string  {
	vs := strings.Split(pattern,"/")
	parts := make([]string,0)
	for _,item := range vs{
		if item != ""{
			parts = append(parts,item)
			if item[0] == '*'{
				break
			}
		}
	}
	return parts
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}
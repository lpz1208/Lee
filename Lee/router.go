package Lee

import (
	"net/http"
	"strings"
	"sync"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
	
	// 性能优化：预分配切片容量
	paramsPool sync.Pool
}

//type router struct {
//	handlers map[string]HandlerFunc
//}

//func newRouter() *router {
//	return &router{handlers: make(map[string]HandlerFunc)}
//}
func newRouter() *router {
	r := &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
	
	// 初始化参数池
	r.paramsPool.New = func() interface{} {
		return make(map[string]string, 4) // 预分配4个参数的容量
	}
	
	return r
}

// Only one * is allowed
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

//func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
//	log.Printf("Route %4s - %s", method, pattern)
//	key := method + "-" + pattern
//	r.handlers[key] = handler
//}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)

	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
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

//func (r *router) handle(c *Context) {
//	key := c.Method + "-" + c.Path
//	if handler, ok := r.handlers[key]; ok {
//		handler(c)
//	} else {
//		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
//	}
//}
//func (r *router) handle(c *Context) {
//	n, params := r.getRoute(c.Method, c.Path)
//	if n != nil {
//		c.Params = params
//		key := c.Method + "-" + n.pattern
//		r.handlers[key](c)
//	} else {
//		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
//	}
//}
func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)

	if n != nil {
		key := c.Method + "-" + n.pattern
		c.Params = params
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	c.Next()
}

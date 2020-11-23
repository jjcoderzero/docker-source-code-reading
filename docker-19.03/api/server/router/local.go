package router // import "docker-19.03/api/server/router"

import (
	"docker-19.03/api/server/httputils"
)

// RouteWrapper包装了具有额外功能的路由。它是在创建新路由时传入的。
type RouteWrapper func(r Route) Route

// localRoute定义了一个单独的API路由来连接docker守护进程。它实现了Route.
type localRoute struct {
	method  string
	path    string
	handler httputils.APIFunc
}

// Handler返回APIFunc以让服务器将其封装在中间件中.
func (l localRoute) Handler() httputils.APIFunc {
	return l.handler
}

// Method返回路由响应的http方法.
func (l localRoute) Method() string {
	return l.method
}

// Path返回路由响应的子路径.
func (l localRoute) Path() string {
	return l.path
}

// NewRoute为路由器初始化一个新的本地路由。
func NewRoute(method, path string, handler httputils.APIFunc, opts ...RouteWrapper) Route {
	var r Route = localRoute{method, path, handler}
	for _, o := range opts {
		r = o(r)
	}
	return r
}

// NewGetRoute使用http方法GET初始化一个新路由。
func NewGetRoute(path string, handler httputils.APIFunc, opts ...RouteWrapper) Route {
	return NewRoute("GET", path, handler, opts...)
}

// NewPostRoute使用http方法POST初始化一个新路由。
func NewPostRoute(path string, handler httputils.APIFunc, opts ...RouteWrapper) Route {
	return NewRoute("POST", path, handler, opts...)
}

// NewPutRoute使用http方法PUT初始化一个新路由。
func NewPutRoute(path string, handler httputils.APIFunc, opts ...RouteWrapper) Route {
	return NewRoute("PUT", path, handler, opts...)
}

// NewDeleteRoute使用http方法DELETE初始化一个新路由。
func NewDeleteRoute(path string, handler httputils.APIFunc, opts ...RouteWrapper) Route {
	return NewRoute("DELETE", path, handler, opts...)
}

// NewOptionsRoute使用http方法OPTIONS初始化一个新路由。
func NewOptionsRoute(path string, handler httputils.APIFunc, opts ...RouteWrapper) Route {
	return NewRoute("OPTIONS", path, handler, opts...)
}

// NewHeadRoute使用http方法HEAD初始化一个新路由。
func NewHeadRoute(path string, handler httputils.APIFunc, opts ...RouteWrapper) Route {
	return NewRoute("HEAD", path, handler, opts...)
}

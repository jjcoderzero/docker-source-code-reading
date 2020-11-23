package router // import "docker-19.03/api/server/router"

import "docker-19.03/api/server/httputils"

// Router定义了一个接口来指定要添加到docker服务器的一组路由。
type Router interface {
	Routes() []Route // Routes返回要添加到docker服务器的路由列表。
}

// Route在docker服务器中定义一个独立的API路由。
type Route interface {
	Handler() httputils.APIFunc // Handler返回原始函数来创建http处理程序。
	Method() string // Method返回路由响应的http方法。
	Path() string // Path返回路由响应的子路径。
}

package router

import "go.uber.org/fx"

// Module exports dependency to container
var Module = fx.Options(
	fx.Provide(NewWebRouter),
	fx.Provide(NewTestRouter),
	fx.Provide(NewRouters),
)

type Routers []IRouter

// Route interface
type IRouter interface {
	Setup()
}

// NewRoutes sets up routes
func NewRouters(webRouter *WebRouter, testRouter *TestRouter) Routers {
	return Routers{
		webRouter,
		testRouter,
	}
}

// Setup all the route
func (r Routers) Setup() {
	for _, route := range r {
		route.Setup()
	}
}

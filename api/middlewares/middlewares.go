package middlewares

import "go.uber.org/fx"

// Module Middleware exported
var Module = fx.Options(
	fx.Provide(NewApiAuthorityMiddleWare),
	fx.Provide(NewWebsocketAuthorityMiddleWare),
	fx.Provide(NewLanguageMiddleware),
	fx.Provide(NewTraceIdMiddleWare),
	fx.Provide(NewRateLimitMiddleWare),
	fx.Provide(NewRecoverMiddleWare),
	fx.Provide(NewMiddlewares),
)

// IMiddleware middleware interface
type IMiddleware interface {
	Setup()
}

// Middlewares contains multiple middleware
type Middlewares []IMiddleware

// NewMiddlewares creates new middlewares
// Register the middleware that should be applied directly (globally)
func NewMiddlewares(
	languageMiddleware *LanguageMiddleware,
	traceIDMiddleware *TraceIdMiddleWare,
	recoverMiddleWare *RecoverMiddleWare,
) Middlewares {
	return Middlewares{
		recoverMiddleWare,
		traceIDMiddleware,
		languageMiddleware,
	}
}

// Setup sets up middlewares
func (m Middlewares) Setup() {
	for _, middleware := range m {
		middleware.Setup()
	}
}

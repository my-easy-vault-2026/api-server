package infra

import "go.uber.org/fx"

// Module exports dependency
var Module = fx.Options(
	fx.Provide(fx.Annotate(NewAPIRouter, fx.ResultTags(`name:"api"`))),
	fx.Provide(fx.Annotate(NewWebsocketRouter, fx.ResultTags(`name:"websocket"`))),
	fx.Provide(NewDatabase),
	fx.Provide(NewRedis),
	fx.Provide(NewMQ),
	fx.Provide(NewWs),
	fx.Provide(NewWorkerPools),
	fx.Provide(NewLockers),
)

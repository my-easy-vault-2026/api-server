package handlers

import (
	"api-server/api/handlers/test"
	"api-server/api/handlers/web"

	"go.uber.org/fx"
)

// Module exported for initializing application
var Module = fx.Options(
	fx.Provide(web.NewAuthHandler),
	fx.Provide(web.NewCommonHandler),
	fx.Provide(web.NewExchangeHandler),
	fx.Provide(web.NewOrderHandler),
	fx.Provide(web.NewQuoteHandler),
	fx.Provide(web.NewSystemHandler),
	fx.Provide(web.NewTransferHandler),
	fx.Provide(web.NewUserHandler),
	fx.Provide(web.NewWalletHandler),
	fx.Provide(web.NewWebsocketHandler),

	fx.Provide(test.NewTestHandler),
	fx.Provide(test.NewAccountHandler),
	fx.Provide(test.NewCardHandler),
)

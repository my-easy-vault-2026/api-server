package handlers

import (
	"api-server/handlers/test"
	"api-server/handlers/web"

	"go.uber.org/fx"
)

// Module exported for initializing application
var Module = fx.Options(
	fx.Provide(web.NewAccountHandler),
	fx.Provide(web.NewAuthHandler),
	fx.Provide(web.NewCardHandler),
	fx.Provide(web.NewCommonHandler),
	fx.Provide(web.NewExchangeHandler),
	fx.Provide(web.NewFinancialHandler),
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

package services

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewAccountService),
	fx.Provide(NewAuthService),
	fx.Provide(NewCardService),
	fx.Provide(NewCoinsdoService),
	fx.Provide(NewFinancialService),
	fx.Provide(NewOrderService),
	fx.Provide(NewQuoteService),
	fx.Provide(NewSystemService),
	fx.Provide(NewTransferService),
	fx.Provide(NewUserService),
	fx.Provide(NewWalletService),
	fx.Provide(NewWebsocketService),
	fx.Provide(NewTestService),
)
